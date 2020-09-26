package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// type userInfo struct {
// 	name string
// 	picture date {

// 	}
// }

type picture struct {
}

func (s *Server) getUserInfoByID(field, id string) (user *model.User, err error) {

	info := map[string]json.RawMessage{}
	appInfo := []map[string]json.RawMessage{}
	picture := map[string]json.RawMessage{}
	queries := map[string]string{}

	switch field {
	case "psid":
		queries["fields"] = "name,picture.width(500).height(500),ids_for_apps"
		queries["access_token"] = s.cfg.FBSecret.PakeToken
	case "appid":
		queries["fields"] = "name,ids_for_pages,picture.width(500).height(500)"
		queries["access_token"] = s.cfg.FBSecret.AppToken
		queries["appsecret_proof"] = s.cfg.FBSecret.AppSecret
	default:
		return nil, errors.New(fmt.Sprintf("Wrong field request, field: %s", field))
	}

	respBody, err := util.MakeGetRequest(s.cfg.Webhook.GraphEndpoint+id, queries)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &info)
	if err != nil {
		return
	}

	user = new(model.User)
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
