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

// Handle text message from user
// Only handle comic page link, other message type is discarded
func handleText(svr ServerInterface, msg Messaging) {

	sendActionBack(msg.Sender.ID, "typing_on")
	defer sendActionBack(msg.Sender.ID, "typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	subID, comic, err := svr.SubscribeComic(ctx, "psid", msg.Sender.ID, msg.Message.Text)
	if err != nil {
		sendTextBack(msg.Sender.ID, err.Error())
		return
	}

	sendTextBack(msg.Sender.ID, "Subscribed")

	// send back message in template with buttons
	sendNormalReply(msg.Sender.ID, subID, comic)
}

func handlePostback(svr ServerInterface, msg Messaging) {

	sendActionBack(msg.Sender.ID, "typing_on")
	defer sendActionBack(msg.Sender.ID, "typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	subID, err := strconv.Atoi(msg.PostBack.Payload)

	s, err := svr.GetSubscriber(ctx, subID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendTextBack(msg.Sender.ID, fmt.Sprintf("Comic %s is not subscribed", s.ComicName))
			return
		}
		util.Danger(err)
		return
	}

	sendQuickReplyChoice(msg.Sender.ID, s)
	return
}

func handleQuickReply(svr ServerInterface, msg Messaging) {

	sendActionBack(msg.Sender.ID, "typing_on")
	defer sendActionBack(msg.Sender.ID, "typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	if msg.Message.QuickReply.Payload == "Not unsub" {
		return
	}

	subID, err := strconv.Atoi(msg.Message.QuickReply.Payload)

	s, err := svr.GetSubscriber(ctx, subID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendTextBack(msg.Sender.ID, fmt.Sprintf("Comic %s is not subscribed", s.ComicName))
			return
		}
		util.Danger(err)
		return
	}

	err = svr.UnsubscribeComic(ctx, subID)
	if err != nil {
		sendTextBack(msg.Sender.ID, "Please try again later")
	} else {
		sendTextBack(msg.Sender.ID, fmt.Sprintf("Unsubscribe %s", s.ComicName))
	}
	return
}
