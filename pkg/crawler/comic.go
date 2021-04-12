package crawler

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/PuerkitoBio/goquery"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type helper interface {
	detectSpoiler(name, chapURL, chapterName, attr1, attr2 string) error
	getPageSource(comicURL string) (doc *goquery.Document, err error)
}
type comicCrawler struct {
	crawlerMap  map[string]func(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error)
	crawlHelper helper
}

func newComicCrawler(crawlHelper helper) *comicCrawler {

	crawlerMap := make(map[string]func(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error))
	crawlerMap["beeng.net"] = crawlBeeng
	crawlerMap["blogtruyen.vn"] = crawlBlogtruyen
	crawlerMap["truyentranhtuan.com"] = crawlTruyentranhtuan
	crawlerMap["truyenqq.com"] = crawlTruyenqq
	crawlerMap["hocvientruyentranh.net"] = crawlHocvientruyentranh

	return &comicCrawler{
		crawlerMap:  crawlerMap,
		crawlHelper: crawlHelper,
	}
}

// GetComicInfo return link of latest chapter of a page
func (c *comicCrawler) GetComicInfo(ctx context.Context, comicURL string, checkSpoiler bool) (comic db.Comic, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}

		return
	}()

	parsedURL, err := url.Parse(comicURL)
	if err != nil /*|| parsedURL.Host == "" */ {
		return db.Comic{}, util.ErrInvalidURL
	}

	if _, ok := c.crawlerMap[parsedURL.Hostname()]; !ok {
		return db.Comic{}, util.ErrPageNotSupported
	}

	// Remove all params in comicURL --> avoid duplicate URL
	parsedURL.RawQuery = ""
	comicURL = parsedURL.String()

	doc, err := c.crawlHelper.getPageSource(comicURL)
	if err != nil {
		if strings.Contains(err.Error(), "Timeout") {
			return db.Comic{}, util.ErrCrawlTimeout
		}
		return db.Comic{}, util.ErrCrawlFailed
	}

	comic = db.Comic{
		Page: parsedURL.Hostname(),
		Url:  comicURL,
	}

	err = c.crawlerMap[parsedURL.Hostname()](ctx, doc, &comic, c.crawlHelper, checkSpoiler)
	if err != nil {
		return
	}

	err = verifyComic(&comic)
	return
}

func crawlBeeng(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

	comic.Name = doc.Find(".detail").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".listChapters").Find(".list").Find("li:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrCrawlFailed
	}

	comic.LatestChap = strings.TrimSpace(firstItem.Find(".titleComic").Text())
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	lastUpdate := strings.Fields(strings.TrimSpace(firstItem.Find(".views").Text()))
	if len(lastUpdate) == 0 {
		return util.ErrCrawlFailed
	}

	comic.LastUpdate, err = time.Parse("02-01-2006", string(lastUpdate[0]))
	if err != nil {
		logging.Danger(err)
		return util.ErrCrawlFailed
	}

	if checkSpoiler {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, comic.LatestChap, ".comicDetail2#lightgallery2", "img")
		if err != nil {
			return
		}
	}

	return
}

/* ------------------------------------------------------------------------------------------------ */

// blogtruyen crawler
func crawlBlogtruyen(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.ImgUrl, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".list-wrap#list-chapters").Find("p:nth-child(1)")
	if firstItem.Nodes == nil {
		logging.Danger(err)
		return util.ErrCrawlFailed
	}

	comic.LatestChap = firstItem.Find(".title").Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find(".title").Find("a[href]").Attr("href")

	comic.ChapUrl = "https://blogtruyen.vn" + comic.ChapUrl

	lastUpdate := strings.Fields(strings.TrimSpace(firstItem.Find(".publishedDate").Text()))
	if len(lastUpdate) == 0 {
		return util.ErrCrawlFailed
	}

	comic.LastUpdate, err = time.Parse("02/01/2006", string(lastUpdate[0]))
	if err != nil {
		logging.Danger(err)
		return util.ErrCrawlFailed
	}

	if checkSpoiler {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, comic.LatestChap, "#content", "img[src]")
		if err != nil {
			return
		}
	}

	return
}

/* ------------------------------------------------------------------------------------------------------------------- */
func crawlTruyentranhtuan(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

	comic.Name = doc.Find("#infor-box").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".manga-cover").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find("#manga-chapter").Find(".chapter-name").First()
	if firstItem.Nodes == nil {
		return util.ErrCrawlFailed
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	lastUpdate := doc.Find("#manga-chapter").Find(".date-name").First().Text()
	if len(lastUpdate) == 0 {
		return util.ErrCrawlFailed
	}

	date := strings.Split(lastUpdate, ".")
	if len(date[0]) == 1 {
		date[0] = "0" + date[0]
	}

	comic.LastUpdate, err = time.Parse("02.01.2006", strings.Join(date, "."))
	if err != nil {
		logging.Danger(err)
		return util.ErrCrawlFailed
	}

	// Page is load by JS, can't get by just using HTTP.Get --> resolve later
	// if comic.ChapUrl != "" {
	// 	err = helper.detectSpoiler(comic.ChapUrl, ".vung_doc", "img")
	// 	if err != nil {
	// 		return
	// 	}
	// }

	return
}

func crawlTruyenqq(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

	comic.Name = doc.Find(".center").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".left").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".works-chapter-list").Find(".works-chapter-item.row").First()
	if firstItem.Nodes == nil {
		return util.ErrCrawlFailed
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	lastUpdate := strings.TrimSpace(firstItem.Find(".text-right").Text())
	if len(lastUpdate) == 0 {
		return util.ErrCrawlFailed
	}

	comic.LastUpdate, err = time.Parse("02/01/2006", string(lastUpdate))
	if err != nil {
		logging.Danger(err)
		return util.ErrCrawlFailed
	}

	if checkSpoiler {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, comic.LatestChap, ".story-see-content", "img")
		if err != nil {
			return
		}
	}

	return
}

func crawlHocvientruyentranh(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

	comic.Name = doc.Find(".__info").Find("h3").Text()
	comic.ImgUrl, _ = doc.Find(".__image").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find("tbody").Find("tr").First()
	if firstItem.Nodes == nil {
		return util.ErrCrawlFailed
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	if checkSpoiler {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, comic.LatestChap, ".manga-container", "img")
		if err != nil {
			return
		}
	}

	return
}

func verifyComic(comic *db.Comic) (err error) {

	err = nil
	switch {
	case comic.Name == "":
		return errors.Errorf("Comic name is missing, url = %s", comic.Url)
	case comic.ChapUrl == "":
		return errors.Errorf("Comic chapURL is missing, url = %s", comic.Url)
	case comic.ImgUrl == "":
		return errors.Errorf("Comic ImgUrl is missing, url = %s", comic.Url)
	case comic.CloudImgUrl == "":
		return errors.Errorf("Comic cloudImgUrl is missing, url = %s", comic.Url)
	case comic.LatestChap == "":
		return errors.Errorf("Comic latestchap is missing, url = %s", comic.Url)
	default:
		err = nil
	}

	if comic.Page != "hocvientruyentranh.net" {
		if comic.LastUpdate.IsZero() {
			return errors.Errorf("Comic date is missing, url = %s", comic.Url)
		}
	}

	return
}

// mangaK crawler: work on this one later
// func crawlMangaK(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper, checkSpoiler bool) (err error) {

// 	comic.Name = doc.Find(".entry-title").Text()
// 	comic.ImgUrl, _ = doc.Find(".info_image").Find("img[src]").Attr("src")
// 	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

// 	// Find latest chap
// 	firstItem := doc.Find(".chapter-list").Find(".row:nth-child(1)")
// 	if firstItem.Nodes == nil {
// 		return util.ErrCrawlFailed
// 	}

// 	comic.LatestChap = firstItem.Find("span:nth-child(1)").Text()
// 	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

// 	if comic.ChapUrl != "" {
// 		err = helper.detectSpoiler(comic.Name, comic.ChapUrl,comic.LatestChap, ".vung_doc", "img")
// 		if err != nil {
// 			return
// 		}
// 	}

// 	err = verifyComic(comic)
// 	return
// }
