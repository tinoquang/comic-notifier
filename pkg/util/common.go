package util

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/valyala/fasthttp"
)

// MakeGetRequest send HTTP GET request with mapped queries
func MakeGetRequest(URL string, queries map[string]string) (respBody []byte, err error) {

	// Create url string with query parameters
	reqURL, err := url.Parse(URL)
	if err != nil {
		return
	}

	q := reqURL.Query()
	for key, val := range queries {
		q.Add(key, val)
	}
	reqURL.RawQuery = q.Encode()

	// Make request using fast http
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(reqURL.String())
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.Danger("Client get failed: %s\n", err)
		return
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		logging.Danger("Expected status code %d but got %d\n", fasthttp.StatusOK, resp.StatusCode())
		return
	}

	respBody = resp.Body()
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
