package crawler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tinoquang/comic-notifier/pkg/conf"
)

func TestGetUserInfoWithPSID(t *testing.T) {

	conf.Init()
	psid := "123"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqURL := fmt.Sprintf("%s/%s", conf.Cfg.Webhook.GraphEndpoint, psid)
	fmt.Println(reqURL)
	httpmock.RegisterResponder("GET", reqURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
	{
		"request_id": "abcdef0123456789abcdef0123456789",
		"status": "0"
	}
	`,
			)

			resp.Header.Add("Content-Type", "application/json")
			return resp, errors.Errorf("Error")
		},
	)

	crawler := NewCrawler()

	_, err := crawler.GetUserInfoFromFacebook("psid", psid)

	assert.Nil(t, err)
}
