package crawler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type crawler struct {
	*comicCrawler
}

// NewCrawler constructor
func NewCrawler() *crawler {

	return &crawler{
		newComicCrawler(crawlHelper{}),
	}
}

// GetUserInfoFromFacebook call facebook API to get user info, include psid, appid and profile picture
func (crwl *crawler) GetUserInfoFromFacebook(field, id string) (user db.User, err error) {

	err = nil
	user = db.User{}
	info := map[string]json.RawMessage{}
	appInfo := []map[string]json.RawMessage{}
	picture := map[string]json.RawMessage{}
	queries := map[string]string{}

	user.Psid = sql.NullString{String: "", Valid: true}
	user.Appid = sql.NullString{String: "", Valid: true}

	switch field {
	case "psid":
		user.Psid.String = id
		queries["fields"] = "name,picture.width(500).height(500),ids_for_apps"
		queries["access_token"] = conf.Cfg.FBSecret.PakeToken
	case "appid":
		user.Appid.String = id
		queries["fields"] = "name,ids_for_pages,picture.width(500).height(500)"
		queries["access_token"] = conf.Cfg.FBSecret.AppToken
		queries["appsecret_proof"] = conf.Cfg.FBSecret.AppSecret
	default:
		err = fmt.Errorf("Wrong field request, field: %s", field)
		return
	}

	respBody, err := util.MakeGetRequest(fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, id), queries)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &info)
	if err != nil {
		return
	}

	user.Name = util.ConvertJSONToString(info["name"])

	if field == "psid" {
		json.Unmarshal(info["ids_for_apps"], &info)
	} else {
		json.Unmarshal(info["ids_for_pages"], &info)
	}

	json.Unmarshal(info["picture"], &picture)
	json.Unmarshal(picture["data"], &picture)
	json.Unmarshal(info["data"], &appInfo)

	if len(appInfo) != 0 {
		if field == "psid" {
			user.Appid.String = util.ConvertJSONToString(appInfo[0]["id"])
		} else {
			user.Psid.String = util.ConvertJSONToString(appInfo[0]["id"])
		}
	}

	user.ProfilePic.String = strings.Replace(util.ConvertJSONToString(picture["url"]), "\\", "", -1)

	return
}
