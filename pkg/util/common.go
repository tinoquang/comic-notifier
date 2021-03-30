package util

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

// MakeGetRequest send HTTP GET request with mapped queries
func MakeGetRequest(URL string, queries map[string]string) (respBody []byte, err error) {

	reqURL, err := url.Parse(URL)
	if err != nil {
		return
	}

	q := reqURL.Query()
	for key, val := range queries {
		q.Add(key, val)
	}
	reqURL.RawQuery = q.Encode()

	c := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := c.Get(reqURL.String())
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	respBody, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	return

}

// DownloadFile simple function for downloading file bypass cloudfare
func DownloadFile(fileURL string, fileName string) (err error) {

	u, err := url.Parse(fileURL)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return
	}

	if u.Hostname() == "i.mangaqq.com" {
		req.Header.Set("Referer", "truyenqq.com")
	} else if strings.Contains(u.Hostname(), "hocvientruyentranh") {
		req.Header.Set("Referer", "https://hocvientruyentranh.net")
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		logging.Danger("error when download file", fileURL, "error code", resp.StatusCode)
		return ErrDownloadFile
	}
	defer resp.Body.Close()

	//open a file for writing
	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	return
}
