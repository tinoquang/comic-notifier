package crawler

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type helper interface {
	detectSpoiler(chapURL string, attr1, attr2 string) error
	getPageSource(pageURL string) (doc *goquery.Document, err error)
}
type comicCrawler struct {
	crawlerMap  map[string]func(ctx context.Context, comic *db.Comic, helper helper) (err error)
	crawlHelper helper
}

func newComicCrawler(crawlHelper helper) *comicCrawler {

	crawlerMap := make(map[string]func(ctx context.Context, comic *db.Comic, helper helper) (err error))
	crawlerMap["beeng.net"] = crawlBeeng
	crawlerMap["blogtruyen.vn"] = crawlBlogtruyen
	crawlerMap["truyentranh.net"] = crawlTruyentranhnet
	crawlerMap["truyentranhtuan.com"] = crawlTruyentranhtuan

	return &comicCrawler{
		crawlerMap:  crawlerMap,
		crawlHelper: crawlHelper,
	}
}

func (c *comicCrawler) crawl(ctx context.Context, comic *db.Comic) (err error) {

	if _, ok := c.crawlerMap[comic.Page]; !ok {
		return util.ErrPageNotSupported
	}

	return c.crawlerMap[comic.Page](ctx, comic, c.crawlHelper)
}

func crawlBeeng(ctx context.Context, comic *db.Comic, helper helper) (err error) {

	var chapURL string

	doc, err := helper.getPageSource(comic.Url)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find(".detail").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")
	comic.CloudImgUrl = fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, comic.Page, comic.Name)

	// Find latest chap
	firstItem := doc.Find(".listChapters").Find(".list").Find("li:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = strings.TrimSpace(firstItem.Find(".titleComic").Text())
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapUrl {
		return util.ErrComicUpToDate
	}

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(chapURL, ".comicDetail2#lightgallery2", "img")
		if err != nil {
			return
		}
	}

	comic.ChapUrl = chapURL
	return
}

/* ------------------------------------------------------------------------------------------------ */

// blogtruyen crawler
func crawlBlogtruyen(ctx context.Context, comic *db.Comic, helper helper) (err error) {

	var chapURL string

	doc, err := helper.getPageSource(comic.Url)
	if err != nil {
		return util.ErrInvalidURL
	}

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.ImgUrl, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".list-wrap#list-chapters").Find("p:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find(".title").Find("a[href]").Text()
	chapURL, _ = firstItem.Find(".title").Find("a[href]").Attr("href")

	chapURL = "https://blogtruyen.vn" + chapURL
	if comic.ChapUrl == chapURL {
		return util.ErrComicUpToDate
	}

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(chapURL, "#content", "img[src]")
		if err != nil {
			return
		}
	}

	comic.ChapUrl = chapURL
	return
}

/* ------------------------------------------------------------------------------------------------------------------- */

// mangaK crawler
func crawlMangaK(ctx context.Context, comic *db.Comic, helper helper) (err error) {

	var chapURL string

	doc, err := helper.getPageSource(comic.Url)
	if err != nil {
		return util.ErrInvalidURL
	}
	logging.Info(doc.Html())

	comic.Name = doc.Find(".entry-title").Text()
	comic.ImgUrl, _ = doc.Find(".info_image").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".row:nth-child(1)")
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("span:nth-child(1)").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapUrl {
		return util.ErrComicUpToDate
	}

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(chapURL, ".vung_doc", "img")
		if err != nil {
			return
		}
	}

	comic.ChapUrl = chapURL
	return

}

/* ---------------------------------------------------------------------------------------------------- */

func crawlTruyentranhtuan(ctx context.Context, comic *db.Comic, helper helper) (err error) {

	var chapURL string

	doc, err := helper.getPageSource(comic.Url)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find("#infor-box").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".manga-cover").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find("#manga-chapter").Find(".chapter-name").First()
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapUrl {
		return util.ErrComicUpToDate
	}

	// Page is load by JS, can't get by just using HTTP.Get --> resolve later
	// if comic.ChapUrl != "" {
	// 	err = helper.detectSpoiler(chapURL, ".vung_doc", "img")
	// 	if err != nil {
	// 		return
	// 	}
	// }

	comic.ChapUrl = chapURL
	return
}

func crawlTruyentranhnet(ctx context.Context, comic *db.Comic, helper helper) (err error) {

	var chapURL string

	url := comic.Url + "?order=desc"
	doc, err := helper.getPageSource(url)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find(".detail-manga-title").Find("h1").Text()
	comic.ImgUrl, _ = doc.Find(".detail-img").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".chapter-select").First()
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapUrl {
		return util.ErrComicUpToDate
	}

	if comic.ChapUrl != "" {
		err = helper.detectSpoiler(chapURL, ".manga-reading-box", "img")
		if err != nil {
			return
		}
	}

	comic.ChapUrl = chapURL
	return
}
