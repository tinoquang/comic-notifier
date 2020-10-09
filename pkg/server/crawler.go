package server

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

func detectSpolier(chapURL string, attr1, attr2 string) error {
	var chapDoc *goquery.Document

	// Check if chapter is full upload (detect spolier chap)
	html, err := getPageSource(chapURL)
	if err != nil {
		util.Danger()
		return err
	}

	chapDoc, err = goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		util.Danger()
		return err
	}

	if chapSelections := chapDoc.Find(attr1).Find(attr2); chapSelections.Size() < 3 {
		util.Danger()
		return errors.New("No new chapter, just some spoilers :)")

	}

	return nil
}

func crawlBeeng(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var max float64 = 0
	var chapURL, chapName string

	comic.Name = doc.Find(".detail").Find("h4").Text()
	comic.DateFormat = "02/01/2006"

	// Download cover image of comic
	comic.ImageURL, _ = doc.Find(".cover").Find("img[src]").Attr("data-src")
	// err = comic.DownloadImage(imageURL, "beeng")
	// if err != nil {
	// 	util.Danger("Download image failed, comic:", comic.Name)
	// 	return
	// }

	// Query latest chap
	selections := doc.Find(".listChapters").Find(".list").Find("li")
	if selections.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	// Find latest chap
	selections.Each(func(index int, item *goquery.Selection) {

		text := strings.Fields(strings.Replace(item.Find(".titleComic").Text(), ":", "", -1))
		if len(text) >= 2 && (text[0] == "Chương" || text[0] == "Chapter") {
			chapNum, err := strconv.ParseFloat(text[1], 64)
			if err != nil {
				util.Danger(err)
			}

			if max < chapNum {
				max = chapNum
				chapName = strings.Join(text, " ")
				chapURL, _ = item.Find("a[href]").Attr("href")
			}
		}
	})

	err = detectSpolier(chapURL, ".comicDetail2#lightgallery2", "img")
	if err != nil {
		return
	}

	if chapURL == comic.ChapURL {
		return errors.New("No new chapter")
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
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
	// err = comic.DownloadImage(imageURL, "blogtruyen")
	// if err != nil {
	// 	util.Danger("Download image failed, comic:", comic.Name)
	// 	return
	// }

	// Query latest chap
	selections := doc.Find(".list-wrap#list-chapters").Find("p")
	if selections.Nodes == nil {
		return errors.New("URL is not a comic page")
	}

	// Find latest chap
	selections.Each(func(index int, item *goquery.Selection) {

		date := item.Find(".publishedDate").Text()
		t, _ := time.Parse("02/01/2006 15:04", date)

		if t.Sub(max) > 0 {
			max = t
			chapURL, _ = item.Find(".title").Find("a[href]").Attr("href")
			chapName = item.Find(".title").Find("a[href]").Text()
			chapDate = date
		}
	})

	chapURL = "https://blogtruyen.vn" + chapURL

	err = detectSpolier(chapURL, "#content", "img[src]")
	if err != nil {
		return
	}

	if comic.ChapURL == chapURL {
		return errors.New("No new chapter")
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	comic.Date = chapDate
	return
}

// Implement later
func crawlTruyenqq(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName, chapDate string
	var html []byte
	var max time.Time
	var chapDoc *goquery.Document

	comic.Name = doc.Find(".center").Find("h1").Text()
	comic.ImageURL, _ = doc.Find(".left").Find("img[src]").Attr("src")

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
			util.Danger(err)
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
	html, err = getPageSource(chapURL)
	if err != nil {
		util.Danger()
		return
	}

	chapDoc, err = goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		util.Danger()
		return
	}

	if chapSelections := chapDoc.Find(".story-see-content").Find("img[src]"); chapSelections.Size() < 3 {
		util.Danger()
		return errors.New("No new chapter, just some spoilers :)")
	}

	if comic.ChapURL == chapURL {
		return errors.New("No new chapter")
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	comic.Date = chapDate
	return
}

// Implement later
func crawlMangaK(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	// chap.ComicName = doc.Find(".entry-title").Text()
	// chap.ImageURL, _ = doc.Find(".info_image").Find("img[src]").Attr("src")
	// chap.DateFormat = "02-01-2006"

	// selections := doc.Find(".chapter-list").Find(".row")
	// if selections.Nodes == nil {
	// 	html, _ := doc.Html()
	// 	util.Debug(html)
	// 	return errors.New("URL is not a comic page")
	// }

	// max := time.Time{}
	// // Iterate through all element which has same container to find link and chapter number
	// selections.Each(func(index int, item *goquery.Selection) {
	// 	// Get date release
	// 	date := strings.TrimSpace(item.Find("span:nth-child(2)").Text())

	// 	t, err := time.Parse("02-01-2006", date)
	// 	if err != nil {
	// 		util.Danger(err)
	// 		return
	// 	}

	// 	if t.Sub(max).Seconds() > 0.0 {
	// 		max = t
	// 		chap.Name = item.Find("a[href]").Text()
	// 		chap.URL, _ = item.Find("a").Attr("href")
	// 		chap.Date = date
	// 	}

	// })

	return
}
