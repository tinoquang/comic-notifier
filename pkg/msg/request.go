package msg

import (
	"context"
	"time"
)

// UserMessage general message form
type UserMessage struct {
	Object  string  `json:"object,omitempty"`
	Entries []Entry `json:"entry,omitempty"`
}

// Entry message from each page
// meaning each POST message can contains more than one message from each page
type Entry struct {
	ID        string      `json:"id,omitempty"` // Page ID
	Time      int         `json:"time,omitempty"`
	Messaging []Messaging `json:"messaging,omitempty"`
}

// Messaging struct per message
type Messaging struct {
	Sender    *User         `json:"sender,omitempty"`    // Page-scope ID different with Page ID
	Recipient *User         `json:"recipient,omitempty"` // Page ID
	Timestamp int           `json:"timestamp,omitempty"`
	Message   *RecvMsg      `json:"message,omitempty"`
	PostBack  *RecvPostBack `json:"postback,omitempty"`
}

// RecvMsg --> message received from user
type RecvMsg struct {
	Mid        string `json:"mid,omitempty"`
	Text       string `json:"text,omitempty"`
	QuickReply *Reply `json:"quick_reply,omitempty"`
}

// Reply : user send msg with type Quickreply (normally when press button)
type Reply struct {
	Payload string `json:"payload,omitempty"`
}

// RecvPostBack --> message when user press button
type RecvPostBack struct {
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// User struct contains ID of Sender and Recipient
type User struct {
	ID string `json:"id,omitempty"`
}

/*---------Request message method------------*/

// Handle text message from user
// Only handle comic page link, other message type is discarded
func (h *Handler) handleText(msg Messaging) {

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	h.svi.HandleTxtMsg(ctx, msg.Sender.ID, msg.Message.Text)

}

func (h *Handler) handlePostback(msg Messaging) {

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	h.svi.HandlePostback(ctx, msg.Sender.ID, msg.PostBack.Payload)
	return
}

func (h *Handler) handleQuickReply(msg Messaging) {

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	if msg.Message.QuickReply.Payload == "Not unsub" {
		return
	}

	h.svi.HandleQuickReply(ctx, msg.Sender.ID, msg.Message.QuickReply.Payload)

	return
}
