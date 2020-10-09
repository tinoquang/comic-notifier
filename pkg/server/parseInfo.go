package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

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

	err = crawler[comic.Page](ctx, doc, comic)
	return errors.Wrapf(err, "Can't get latest chap")
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
