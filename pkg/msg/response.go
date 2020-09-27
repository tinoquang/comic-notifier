package msg

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

func (mh *msgHandler) sendTextBack(message string) {

	res := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: mh.getID("sender")},
		Message:   &RespMsg{Text: message},
	}

	callSendAPI(res)
}

func (mh *msgHandler) sendActionBack(action string) {

	res := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: mh.getID("sender")},
		Action:    action,
	}
	util.Info("Send action " + action + " to user")

	callSendAPI(res)
}

// Use to send message within 24-hour window of FACEBOOK policy
func (mh *msgHandler) sendNormalReply(comic *model.Comic) {

	response := &Response{
		Recipient: &User{ID: mh.getID("sender")},
		Message: &RespMsg{
			Template: &Attachment{
				Type: "template",
				Payloads: &Payload{
					TemplateType: "generic",
					Elements: []Element{
						{
							Title:    comic.Name + comic.LatestChap,
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
									Payload: comic.URL,
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
