package msg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*---------Request message method------------*/

// Handle text message from user
// Only handle comic page link, other message type is discarded
func handleText(svr ServerInterface, msg Messaging) {

	sendActionBack(msg.Sender.ID, "typing_on")
	defer sendActionBack(msg.Sender.ID, "typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	comic, err := svr.SubscribeComic(ctx, "psid", msg.Sender.ID, msg.Message.Text)
	if err != nil {
		sendTextBack(msg.Sender.ID, err.Error())
		return
	}

	sendTextBack(msg.Sender.ID, "Subscribed")

	// send back message in template with buttons
	sendNormalReply(msg.Sender.ID, comic)
}

func handlePostback(svr ServerInterface, msg Messaging) {

	sendActionBack(msg.Sender.ID, "typing_on")
	defer sendActionBack(msg.Sender.ID, "typing_off")

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	comicID, _ := strconv.Atoi(msg.PostBack.Payload)

	c, err := svr.GetUserComic(ctx, msg.Sender.ID, comicID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendTextBack(msg.Sender.ID, fmt.Sprintf("Comic %s is not subscribed", c.Name))
			return
		}
		return
	}

	sendQuickReplyChoice(msg.Sender.ID, c)
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

	comicID, err := strconv.Atoi(msg.Message.QuickReply.Payload)

	c, _ := svr.GetUserComic(ctx, msg.Sender.ID, comicID)

	err = svr.UnsubscribeComic(ctx, msg.Sender.ID, comicID)
	if err != nil {
		sendTextBack(msg.Sender.ID, "Please try again later")
	} else {
		sendTextBack(msg.Sender.ID, fmt.Sprintf("Unsubscribe %s", c.Name))
	}
	return
}
