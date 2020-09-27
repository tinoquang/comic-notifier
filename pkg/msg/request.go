package msg

import (
	"context"
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

	comic, err := mh.svr.SubscribeComic(ctx, "psid", mh.getID("sender"), mh.getUserMsg())
	if err != nil {
		mh.sendTextBack(err.Error())
		return
	}

	mh.sendTextBack("Subscribed")

	// send back message in template with buttons
	mh.sendNormalReply(comic)
}
