package msg

import (
	"context"
	"fmt"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comic, err := mh.svr.SubscribeComic(ctx, "psid", mh.getID("sender"), mh.getUserMsg())
	if err != nil {
		mh.sendTextBack(err.Error())
		return
	}

	fmt.Println("comic", comic)

	// Call send API
	util.Info("Parsing complelte, send URL back to user")
	// send back message in template with buttons
	// sendNormalReply(m.Sender.ID, &comic)
}
