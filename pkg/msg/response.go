package msg

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/tinoquang/comic-notifier/pkg/util"
)

func sendTextBack(psid string, message string) {

	response := Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: psid},
		Message:   &RespMsg{Text: message},
	}

	response.callSendAPI()
	return

}

func sendActionBack(psid string, action string) {
	response := Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: psid},
		Action:    action,
	}
	util.Info("Send action " + action + " to user")
	response.callSendAPI()

}

func (r *Response) callSendAPI() {

	body := new(bytes.Buffer)
	encoder := json.NewEncoder(body)

	if err := encoder.Encode(&r.Message.Text); err != nil {
		util.Danger(err)
		return
	}

	request, err := http.NewRequest("POST", glob.FacebookURL, body)
	if err != nil {
		util.Danger(err)
		return
	}

	// Add header and query params for request
	request.Header.Add("Content-Type", "application/json")
	q := request.URL.Query()
	q.Add("access_token", glob.PageToken)
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
