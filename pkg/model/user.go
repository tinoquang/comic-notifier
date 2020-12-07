package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// User model
type User struct {
	Name       string `json:"name"`
	PSID       string `json:"psid"`
	AppID      string `json:"appid"`
	ProfilePic string `json:"profile-pic"`
}

// UserList contains multiple users
type UserList struct {
	Users []User `json:"users"`
}

// Session model
type Session struct {
	ID    int    `json:"id"`
	UUID  string `json:"uuid"`
	AppID string `json:"appid"`
}

// GetInfoFromFB get user AppID using PSID or vice-versa
func (u *User) GetInfoFromFB(field, id string) error {

	info := map[string]json.RawMessage{}
	appInfo := []map[string]json.RawMessage{}
	picture := map[string]json.RawMessage{}
	queries := map[string]string{}

	switch field {
	case "psid":
		u.PSID = id
		queries["fields"] = "name,picture.width(500).height(500),ids_for_apps"
		queries["access_token"] = conf.Cfg.FBSecret.PakeToken
	case "appid":
		u.AppID = id
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

	u.Name = util.ConvertJSONToString(info["name"])

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
			u.AppID = util.ConvertJSONToString(appInfo[0]["id"])
		} else {
			u.PSID = util.ConvertJSONToString(appInfo[0]["id"])
		}
	}

	u.ProfilePic = util.ConvertJSONToString(picture["url"])
	u.ProfilePic = strings.Replace(u.ProfilePic, "\\", "", -1)

	return nil
}
