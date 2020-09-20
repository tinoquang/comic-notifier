package msg

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/tinoquang/comic-notifier/pkg/util"
)

func sendTextBack(psid string, message string) {

	response := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: psid},
		Message:   &RespMsg{Text: message},
	}

	callSendAPI(response)
}

func sendActionBack(psid string, action string) {
	response := &Response{
		Type:      "RESPONSE",
		Recipient: &User{ID: psid},
		Action:    action,
	}
	util.Info("Send action " + action + " to user")

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
