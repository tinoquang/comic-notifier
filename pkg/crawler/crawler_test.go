package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"github.com/tinoquang/comic-notifier/pkg/conf"
)

func TestGetUserInfoWithPSID(t *testing.T) {

	conf.Init()
	psid := "123"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, psid)
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			{
				"name": "get_info_test",
				"ids_for_apps": {
					"data": [{
						"id": "123456"
					}]
				},
				"picture":{
					"data":{
					   "url":"picture_url"
					}
				}
			}`,
			)

			resp.Header.Add("Content-Type", "application/json")
			return resp, nil
		},
	)

	crawler := NewCrawler()

	u, err := crawler.GetUserInfoFromFacebook("psid", psid)

	require.Nil(t, err)

	require.Equal(t, u.Name, "get_info_test")
	require.Equal(t, u.Appid.String, "123456")
	require.Equal(t, u.ProfilePic.String, "picture_url")
}

func TestGetUserInfoWithAppID(t *testing.T) {

	conf.Init()
	appID := "123"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, appID)
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			{
				"name": "get_info_test",
				"ids_for_pages": {
					"data": [{
						"id": "123456"
					}]
				},
				"picture":{
					"data":{
					   "url":"picture_url"
					}
				}
			}`,
			)

			resp.Header.Add("Content-Type", "application/json")
			return resp, nil
		},
	)

	crawler := NewCrawler()

	u, err := crawler.GetUserInfoFromFacebook("appid", appID)

	require.Nil(t, err)

	require.Equal(t, u.Name, "get_info_test")
	require.Equal(t, u.Psid.String, "123456")
	require.Equal(t, u.ProfilePic.String, "picture_url")
}

func TestSendRequestFailed(t *testing.T) {

	conf.Init()
	psid := "123"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, psid)
	httpmock.RegisterResponder("GET", reqURL, httpmock.NewErrorResponder(errors.New("Send request failed")))

	crawler := NewCrawler()

	_, err := crawler.GetUserInfoFromFacebook("appid", psid)

	require.Contains(t, err.Error(), "Send request failed")
}

func TestParsingResponseFailed(t *testing.T) {

	conf.Init()
	psid := "123"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, psid)
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
			{
				"name": "get_info_test",
				"ids_for_apps": {
					"data": [{
						"id": "123456",
					}]
				},
				"picture":{
					"data":{
					   "url":"picture_url"
					}
				}
			}`,
			)

			resp.Header.Add("Content-Type", "application/json")
			return resp, nil
		},
	)

	crawler := NewCrawler()

	_, err := crawler.GetUserInfoFromFacebook("psid", psid)

	require.Contains(t, err.Error(), "invalid characte")
}

func TestInvalidField(t *testing.T) {
	conf.Init()
	crawler := NewCrawler()
	psid := "123"
	_, err := crawler.GetUserInfoFromFacebook("wrong field", psid)

	require.EqualError(t, err, "Wrong field request, field: wrong field")
}
