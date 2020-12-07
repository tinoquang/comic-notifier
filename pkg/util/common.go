package util

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	scraper "github.com/tinoquang/go-cloudflare-scraper"
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
func DownloadFile(url string, fileName string) (err error) {

	c, err := scraper.NewClient()
	if err != nil {
		return
	}
	res, err := c.Get(url)
	if err != nil {
		return
	}

	defer res.Body.Close()
	if err != nil {
		return
	}

	//open a file for writing
	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return
	}
	return
}
