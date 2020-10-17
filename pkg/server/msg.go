package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	cfg   *conf.Config
	store *store.Stores
}

// NewMSG return new api interface
func NewMSG(c *conf.Config, s *store.Stores) *MSG {
	return &MSG{cfg: c, store: s}
}

/* Message handler function */

// HandleTxtMsg handle text messages from facebook user
func (m *MSG) HandleTxtMsg(ctx context.Context, senderID, text string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	comic, err := subscribeComic(ctx, m.cfg, m.store, "psid", senderID, text)
	if err != nil {
		sendTextBack(senderID, err.Error())
		return
	}

	sendTextBack(senderID, "Subscribed")

	// send back message in template with buttons
	sendNormalReply(senderID, comic)

}

// HandlePostback handle messages when user click "Unsucsribe button"
func (m *MSG) HandlePostback(ctx context.Context, senderID, payload string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	comicID, _ := strconv.Atoi(payload)

	c, err := m.store.Comic.GetByPSID(ctx, senderID, comicID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendTextBack(senderID, fmt.Sprintf("Comic %s is not subscribed", c.Name))
			return
		}
		return
	}

	sendQuickReplyChoice(senderID, c)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {
	comicID, err := strconv.Atoi(payload)

	c, _ := m.store.Comic.GetByPSID(ctx, senderID, comicID)

	err = m.store.Subscriber.Delete(ctx, senderID, comicID)
	if err != nil {
		sendTextBack(senderID, "Please try again later")
	} else {
		sendTextBack(senderID, fmt.Sprintf("Unsubscribe %s", c.Name))
	}
}

// Update comic helper
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

			comic = &model.Comic{
				Page: parsedURL.Hostname(),
				URL:  comicURL,
			}
			// Get all comic infos includes latest chapter
			err = crawler.GetComicInfo(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
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

	util.Info("Added new comic: ", comic.Name)
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

	_, err = store.Subscriber.Get(ctx, user.PSID, comic.ID)
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

// GetUserInfoByID get user AppID using PSID or vice-versa
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
