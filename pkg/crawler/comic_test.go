package crawler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
)

type comicData struct {
	URL      string
	testData string
}

type mockHelper struct {
	testData string
}

var detectSpoilerMock func() error

func (m mockHelper) detectSpoiler(name, chapURL, attr1, attr2 string) error {

	return detectSpoilerMock()
}

func (m mockHelper) getPageSource(pageURL string) (doc *goquery.Document, err error) {

	f, err := os.Open(m.testData)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, err
	}

	return
}

func TestInvalidURL(t *testing.T) {

	mockBeeng := mockHelper{
		testData: "",
	}

	conf.Init()
	c := newComicCrawler(mockBeeng)

	_, err := c.GetComicInfo(context.Background(), "https://beeng")
	assert.EqualError(t, err, "Page is not supported yet")
}

func TestCrawlComic(t *testing.T) {

	conf.Init()
	comicTests := []comicData{
		{
			URL:      "https://beeng.net/dao-hai-tac-31953.html",
			testData: "./test_data/beeng_daohaitac.html",
		},
		{
			URL:      "https://blogtruyen.vn/139/one-piece",
			testData: "./test_data/blogtruyen_onepiece.html",
		},
		{
			URL:      "https://truyentranh.net/one-piece",
			testData: "./test_data/truyentranhnet_onepiece.html",
		},
		{
			URL:      "http://truyentranhtuan.com/one-piece/",
			testData: "./test_data/truyentranhtuan_onepiece.html",
		},
		{
			URL:      "http://truyenqq.com/truyen-tranh/dao-hai-tac-128",
			testData: "./test_data/truyenqq_daohaitac.html",
		},
		{
			URL:      "https://hocvientruyentranh.net/truyen/67/one-piece",
			testData: "./test_data/hocvientruyentranh_onepiece.html",
		},
	}

	// Testing with dectectSpolier return true
	detectSpoilerMock = func() error {
		return nil
	}

	for _, comic := range comicTests {
		mockBeeng := mockHelper{
			testData: comic.testData,
		}

		c := newComicCrawler(mockBeeng)

		_, err := c.GetComicInfo(context.Background(), comic.URL)

		assert.Nil(t, err)
	}

}

func TestDetectSpolierFailed(t *testing.T) {

	conf.Init()
	comicTests := []comicData{
		{
			URL:      "https://beeng.net/dao-hai-tac-31953.html",
			testData: "./test_data/beeng_daohaitac.html",
		},
		{
			URL:      "https://blogtruyen.vn/139/one-piece",
			testData: "./test_data/blogtruyen_onepiece.html",
		},
		{
			URL:      "https://truyentranh.net/one-piece",
			testData: "./test_data/truyentranhnet_onepiece.html",
		},
		{
			URL:      "http://truyenqq.com/truyen-tranh/dao-hai-tac-128",
			testData: "./test_data/truyenqq_daohaitac.html",
		},
		{
			URL:      "https://hocvientruyentranh.net/truyen/67/one-piece",
			testData: "./test_data/hocvientruyentranh_onepiece.html",
		},
	}

	// Testing with dectectSpolier return true
	detectSpoilerMock = func() error {
		return errors.Errorf("Check spoiler failed")
	}

	for _, comic := range comicTests {
		mockBeeng := mockHelper{
			testData: comic.testData,
		}

		c := newComicCrawler(mockBeeng)

		_, err := c.GetComicInfo(context.Background(), comic.URL)

		assert.EqualError(t, err, "Check spoiler failed")
	}

}

func TestVerifycomic(t *testing.T) {

	comic := db.Comic{}

	assert.Contains(t, verifyComic(&comic).Error(), "Comic name is missing")

	comic.Name = "name"
	assert.Contains(t, verifyComic(&comic).Error(), "Comic chapURL is missing")

	comic.ChapUrl = "chapUrl"
	assert.Contains(t, verifyComic(&comic).Error(), "Comic ImgUrl is missing")

	comic.ImgUrl = "imgUrl"
	assert.Contains(t, verifyComic(&comic).Error(), "Comic cloudImgUrl is missing")

	comic.CloudImgUrl = "CloudImgUrl"
	assert.Contains(t, verifyComic(&comic).Error(), "Comic latestchap is missing")

	comic.LatestChap = "latestChap"

	comic.Page = "hocvientruyentranh.net"
	assert.Nil(t, verifyComic(&comic))

	comic.Page = "beeng.net"
	assert.Contains(t, verifyComic(&comic).Error(), "Comic date is missing")

	comic.LastUpdate = time.Now()
	assert.Nil(t, verifyComic(&comic))

}
