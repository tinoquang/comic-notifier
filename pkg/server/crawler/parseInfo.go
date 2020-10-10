package crawler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type comicCrawler func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error)

var crawler map[string]comicCrawler

func init() {

	crawler = make(map[string]comicCrawler)
	crawler["beeng.net"] = crawlBeeng
	crawler["mangak.info"] = crawlMangaK
	crawler["truyenqq.com"] = crawlTruyenqq
	crawler["blogtruyen.vn"] = crawlBlogTruyen

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
			util.Danger()
		}
		return
	}()

	html, err := getPageSource(comic.URL)
	if err != nil {
		return errors.Wrapf(err, "Can't retrieve page's HTML")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return errors.Wrapf(err, "Can't create goquery document object")
	}

	err = crawler[comic.Page](ctx, doc, comic)
	return errors.Wrapf(err, "Can't get latest chap")
}

// GetUserInfoByID get user AppID using PSID or vice-versa
func GetUserInfoByID(cfg *conf.Config, field, id string) (user *model.User, err error) {

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
