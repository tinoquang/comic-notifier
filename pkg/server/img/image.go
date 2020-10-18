package img

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	apiEndpoint  string
	accessToken  string
	refreshToken string
	clientID     string
)

// Img --> imageResponse content
type Img struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

// response when create add new image to Imgur gallery
type imageResponse struct {
	Img *Img `json:"data"`
}

// SetEnvVar set environment var for interacting with Imgur API
func SetEnvVar(cfg *conf.Config) {
	apiEndpoint = cfg.Imgur.Endpoint
	accessToken = cfg.Imgur.AccessToken
	refreshToken = cfg.Imgur.RefreshToken
	clientID = cfg.Imgur.ClientID
}

// UploadImagetoImgur add image to Imgur gallery and return link to new image
func UploadImagetoImgur(title string, imageURL string) (*Img, error) {

	response := &imageResponse{}
	url := apiEndpoint + "image"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	writer.WriteField("image", imageURL)
	writer.WriteField("type", "url")
	writer.WriteField("title", title)

	err := writer.Close()
	if err != nil {
		util.Danger(err)
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		util.Danger(err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		util.Danger(err)
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(response)

	return response.Img, err
}

// // UpdateImage update imgur image
// func UpdateImage(imageID, imageURL string) (img *Img, err error) {

// 	img, err = GetImageFromImgur(imageID)
// 	if err != nil {
// 		util.Danger(err)
// 		return
// 	}

// 	if strings.Compare(img.Description, imageURL) != 0 {
// 		err = DeleteImg(imageID)
// 		if err != nil {
// 			util.Danger(err)
// 			return
// 		}

// 		img, err = UploadImagetoImgur(img.Title, imageURL)
// 		if err != nil {
// 			util.Danger(err)
// 			return
// 		}
// 	}

// 	return
// }

// GetImageFromImgur get img using image ID from Imgur
// func GetImageFromImgur(imageID string) (*Img, error) {

// 	response := &imageResponse{}
// 	url := apiEndpoint + "image/" + imageID

// 	client := &http.Client{}
// 	req, err := http.NewRequest("GET", url, nil)

// 	if err != nil {
// 		util.Danger(err)
// 		return nil, err
// 	}

// 	req.Header.Add("Authorization", "Client-ID "+clientID)
// 	res, err := client.Do(req)
// 	if err != nil {
// 		util.Danger(err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	decoder := json.NewDecoder(res.Body)
// 	err = decoder.Decode(response)

// 	return response.Img, err
// }

// DeleteImg delete img in imgur
func DeleteImg(imageID string) {

	url := apiEndpoint + "image/" + imageID

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		util.Danger(err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	_, err = client.Do(req)
	if err != nil {
		util.Danger(err)
	}

	return
}
