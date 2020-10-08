package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

func subscribeComic(ctx context.Context, cfg *conf.Config, store *store.Stores, field string, id string, comicURL string) (*model.Comic, error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	_, err = store.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	// Page URL validated, now check comics already in database
	// util.Info("Validated " + page.Name)
	comic, err := store.Comic.GetByURL(ctx, comicURL)

	// If comic is not in database, query it's latest chap,
	// add to database, then prepare response with latest chapter
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Comic is not in DB yet, insert it")
			comic = &model.Comic{
				Page: parsedURL.Hostname(),
				URL:  comicURL,
			}
			// Get all comic infos includes latest chapter
			err = getComicInfo(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please check your URL")
			}

			// Add new comic to DB
			err = store.Comic.Create(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return nil, errors.New("Please try again later")
		}
	}

	// Validate users is in user DB or not
	// If not, add user to database, return "Subscribed to ..."
	// else return "Already subscribed"
	user, err := store.User.GetByFBID(ctx, field, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Add new user")

			user, err = getUserInfoByID(cfg, field, id)
			// Check user already exist
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
			err = store.User.Create(ctx, user)

			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return nil, errors.New("Please try again later")
		}
	}

	_, err = store.Comic.GetByPSID(ctx, user.PSID, comic.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			subscriber := &model.Subscriber{
				PSID:    user.PSID,
				ComicID: comic.ID,
			}

			err = store.Subscriber.Create(ctx, subscriber)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
			return comic, nil
		}
		util.Danger(err)
		return nil, errors.New("Please try again later")
	}
	return nil, errors.New("Already subscribed")
}

func getUserInfoByID(cfg *conf.Config, field, id string) (user *model.User, err error) {

	user = &model.User{}

	info := map[string]json.RawMessage{}
	appInfo := []map[string]json.RawMessage{}
	picture := map[string]json.RawMessage{}
	queries := map[string]string{}

	switch field {
	case "psid":
		user.PSID = id
		queries["fields"] = "name,picture.width(500).height(500),ids_for_apps"
		queries["access_token"] = cfg.FBSecret.PakeToken
	case "appid":
		user.AppID = id
		queries["fields"] = "name,ids_for_pages,picture.width(500).height(500)"
		queries["access_token"] = cfg.FBSecret.AppToken
		queries["appsecret_proof"] = cfg.FBSecret.AppSecret
	default:
		return nil, errors.New(fmt.Sprintf("Wrong field request, field: %s", field))
	}

	respBody, err := util.MakeGetRequest(cfg.Webhook.GraphEndpoint+id, queries)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &info)
	if err != nil {
		return
	}

	user.Name = util.ConvertJSONToString(info["name"])

	json.Unmarshal(info["ids_for_apps"], &info)
	json.Unmarshal(info["picture"], &picture)
	json.Unmarshal(picture["data"], &picture)
	json.Unmarshal(info["data"], &appInfo)

	user.AppID = util.ConvertJSONToString(appInfo[0]["id"])
	user.ProfilePic = util.ConvertJSONToString(picture["url"])
	user.ProfilePic = strings.Replace(user.ProfilePic, "\\", "", -1)

	return user, nil
}

// getComicInfo return link of latest chapter of a page
func getComicInfo(ctx context.Context, comic *model.Comic) error {

	defer func() {
		if err := recover(); err != nil {
			util.Danger(err)
			return
		}
	}()

	html, err := getPageSource(comic.URL)
	if err != nil {
		return errors.Wrapf(err, "Can't retrieve page's HTML")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return errors.Wrapf(err, "Can't create goquery document object")
	}

	// Find the class contain <a> tag with link of chapter
	err = handler[comic.Page](ctx, doc, comic)
	return errors.Wrapf(err, "Can't get latest chap")
}

func handleBeeng(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var max float64 = 0
	var html []byte
	var chapURL, chapName string
	var chapDoc *goquery.Document

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

	// Check if chapter is full upload (detect spolier chap)
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

	if chapSelections := chapDoc.Find(".comicDetail2#lightgallery2").Find("a[href]"); chapSelections.Size() < 3 {
		util.Danger()
		return errors.New("No new chapter, just some spoilers :)")

	}

	if chapURL == comic.ChapURL {
		return errors.New("No new chapter")
	}

	comic.LatestChap = chapName
	comic.ChapURL = chapURL
	return
}

func getPageSource(pageURL string) (body []byte, err error) {

	resp, err := http.Get(pageURL)

	if err != nil {
		util.Danger(err)
		return
	}

	// do this now so it won't be forgotten
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	return
}

/* Comic page crawler function */
func handleBlogTruyen(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

	var chapURL, chapName, chapDate string
	var html []byte
	var max time.Time
	var chapDoc *goquery.Document

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

	if chapSelections := chapDoc.Find("#content").Find("img[src]"); chapSelections.Size() < 3 {
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

func handleTruyenqq(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

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

func handleMangaK(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error) {

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
