package crawler

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type crawlHelperWrapper struct{}

func (ch crawlHelperWrapper) detectSpoiler(name, chapURL, attr1, attr2 string) error {

	// Check if chapter is full upload (detect spolier chap)
	doc, err := ch.getPageSource(chapURL)
	if err != nil {
		logging.Danger()
		return err
	}

	if chapSelections := doc.Find(attr1).Find(attr2); chapSelections.Size() < 3 {
		return errors.Errorf("%s has spoiler chapter", name)

	}

	return nil
}

func (ch crawlHelperWrapper) getPageSource(pageURL string) (doc *goquery.Document, err error) {

	pageBody, err := util.MakeGetRequest(pageURL, map[string]string{})
	if err != nil {
		return
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(pageBody))
	if err != nil {
		logging.Danger()
		return
	}
	return
}
