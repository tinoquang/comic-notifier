package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

/* -------------Message response format----------- */

// Response : general response format
type Response struct {
	Type      string   `json:"messaging_type,omitempty"`
	Recipient *User    `json:"recipient,omitempty"`
	Message   *RespMsg `json:"message,omitempty"`
	Action    string   `json:"sender_action,omitempty"`
	Tag       string   `json:"tag,omitempty"`
}

// RespMsg : message content, include some type: text, template, quick-reply,..
type RespMsg struct {
	Text     string       `json:"text,omitempty"`
	Template *Attachment  `json:"attachment,omitempty"`
	Options  []QuickReply `json:"quick_replies,omitempty"`
}

// Attachment such as image, link
type Attachment struct {
	Type     string   `json:"type,omitempty"`
	Payloads *Payload `json:"payload,omitemtpy"`
}

// Payload : attachment content, usually image, button, ...
type Payload struct {
	TemplateType string    `json:"template_type,omitempty"`
	Elements     []Element `json:"elements,omitempty"`
}

// Element : template elements
type Element struct {
	Title         string   `json:"title,omitempty"`
	ImgURL        string   `json:"image_url,omitempty"`
	Subtitle      string   `json:"subtitle,omitempty"`
	DefaultAction *Action  `json:"default_action,omitempty"`
	Buttons       []Button `json:"buttons,omitempty"`
}

// Action : contains URL of comic
type Action struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Button : button include in attachment, example: Read button, Unsubscribe button,...
type Button struct {
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// QuickReply : button, link,... generate quick-reply request when user click
type QuickReply struct {
	Type    string `json:"content_type,omitempty"`
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
	ImgURL  string `json:"image_url,omitempty"`
}

// User contain msg.Sender.ID or msg.Recipient.ID
type User struct {
	ID string `json:"id,omitempty"`
}

func sendTextBack(senderid, message string) {

	res := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: senderid},
		Message:   &RespMsg{Text: message},
	}

	callSendAPI(res)
}

func sendActionBack(senderid, action string) {

	res := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: senderid},
		Action:    action,
	}

	callSendAPI(res)
}

// Use to send message within 24-hour window of FACEBOOK policy
func sendNormalReply(senderid string, comic *db.Comic) {

	response := &Response{
		Recipient: &User{ID: senderid},
		Message: &RespMsg{
			Template: &Attachment{
				Type: "template",
				Payloads: &Payload{
					TemplateType: "generic",
					Elements: []Element{
						{
							Title:    comic.Name + "\n" + comic.LatestChap,
							ImgURL:   comic.CloudImgUrl,
							Subtitle: comic.Page,
							// DefaultAction: &Action{
							// 	Type: "web_url",
							// 	URL:  comic.ChapURL,
							// },
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   comic.ChapUrl,
									Title: "Đọc chap mới",
								},
								{
									Type:    "postback",
									Title:   "Hủy đăng ký",
									Payload: strconv.Itoa(int(comic.ID)),
								},
							},
						},
					},
				},
			},
		},
	}

	callSendAPI(response)
}

func sendMsgTagsReply(senderid string, comic *db.Comic) {

	response := &Response{
		Recipient: &User{ID: senderid},
		Message: &RespMsg{
			Template: &Attachment{
				Type: "template",
				Payloads: &Payload{
					TemplateType: "generic",
					Elements: []Element{
						{
							Title:    comic.Name + "\n" + comic.LatestChap,
							ImgURL:   comic.CloudImgUrl,
							Subtitle: comic.Page,
							// DefaultAction: &Action{
							// 	Type: "web_url",
							// 	URL:  comic.ChapUrl,
							// },
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   comic.ChapUrl,
									Title: "Đọc chap mới",
								},
								{
									Type:    "postback",
									Title:   "Hủy đăng ký",
									Payload: strconv.Itoa(int(comic.ID)),
								},
							},
						},
					},
				},
			},
		},
		Type: "MESSAGE_TAG",
		Tag:  "CONFIRMED_EVENT_UPDATE",
	}

	callSendAPI(response)
	return
}

func sendQuickReplyChoice(senderid string, comic db.Comic) {

	// send back quick reply "Are you sure ?" for user to confirm
	response := &Response{
		Recipient: &User{ID: senderid},
		Type:      "RESPONSE",
		Message: &RespMsg{
			Text: fmt.Sprintf("Hủy nhận thông báo khi %s update chap mới ?", comic.Name),
			Options: []QuickReply{
				{
					Type:    "text",
					Title:   "Yes",
					Payload: strconv.Itoa(int(comic.ID)),
					ImgURL:  "https://www.vhv.rs/dpng/d/356-3568543_check-icon-green-tick-hd-png-download.png",
				},
				{
					Type:    "text",
					Title:   "No",
					Payload: "Not unsub",
					ImgURL:  "https://cdn3.vectorstock.com/i/1000x1000/59/87/red-cross-check-mark-icon-simple-style-vector-8375987.jpg",
				},
			},
		},
	}
	callSendAPI(response)
}

func callSendAPI(r *Response) {

	body := new(bytes.Buffer)
	encoder := json.NewEncoder(body)

	if err := encoder.Encode(&r); err != nil {
		logging.Danger(err)
		return
	}

	request, err := http.NewRequest("POST", messengerEndpoint, body)
	if err != nil {
		logging.Danger(err)
		return
	}

	// Add header and query params for request
	request.Header.Add("Content-Type", "application/json")
	q := request.URL.Query()
	q.Add("access_token", pageToken)
	request.URL.RawQuery = q.Encode()

	// Create client to send request
	client := http.Client{}

	// Send POST message to FACEBOOK API
	_, err = client.Do(request)
	if err != nil {
		logging.Danger(err)
	}

	// defer resp.Body.Close()
	// respBody, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	logging.Danger(err)
	// }

	// fmt.Println(string(respBody))
	return
}
