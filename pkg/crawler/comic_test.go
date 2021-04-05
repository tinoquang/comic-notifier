package crawler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
)

type comicData struct {
	URL      string
	testData string
}

type mockHelper struct {
	testData          string
	detectSpoilerMock func() error
	getPageSourceMock func(testData string) (*goquery.Document, error)
}

func (m mockHelper) detectSpoiler(name, chapURL, attr1, attr2 string) error {

	return m.detectSpoilerMock()
}

func (m mockHelper) getPageSource(pageURL string) (doc *goquery.Document, err error) {

	return m.getPageSourceMock(m.testData)
}

func readTestFile(path string) (*goquery.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, err
	}

	return doc, err
}

func TestInvalidURL(t *testing.T) {

	mockBeeng := mockHelper{
		testData:          "",
		getPageSourceMock: readTestFile,
	}

	conf.Init()
	c := newComicCrawler(mockBeeng)

	_, err := c.GetComicInfo(context.Background(), "https://beeng")
	assert.EqualError(t, err, "Page is not supported yet")
}

func TestGetPageSourceTimeout(t *testing.T) {

	mockBeeng := mockHelper{
		testData: "",
		getPageSourceMock: func(testData string) (*goquery.Document, error) {
			return nil, errors.Errorf("Timeout")
		},
	}

	conf.Init()
	c := newComicCrawler(mockBeeng)

	_, err := c.GetComicInfo(context.Background(), "https://beeng.net")
	assert.EqualError(t, err, "Time out when crawl comic")

}

func TestGetPageSourceFailed(t *testing.T) {

	mockBeeng := mockHelper{
		testData: "",
		getPageSourceMock: func(testData string) (*goquery.Document, error) {
			return nil, errors.Errorf("Failed")
		},
	}

	conf.Init()
	c := newComicCrawler(mockBeeng)

	_, err := c.GetComicInfo(context.Background(), "https://beeng.net")
	assert.EqualError(t, err, "Crawl failed")

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

	want := []db.Comic{
		{
			Page:        "beeng.net",
			Name:        "Đảo Hải Tặc",
			Url:         "https://beeng.net/dao-hai-tac-31953.html",
			ImgUrl:      "https://cdn2.beeng.net/mangas/2020/07/26/05/dao-hai-tac.jpg",
			CloudImgUrl: "https://storage.googleapis.com/comicnotifier-cde7d.appspot.com/beeng.net/Đảo Hải Tặc",
			LatestChap:  "Chapter 1008",
			ChapUrl:     "https://beeng.net/dao-hai-tac-31953/chapter-1008-959587.html",
			LastUpdate:  time.Date(2021, 3, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			Page:        "blogtruyen.vn",
			Name:        "One Piece",
			Url:         "https://blogtruyen.vn/139/one-piece",
			ImgUrl:      "https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg",
			CloudImgUrl: "https://storage.googleapis.com/comicnotifier-cde7d.appspot.com/blogtruyen.vn/One Piece",
			LatestChap:  "One Piece Chapter 1008",
			ChapUrl:     "https://blogtruyen.vn/c562868/one-piece-chapter-1008",
			LastUpdate:  time.Date(2021, 3, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			Page:        "truyentranhtuan.com",
			Name:        "One Piece",
			Url:         "http://truyentranhtuan.com/one-piece/",
			ImgUrl:      "http://truyentranhtuan.com/wp-content/uploads/2013/01/one-piece-anh-bia-200x304.jpg",
			CloudImgUrl: "https://storage.googleapis.com/comicnotifier-cde7d.appspot.com/truyentranhtuan.com/One Piece",
			LatestChap:  "One Piece 1008",
			ChapUrl:     "http://truyentranhtuan.com/one-piece-chuong-1008/",
			LastUpdate:  time.Date(2021, 3, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			Page:        "truyenqq.com",
			Name:        "Đảo Hải Tặc",
			Url:         "http://truyenqq.com/truyen-tranh/dao-hai-tac-128",
			ImgUrl:      "http://i.mangaqq.com/ebook/190x247/dao-hai-tac_1552224567.jpg?r=r8645456",
			CloudImgUrl: "https://storage.googleapis.com/comicnotifier-cde7d.appspot.com/truyenqq.com/Đảo Hải Tặc",
			LatestChap:  "Chương 1008",
			ChapUrl:     "http://truyenqq.com/truyen-tranh/dao-hai-tac-128-chap-1008.html",
			LastUpdate:  time.Date(2021, 3, 23, 0, 0, 0, 0, time.UTC),
		},
		{
			Page:        "hocvientruyentranh.net",
			Name:        "One Piece",
			Url:         "https://hocvientruyentranh.net/truyen/67/one-piece",
			ImgUrl:      "https://i.imgur.com/62yFIVR.png",
			CloudImgUrl: "https://storage.googleapis.com/comicnotifier-cde7d.appspot.com/hocvientruyentranh.net/One Piece",
			LatestChap:  "Chapter 1008",
			ChapUrl:     "https://hocvientruyentranh.net/chapter/254911/one-piece-chapter-1008",
			LastUpdate:  time.Time{},
		},
	}
	for i, comic := range comicTests {
		mockBeeng := mockHelper{
			testData: comic.testData,
			detectSpoilerMock: func() error {
				return nil
			},
			getPageSourceMock: readTestFile,
		}

		crawler := newComicCrawler(mockBeeng)

		c, err := crawler.GetComicInfo(context.Background(), comic.URL)

		require.Nil(t, err)
		require.Equal(t, c, want[i])
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
			URL:      "http://truyenqq.com/truyen-tranh/dao-hai-tac-128",
			testData: "./test_data/truyenqq_daohaitac.html",
		},
		{
			URL:      "https://hocvientruyentranh.net/truyen/67/one-piece",
			testData: "./test_data/hocvientruyentranh_onepiece.html",
		},
	}

	for _, comic := range comicTests {
		mockBeeng := mockHelper{
			testData: comic.testData,
			detectSpoilerMock: func() error {
				return errors.Errorf("Check spoiler failed")
			},
			getPageSourceMock: readTestFile,
		}

		c := newComicCrawler(mockBeeng)

		_, err := c.GetComicInfo(context.Background(), comic.URL)

		require.EqualError(t, err, "Check spoiler failed")
	}

}

func TestVerifycomic(t *testing.T) {

	comic := db.Comic{}

	require.Contains(t, verifyComic(&comic).Error(), "Comic name is missing")

	comic.Name = "name"
	require.Contains(t, verifyComic(&comic).Error(), "Comic chapURL is missing")

	comic.ChapUrl = "chapUrl"
	require.Contains(t, verifyComic(&comic).Error(), "Comic ImgUrl is missing")

	comic.ImgUrl = "imgUrl"
	require.Contains(t, verifyComic(&comic).Error(), "Comic cloudImgUrl is missing")

	comic.CloudImgUrl = "CloudImgUrl"
	require.Contains(t, verifyComic(&comic).Error(), "Comic latestchap is missing")

	comic.LatestChap = "latestChap"

	comic.Page = "hocvientruyentranh.net"
	require.Nil(t, verifyComic(&comic))

	comic.Page = "beeng.net"
	require.Contains(t, verifyComic(&comic).Error(), "Comic date is missing")

	comic.LastUpdate = time.Now()
	require.Nil(t, verifyComic(&comic))

}
