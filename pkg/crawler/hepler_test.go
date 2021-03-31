package crawler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestGetPageSourceSuccess(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := "test.vn"
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			<!DOCTYPE html>
			<html lang="en">
			  <head>
				<meta charset="UTF-8" />
				<meta http-equiv="X-UA-Compatible" content="IE=edge" />
				<meta name="viewport" content="width=device-width, initial-scale=1.0" />
				<title>Document</title>
			  </head>
			  <body>
				<p>Mock page</p>
			  </body>
			</html>`,
			)
			return resp, nil
		},
	)

	h := crawlHelper{}

	_, err := h.getPageSource(reqURL)
	require.Nil(t, err)
}

func TestGetRequestFailed(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := "test.vn"
	httpmock.RegisterResponder("GET", reqURL, httpmock.NewErrorResponder(errors.New("Make get request failed")))

	h := crawlHelper{}

	_, err := h.getPageSource(reqURL)
	require.Contains(t, err.Error(), "Make get request failed")
}

func TestNoSpoiler(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := "test.vn"
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			<!DOCTYPE html>
			<html lang="en">
			  <head>
				<meta charset="UTF-8" />
				<meta http-equiv="X-UA-Compatible" content="IE=edge" />
				<meta name="viewport" content="width=device-width, initial-scale=1.0" />
				<title>Document</title>
			  </head>
			  <body>
			  <div class="story-see-content">
			  <img class="lazy" src="http://tintruyen.net/499/fix-286/0.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/1.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/2.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/3.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/4.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/5.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<br>                </div>
			  </body>
			</html>`,
			)
			return resp, nil
		},
	)

	h := crawlHelper{}

	err := h.detectSpoiler("name", "test.vn", ".story-see-content", "img")
	require.Nil(t, err)
}

func TestGetRequestFailedWhenDetectSpolier(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := "test.vn"
	httpmock.RegisterResponder("GET", reqURL, httpmock.NewErrorResponder(errors.New("Make get request failed")))

	h := crawlHelper{}

	err := h.detectSpoiler("name", "test.vn", ".story-see-content", "img")
	require.Contains(t, err.Error(), "Make get request failed")
}

func TestDetectSpoiler(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := "test.vn"
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			<!DOCTYPE html>
			<html lang="en">
			  <head>
				<meta charset="UTF-8" />
				<meta http-equiv="X-UA-Compatible" content="IE=edge" />
				<meta name="viewport" content="width=device-width, initial-scale=1.0" />
				<title>Document</title>
			  </head>
			  <body>
			  <div class="story-see-content">
			  <img class="lazy" src="http://tintruyen.net/499/fix-286/0.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<img class="lazy" src="http://tintruyen.net/499/fix-286/1.jpg?d=dfgd6546" alt="Black Clover Chap 286 - Next Chap 287"><br>
<br>                </div>
			  </body>
			</html>`,
			)
			return resp, nil
		},
	)

	h := crawlHelper{}

	err := h.detectSpoiler("name", "test.vn", ".story-see-content", "img")
	require.Contains(t, err.Error(), "has spoiler chapter")
}
