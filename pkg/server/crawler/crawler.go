package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

var (
	ErrComicUpToDate    = errors.Errorf("Comic is up-to-date, no new chapter")
	ErrPageNotSupported = errors.Errorf("Page is not supported yet")
	ErrInvalidURL       = errors.Errorf("URL is not a comic page")
)

var crawlerMap map[string]func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error)

// New create new crawler
func New() {

	crawlerMap = make(map[string]func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error))
	crawlerMap["beeng.net"] = crawlBeeng
	crawlerMap["truyendep.com"] = crawlMangaK // mangaK is changed to truyendep.com
	crawlerMap["blogtruyen.vn"] = crawlBlogTruyen

}

// GetComicInfo return link of latest chapter of a page
func GetComicInfo(ctx context.Context, comic *model.Comic) (err error) {

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
			logging.Danger()
		}
		return
	}()

	doc, err := getPageSource(comic.URL)
	if err != nil {
		return errors.Wrapf(err, "Can't retrieve page's HTML")
	}

	if _, ok := crawlerMap[comic.Page]; !ok {
		logging.Danger(fmt.Sprintf("%s is not supported", comic.Page))
		return ErrPageNotSupported
	}
	err = crawlerMap[comic.Page](ctx, doc, comic)
	return err
}

func crawlBeeng(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName string

	comic.Name = doc.Find(".detail").Find("h4").Text()
	comic.DateFormat = "02-01-2006 15:04"
	comic.ImageURL, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")

	// Find latest chap
	firstItem := doc.Find(".listChapters").Find(".list").Find("li:nth-child(1)")
	if firstItem.Nodes == nil {
		return ErrInvalidURL
	}

	chapName = strings.TrimSpace(firstItem.Find(".titleComic").Text())
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detectSpolier(chapURL, ".comicDetail2#lightgallery2", "img")
		if err != nil {
			return
		}
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	return
}

/* Comic page crawler function */
func crawlBlogTruyen(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName string

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.DateFormat = "02/01/2006 15:04"
	comic.ImageURL, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".list-wrap#list-chapters").Find("p:nth-child(1)")
	if firstItem.Nodes == nil {
		return ErrInvalidURL
	}

	chapURL, _ = firstItem.Find(".title").Find("a[href]").Attr("href")
	chapName = firstItem.Find(".title").Find("a[href]").Text()

	chapURL = "https://blogtruyen.vn" + chapURL
	if comic.ChapURL == chapURL {
		return ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detectSpolier(chapURL, "#content", "img[src]")
		if err != nil {
			return
		}
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	return
}

// Implement later
func crawlMangaK(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName string

	comic.Name = doc.Find(".entry-title").Text()
	comic.ImageURL, _ = doc.Find(".info_image").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".row:nth-child(1)")
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	chapName = firstItem.Find("span:nth-child(1)").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detectSpolier(chapURL, ".vung_doc", "img")
		if err != nil {
			return
		}
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	return

}

func getPageSource(pageURL string) (doc *goquery.Document, err error) {

	c := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := c.Get(pageURL)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Danger()
		return
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		logging.Danger()
		return
	}

	return
}

func detectSpolier(chapURL string, attr1, attr2 string) error {

	// Check if chapter is full upload (detect spolier chap)
	doc, err := getPageSource(chapURL)
	if err != nil {
		logging.Danger()
		return err
	}

	if chapSelections := doc.Find(attr1).Find(attr2); chapSelections.Size() < 3 {
		logging.Danger()
		return errors.New("No new chapter, just some spoilers :)")

	}

	return nil
}
