package img

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

var (
	apiEndpoint  string
	accessToken  string
	refreshToken string
	clientID     string
)

var (
	ErrUpToDate = errors.Errorf("Image is up-to-date")
)

type imgError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Img --> imageResponse content
type Img struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Link        string   `json:"link"`
	Description string   `json:"description"`
	Error       imgError `json:"error"`
}

// response when create add new image to Imgur gallery
type imageResponse struct {
	Img     *Img `json:"data"`
	Success bool `json:"success"`
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

	if imageURL == "" {
		return nil, errors.Errorf("Can't upload image to imgur: URL is empty")
	}

	response := &imageResponse{}
	url := apiEndpoint + "image"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	writer.WriteField("image", imageURL)
	writer.WriteField("type", "url")
	writer.WriteField("title", title)
	writer.WriteField("description", imageURL)

	err := writer.Close()
	if err != nil {
		logging.Danger(err)
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logging.Danger(err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		logging.Danger(err)
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(response)

	if response.Success != true {
		return nil, errors.New(response.Img.Error.Message)
	}
	return response.Img, err
}

// UpdateImage update imgur image
func UpdateImage(imageID string, comic *model.Comic) (err error) {

	if imageID == "" {
		return errors.New("Image ID is empty")
	}

	img, err := GetImageFromImgur(imageID)
	if err != nil {
		logging.Danger(err)
		return
	}

	img.Description = strings.Replace(img.Description, " . ", ".", -1)
	if strings.Compare(img.Description, comic.ImageURL) == 0 {
		return ErrUpToDate
	}

	img, err = UploadImagetoImgur(img.Title, comic.ImageURL)
	if err != nil {
		logging.Danger("Can't upload image to imgur, err :", err)
		return
	}

	DeleteImg(imageID)

	comic.ImgurID = model.NullString(img.ID)
	comic.ImgurLink = model.NullString(img.Link)
	return
}

// GetImageFromImgur get img using image ID from Imgur
func GetImageFromImgur(imageID string) (*Img, error) {

	response := &imageResponse{}
	url := apiEndpoint + "image/" + imageID

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		logging.Danger(err)
		return nil, err
	}

	req.Header.Add("Authorization", "Client-ID "+clientID)
	res, err := client.Do(req)
	if err != nil {
		logging.Danger(err)
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(response)

	return response.Img, err
}

// DeleteImg delete img in imgur
func DeleteImg(imageID string) error {

	if imageID == "" {
		return errors.New("Image ID is empty")
	}

	url := apiEndpoint + "image/" + imageID

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		logging.Danger(err)
		return err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	_, err = client.Do(req)
	if err != nil {
		logging.Danger(err)
	}

	return err
}
