package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Crawler contain comic, user and image crawler
type Crawler struct {
	*comicCrawler
	*firebaseImg
}

// NewCrawler constructor
func NewCrawler() *Crawler {

	return &Crawler{
		newComicCrawler(crawlHelperWrapper{}),
		newFirebaseConnection(),
	}
}

// GetComicInfo return link of latest chapter of a page
func (crwl *Crawler) GetComicInfo(ctx context.Context, comic *db.Comic) (err error) {

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

	return crwl.crawl(ctx, comic)
}

// GetUserInfoFromFacebook call facebook API to get user info, include psid, appid and profile picture
func (crwl *Crawler) GetUserInfoFromFacebook(field, id string, user *db.CreateUserParams) error {

	info := map[string]json.RawMessage{}
	appInfo := []map[string]json.RawMessage{}
	picture := map[string]json.RawMessage{}
	queries := map[string]string{}

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
		return fmt.Errorf("Wrong field request, field: %s", field)
	}

	respBody, err := util.MakeGetRequest(fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, id), queries)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &info)
	if err != nil {
		return err
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

	return nil
}
