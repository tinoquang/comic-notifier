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

	// // Validate users is in user DB or not
	// // If not, add user to database, return "Subscribed to ..."
	// // else return "Already subscribed"
	// user, err := data.GetUser("page_id", m.Sender.ID)
	// if err != nil {

	// 	util.Info("Add new user")
	// 	user.PageID = m.Sender.ID

	// 	// Check user already exist
	// 	err = user.GetInfoByPageID()
	// 	if err != nil {
	// 		util.Danger(err)
	// 		return
	// 	}

	// 	err = data.AddUser(&user)

	// 	if err != nil {
	// 		sendTextBack(m.Sender.ID, "Server busy, try again later")
	// 		return
	// 	}
	// }

	// subscriber, err := data.GetSubscriber(user.ID, comic.ID)
	// if err != nil {
	// 	subscriber.UserID = user.ID
	// 	subscriber.ComicID = comic.ID
	// 	err = data.AddSubscriber(&subscriber)
	// 	if err != nil {
	// 		util.Danger(err)
	// 		sendTextBack(m.Sender.ID, "Try again later")
	// 		sendActionBack(m.Sender.ID, "typing_off")
	// 		return
	// 	}
	// 	sendTextBack(m.Sender.ID, "Subscribed")
	// } else {
	// 	sendTextBack(m.Sender.ID, "Already subscribed")
	// }

	// // Call send API
	// util.Info("Parsing complelte, send URL back to user")
	// // send back message in template with buttons
	// sendNormalReply(m.Sender.ID, &comic)
}
