package crawler

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// beeng crawler
type (
	beeng           struct{}
	blogtruyen      struct{}
	mangaK          struct{}
	truyentranhtuan struct{}
	truyentranhnet  struct{}
)

func (b beeng) crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error) {

	var chapURL string

	doc, err := getPageSource(comic.URL)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find(".detail").Find("h4").Text()
	comic.OriginImgURL, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")

	// Find latest chap
	firstItem := doc.Find(".listChapters").Find(".list").Find("li:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = strings.TrimSpace(firstItem.Find(".titleComic").Text())
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return util.ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detector.detect(chapURL, ".comicDetail2#lightgallery2", "img")
		if err != nil {
			return
		}
	}

	comic.ChapURL = chapURL
	return
}

/* ------------------------------------------------------------------------------------------------ */

// blogtruyen crawler
func (b blogtruyen) crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error) {

	var chapURL string

	doc, err := getPageSource(comic.URL)
	if err != nil {
		return util.ErrInvalidURL
	}

	name, _ := doc.Find(".entry-title").Find("a[title]").Attr("title")
	comic.Name = strings.TrimLeft(strings.TrimSpace(name), "truyện tranh")
	comic.OriginImgURL, _ = doc.Find(".thumbnail").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".list-wrap#list-chapters").Find("p:nth-child(1)")
	if firstItem.Nodes == nil {
		return util.ErrInvalidURL
	}

	comic.LatestChap = firstItem.Find(".title").Find("a[href]").Text()
	chapURL, _ = firstItem.Find(".title").Find("a[href]").Attr("href")

	chapURL = "https://blogtruyen.vn" + chapURL
	if comic.ChapURL == chapURL {
		return util.ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detector.detect(chapURL, "#content", "img[src]")
		if err != nil {
			return
		}
	}

	comic.ChapURL = chapURL
	return
}

/* ------------------------------------------------------------------------------------------------------------------- */

// mangaK crawler
func (m mangaK) crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error) {

	var chapURL string

	doc, err := getPageSource(comic.URL)
	if err != nil {
		return util.ErrInvalidURL
	}
	logging.Info(doc.Html())

	comic.Name = doc.Find(".entry-title").Text()
	comic.OriginImgURL, _ = doc.Find(".info_image").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".row:nth-child(1)")
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("span:nth-child(1)").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return util.ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detector.detect(chapURL, ".vung_doc", "img")
		if err != nil {
			return
		}
	}

	comic.ChapURL = chapURL
	return

}

/* ---------------------------------------------------------------------------------------------------- */

func (t truyentranhtuan) crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error) {

	var chapURL string

	doc, err := getPageSource(comic.URL)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find("#infor-box").Find("h1").Text()
	comic.OriginImgURL, _ = doc.Find(".manga-cover").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find("#manga-chapter").Find(".chapter-name").First()
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return util.ErrComicUpToDate
	}

	// if comic.ChapURL != "" {
	// 	err = detector.detect(chapURL, ".vung_doc", "img")
	// 	if err != nil {
	// 		return
	// 	}
	// }

	comic.ChapURL = chapURL
	return
}

func (t truyentranhnet) crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error) {

	var chapURL string

	url := comic.URL + "?order=desc"
	doc, err := getPageSource(url)
	if err != nil {
		return util.ErrInvalidURL
	}

	comic.Name = doc.Find(".detail-manga-title").Find("h1").Text()
	comic.OriginImgURL, _ = doc.Find(".detail-img").Find("img[src]").Attr("src")

	// Find latest chap
	firstItem := doc.Find(".chapter-list").Find(".chapter-select").First()
	if firstItem.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	comic.LatestChap = firstItem.Find("a[href]").Text()
	chapURL, _ = firstItem.Find("a[href]").Attr("href")

	if chapURL == comic.ChapURL {
		return util.ErrComicUpToDate
	}

	if comic.ChapURL != "" {
		err = detector.detect(chapURL, ".manga-reading-box", "img")
		if err != nil {
			return
		}
	}

	comic.ChapURL = chapURL
	return
}
