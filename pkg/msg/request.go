package msg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/util"
)

/*---------Request message method------------*/
func (mh *msgHandler) getID(name string) string {
	switch name {
	case "sender":
		return mh.req.Sender.ID
	default:
		util.Warning("Invalid type to get ID")
		return ""
	}
}

func (mh *msgHandler) getUserMsg() string {
	return mh.req.Message.Text
}

// Handle text message from user
// Only handle comic page link, other message type is discarded
func (mh *msgHandler) handleText() {

	mh.sendActionBack("typing_on")
	defer mh.sendActionBack("typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	subID, comic, err := mh.svr.SubscribeComic(ctx, "psid", mh.getID("sender"), mh.getUserMsg())
	if err != nil {
		mh.sendTextBack(err.Error())
		return
	}

	mh.sendTextBack("Subscribed")

	// send back message in template with buttons
	mh.sendNormalReply(subID, comic)
}

func (mh *msgHandler) handlePostback() {

	mh.sendActionBack("typing_on")
	defer mh.sendActionBack("typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	subID, err := strconv.Atoi(mh.req.PostBack.Payload)

	s, err := mh.svr.GetSubscriber(ctx, subID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			mh.sendTextBack(fmt.Sprintf("Comic %s is not subscribed", s.ComicName))
			return
		}
		util.Danger(err)
		return
	}

	mh.sendQuickReplyChoice(s)
	return
}

func (mh *msgHandler) handleQuickReply() {

	mh.sendActionBack("typing_on")
	defer mh.sendActionBack("typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	if mh.req.Message.QuickReply.Payload == "Not unsub" {
		return
	}

	subID, err := strconv.Atoi(mh.req.Message.QuickReply.Payload)

	s, err := mh.svr.GetSubscriber(ctx, subID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			mh.sendTextBack(fmt.Sprintf("Comic %s is not subscribed", s.ComicName))
			return
		}
		util.Danger(err)
		return
	}

	err = mh.svr.UnsubscribeComic(ctx, subID)
	if err != nil {
		mh.sendTextBack("Please try again later")
	} else {
		mh.sendTextBack(fmt.Sprintf("Unsubscribe %s", s.ComicName))
	}
	return
}
