package msg

import (
	"context"
	"fmt"
	"net/url"
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

func (mh *msgHandler) getContent() string {
	return mh.req.Message.Text
}

// Handle text message from user
// Only handle comic page link, other message type is discarded
func (mh *msgHandler) handleText() {

	mh.sendActionBack("typing_on")
	defer mh.sendActionBack("typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pageURL, err := url.Parse(mh.getContent())
	if err != nil {
		mh.sendTextBack("Please check your link !")
		util.Danger(err)
		return
	}

	// Check page support, if not send back "Page is not supported"
	page, err := mh.svr.GetPage(ctx, pageURL.Hostname())
	if err != nil {
		util.Danger(err)
		mh.sendTextBack("Sorry, page " + pageURL.Hostname() + " is not supported yet!")
		return
	}

	fmt.Println(page)
	// // Page URL validated, now check comics already in database
	// util.Info("Validated " + page.Name)
	// comic, err := data.GetComic(userURL)

	// // If comic is not in database, query it's latest chap,
	// // add to database, then prepare response with latest chapter URL
	// if err != nil {

	// 	util.Info("Comic is not in DB yet, insert it")
	// 	comic.ComicURL = userURL
	// 	err := GetLatestChapter(&comic)
	// 	if err != nil {
	// 		util.Danger(err)
	// 		sendTextBack(m.Sender.ID, "Please check your URL !!!")
	// 		return
	// 	}

	// 	err = data.AddComic(&comic)
	// 	if err != nil {
	// 		util.Danger(err)
	// 		sendTextBack(m.Sender.ID, "Try again later !")
	// 		return
	// 	}
	// }

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
