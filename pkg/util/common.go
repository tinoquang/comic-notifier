package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
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

	resp, err := http.Get(reqURL.String())
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
