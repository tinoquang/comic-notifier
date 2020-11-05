package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// MakeGetRequest send HTTP GET request with mapped queries
func MakeGetRequest(URL string, queries map[string]string) (respBody []byte, err error) {

	reqURL, err := url.Parse(URL)
	if err != nil {
		return
	}

	q := reqURL.Query()
	for key, val := range queries {
		q.Add(key, val)
	}
	reqURL.RawQuery = q.Encode()

	c := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := c.Get(reqURL.String())
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	respBody, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	return
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
		return nil, fmt.Errorf("Wrong field request, field: %s", field)
	}

	respBody, err := MakeGetRequest(fmt.Sprintf("%s/%s", cfg.Webhook.GraphEndpoint, id), queries)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &info)
	if err != nil {
		return
	}

	user.Name = ConvertJSONToString(info["name"])

	if field == "psid" {
		json.Unmarshal(info["ids_for_apps"], &info)
	} else {
		json.Unmarshal(info["ids_for_pages"], &info)
	}

	json.Unmarshal(info["picture"], &picture)
	json.Unmarshal(picture["data"], &picture)
	json.Unmarshal(info["data"], &appInfo)

	if field == "psid" {
		user.AppID = ConvertJSONToString(appInfo[0]["id"])

	} else {
		user.PSID = ConvertJSONToString(appInfo[0]["id"])
	}
	user.ProfilePic = ConvertJSONToString(picture["url"])
	user.ProfilePic = strings.Replace(user.ProfilePic, "\\", "", -1)

	return user, nil
}
