package crawler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

type helper interface {
	detectSpoiler(chapURL string, attr1, attr2 string) error
	getPageSource(pageURL string) (doc *goquery.Document, err error)
}

type comicHelper struct{}

func (ch comicHelper) detectSpoiler(chapURL string, attr1, attr2 string) error {

	// Check if chapter is full upload (detect spolier chap)
	doc, err := ch.getPageSource(chapURL)
	if err != nil {
		logging.Danger()
		return err
	}

	if chapSelections := doc.Find(attr1).Find(attr2); chapSelections.Size() < 3 {
		logging.Danger()
		return errors.New("No new chapter, just some spoilers :)")

	}

	return nil
}

func (ch comicHelper) getPageSource(pageURL string) (doc *goquery.Document, err error) {

	c := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := c.Get(pageURL)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Danger()
		return
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		logging.Danger()
		return
	}

	return
}
