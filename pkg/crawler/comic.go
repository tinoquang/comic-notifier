package crawler

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type helper interface {
	detectSpoiler(name, chapURL, attr1, attr2 string) error
	getPageSource(comicURL string) (doc *goquery.Document, err error)
}
type comicCrawler struct {
	crawlerMap  map[string]func(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error)
	crawlHelper helper
}

func newComicCrawler(crawlHelper helper) *comicCrawler {

	crawlerMap := make(map[string]func(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error))
	crawlerMap["beeng.net"] = crawlBeeng
	crawlerMap["blogtruyen.vn"] = crawlBlogtruyen
	crawlerMap["truyentranh.net"] = crawlTruyentranhnet
	crawlerMap["truyentranhtuan.com"] = crawlTruyentranhtuan

	return &comicCrawler{
		crawlerMap:  crawlerMap,
		crawlHelper: crawlHelper,
	}
}

func (c *comicCrawler) crawl(ctx context.Context, comicURL string) (comic db.Comic, err error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return db.Comic{}, util.ErrInvalidURL
	}

	if _, ok := c.crawlerMap[parsedURL.Hostname()]; !ok {
		return db.Comic{}, util.ErrPageNotSupported
	}

	if parsedURL.Hostname() == "truyentranh.net" {
		comicURL += "?order=desc"
	}

	doc, err := c.crawlHelper.getPageSource(comicURL)
	if err != nil {
		if strings.Contains(err.Error(), "Timeout") {
			return db.Comic{}, util.ErrCrawlTimeout
		}
		return db.Comic{}, util.ErrInvalidURL
	}

	comic = db.Comic{
		Page: parsedURL.Hostname(),
		Url:  comicURL,
	}

	err = c.crawlerMap[parsedURL.Hostname()](ctx, doc, &comic, c.crawlHelper)
	return
}

func crawlBeeng(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error) {

	comic.Name = doc.Find(".detail").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".listChapters").Find(".list").Find("li:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = strings.TrimSpace(firstItem.Find(".titleComic").Text())
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, ".comicDetail2#lightgallery2", "img")
		if err != nil {
			return
		}
	}

	err = verifyComic(comic)
	return
}

/* ------------------------------------------------------------------------------------------------ */

// blogtruyen crawler
func crawlBlogtruyen(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error) {

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.ImgUrl, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".list-wrap#list-chapters").Find("p:nth-child(1)")
	if firstItem.Nodes == nil {
		logging.Danger(err)
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find(".title").Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find(".title").Find("a[href]").Attr("href")

	comic.ChapUrl = "https://blogtruyen.vn" + comic.ChapUrl

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, "#content", "img[src]")
		if err != nil {
			return
		}
	}

	err = verifyComic(comic)
	return
}

/* ------------------------------------------------------------------------------------------------------------------- */

// mangaK crawler
func crawlMangaK(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error) {

	comic.Name = doc.Find(".entry-title").Text()
	comic.ImgUrl, _ = doc.Find(".info_image").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".row:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find("span:nth-child(1)").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, ".vung_doc", "img")
		if err != nil {
			return
		}
	}

	err = verifyComic(comic)
	return
}

/* ---------------------------------------------------------------------------------------------------- */

func crawlTruyentranhtuan(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error) {

	comic.Name = doc.Find("#infor-box").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".manga-cover").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find("#manga-chapter").Find(".chapter-name").First()
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	// Page is load by JS, can't get by just using HTTP.Get --> resolve later
	// if comic.ChapUrl != "" {
	// 	err = helper.detectSpoiler(comic.ChapUrl, ".vung_doc", "img")
	// 	if err != nil {
	// 		return
	// 	}
	// }

	err = verifyComic(comic)
	return
}

func crawlTruyentranhnet(ctx context.Context, doc *goquery.Document, comic *db.Comic, helper helper) (err error) {

	comic.Name = doc.Find(".detail-manga-title").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".detail-img").Find("img[src]").Attr("src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".chapter-select").First()
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	comic.ChapUrl, _ = firstItem.Find("a[href]").Attr("href")

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(comic.Name, comic.ChapUrl, ".manga-reading-box", "img")
		if err != nil {
			return
		}
	}

	err = verifyComic(comic)
	return
}

func verifyComic(comic *db.Comic) (err error) {

	err = util.ErrCrawlFailed
	switch {
	case comic.Name == "":
		logging.Danger("Comic name is missing, url", comic.Url)
	case comic.ChapUrl == "":
		logging.Danger("Comic chapURL is missing, url", comic.Url)
	case comic.ImgUrl == "":
		logging.Danger("Comic ImgUrl is missing, url", comic.Url)
	case comic.CloudImgUrl == "":
		logging.Danger("Comic cloudImgUrl is missing, url", comic.Url)
	case comic.LatestChap == "":
		logging.Danger("Comic latestchap is missing, url", comic.Url)
	default:
		err = nil
	}

	return
}
