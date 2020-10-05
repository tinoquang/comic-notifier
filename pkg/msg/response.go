package msg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

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
	// util.Info("Send action " + action + " to user")

	callSendAPI(res)
}

// Use to send message within 24-hour window of FACEBOOK policy
func sendNormalReply(senderid string, comic *model.Comic) {

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
							ImageURL: comic.ImageURL,
							Subtitle: comic.Page,
							DefaultAction: &Action{
								Type: "web_url",
								URL:  comic.ChapURL,
							},
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   comic.ChapURL,
									Title: "Read Now",
								},
								{
									Type:    "postback",
									Title:   "Unsubscribe",
									Payload: strconv.Itoa(comic.ID),
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

func sendMsgTagsReply(senderid string, comic *model.Comic) {

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
							ImageURL: comic.ImageURL,
							Subtitle: comic.Page,
							DefaultAction: &Action{
								Type: "web_url",
								URL:  comic.ChapURL,
							},
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   comic.ChapURL,
									Title: "Read Now",
								},
								{
									Type:    "postback",
									Title:   "Unsubscribe",
									Payload: strconv.Itoa(comic.ID),
								},
							},
						},
					},
				},
			},
		},
		MessagingType: "MESSAGE_TYPE",
		Tag:           "CONFIRMED_EVENT_UPDATE",
	}

	callSendAPI(response)
	return
}

func sendQuickReplyChoice(senderid string, c *model.Comic) {

	// send back quick reply "Are you sure ?" for user to confirm
	response := &Response{
		Recipient: &User{ID: senderid},
		Type:      "RESPONSE",
		Message: &RespMsg{
			Text: "Unsubscribe " + c.Name + "\nAre you sure ?",
			Options: []QuickReply{
				{
					Type:     "text",
					Title:    "Yes",
					Payload:  strconv.Itoa(c.ID),
					ImageURL: "https://www.vhv.rs/dpng/d/356-3568543_check-icon-green-tick-hd-png-download.png",
				},
				{
					Type:     "text",
					Title:    "No",
					Payload:  "Not unsub",
					ImageURL: "https://cdn3.vectorstock.com/i/1000x1000/59/87/red-cross-check-mark-icon-simple-style-vector-8375987.jpg",
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
		util.Danger(err)
		return
	}

	request, err := http.NewRequest("POST", messengerEndpoint, body)
	if err != nil {
		util.Danger(err)
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
		util.Danger(err)
	}

	return
}
