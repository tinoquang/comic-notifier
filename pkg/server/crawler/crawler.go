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
	"github.com/tinoquang/comic-notifier/pkg/conf"
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
func New(cfg *conf.Config) {

	crawlerMap = make(map[string]func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error))
	crawlerMap["beeng.net"] = crawlBeeng
	crawlerMap["truyendep.com"] = crawlMangaK // mangaK is changed to truyendep.com
	crawlerMap["truyenqq.com"] = crawlTruyenqq
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

	var chapURL, chapName, chapDate string
	var max time.Time

	comic.Name = doc.Find(".detail").Find("h4").Text()
	comic.DateFormat = "02-01-2006 15:04"
	comic.ImageURL, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")

	// Query latest chap
	selections := doc.Find(".listChapters").Find(".list").Find("li")
	if selections.Nodes == nil {
		return ErrInvalidURL
	}

	// Find latest chap
	selections.Each(func(index int, item *goquery.Selection) {

		rawDate := strings.Split((item.Find(".name").Find("span:nth-child(2)").Text()), "-")
		date := fmt.Sprintf("%s-%s-%s", strings.TrimSpace(rawDate[0]), rawDate[1], strings.TrimSpace(rawDate[2]))

		t, err := time.Parse("02-01-2006 15:04", date)
		if err != nil {
			fmt.Println(err)
		}
		if t.Sub(max) > 0 {
			max = t

			text := strings.Fields(strings.Replace(item.Find(".titleComic").Text(), ":", "", -1))
			chapName = strings.Join(text, " ")
			chapURL, _ = item.Find("a[href]").Attr("href")
			chapDate = date
		}
	})

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
	comic.Date = chapDate
	return
}

/* Comic page crawler function */
func crawlBlogTruyen(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName, chapDate string
	var max time.Time

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.DateFormat = "02/01/2006 15:04"
	comic.ImageURL, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")

	// Query latest chap
	selections := doc.Find(".list-wrap#list-chapters").Find("p")
	if selections.Nodes == nil {
		return ErrInvalidURL
	}

	// Find latest chap
	selections.Each(func(index int, item *goquery.Selection) {

		date := item.Find(".publishedDate").Text()
		t, _ := time.Parse(comic.DateFormat, date)

		if t.Sub(max) > 0 {
			max = t
			chapURL, _ = item.Find(".title").Find("a[href]").Attr("href")
			chapName = item.Find(".title").Find("a[href]").Text()
			chapDate = date
		}
	})

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
	comic.Date = chapDate
	return
}

// Implement later
func crawlMangaK(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	max := time.Time{}

	var chapURL, chapName string

	comic.Name = doc.Find(".entry-title").Text()
	comic.DateFormat = "02-01-2006" // DD-MM-YYYY
	comic.ImageURL, _ = doc.Find(".info_image").Find("img[src]").Attr("src")

	// Query latest chap
	selections := doc.Find(".chapter-list").Find(".row")
	if selections.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	// Find latest chap
	selections.Each(func(index int, item *goquery.Selection) {

		date := strings.TrimSpace(item.Find("span:nth-child(2)").Text())
		t, _ := time.Parse(comic.DateFormat, date)

		if t.Sub(max).Seconds() > 0.0 {
			max = t
			chapName = item.Find("span:nth-child(1)").Text()
			chapURL, _ = item.Find("a[href]").Attr("href")
		}
	})

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

// Implement later
func crawlTruyenqq(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName, chapDate string
	var max time.Time
	var chapDoc *goquery.Document

	comic.Name = doc.Find(".center").Find("h1").Text()
	// comic.ImageURL, _ = doc.Find(".left").Find("img[src]").Attr("src")

	// u, err := url.Parse(imageURL)
	// u.Host = "truyenqq.com"
	// chap.ImageURL = u.String()

	comic.DateFormat = "02/01/2006"

	selections := doc.Find(".works-chapter-item.row")
	if selections.Nodes == nil {
		return errors.New("Please check your URL")
	}

	max = time.Time{}
	// Iterate through all element which has same container to find link and chapter number
	selections.Each(func(index int, item *goquery.Selection) {
		// Get date release
		date := strings.TrimSpace(item.Find(".col-md-2.col-sm-2.col-xs-4.text-right").Text())

		t, err := time.Parse("02/01/2006", date)
		if err != nil {
			logging.Danger(err)
			return
		}

		if t.Sub(max).Seconds() > 0.0 {
			max = t
			chapName = item.Find("a[href]").Text()
			chapURL, _ = item.Find("a").Attr("href")
			chapDate = date
		}

	})

	// Check if chapter is full uploaded (to avoid spolier chap)
	chapDoc, err = getPageSource(chapURL)
	if err != nil {
		logging.Danger()
		return
	}

	if chapSelections := chapDoc.Find(".story-see-content").Find("img[src]"); chapSelections.Size() < 3 {
		logging.Danger()
		return errors.New("No new chapter, just some spoilers :)")
	}

	if comic.ChapURL == chapURL {
		return ErrComicUpToDate
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	comic.Date = chapDate
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
