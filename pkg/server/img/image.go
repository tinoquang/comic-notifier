package img

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	apiEndpoint  string
	accessToken  string
	refreshToken string
)

// Data --> imageResponse content
type Data struct {
	ID         string `json:"id"`
	DeleteHash string `json:"deletehash"`
	Link       string `json:"link"`
}

// response when create add new image to Imgur gallery
type imageResponse struct {
	Data *Data
}

// SetEnvVar set environment var for interacting with Imgur API
func SetEnvVar(cfg *conf.Config) {
	apiEndpoint = cfg.Imgur.Endpoint
	accessToken = cfg.Imgur.AccessToken
	refreshToken = cfg.Imgur.RefreshToken
}

// UploadImagetoImgur add image to Imgur gallery and return link to new image
func UploadImagetoImgur(title string, imageURL string) (string, error) {

	response := &imageResponse{}
	url := apiEndpoint + "image"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	writer.WriteField("image", imageURL)
	writer.WriteField("type", "url")
	writer.WriteField("title", title)

	err := writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		util.Danger(err)
		return "", err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(response)

	return response.Data.Link, err
}
