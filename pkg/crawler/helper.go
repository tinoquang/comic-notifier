package crawler

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type crawlHelper struct{}

func (ch crawlHelper) detectSpoiler(name, chapURL, chapterName, attr1, attr2 string) error {

	if strings.Contains(chapterName, "leak") || strings.Contains(chapterName, "spoil") {
		return errors.Errorf("%s has spoiler chapter", name)
	}

	// Check if chapter is full upload (detect spolier chap)
	doc, err := ch.getPageSource(chapURL)
	if err != nil {
		return err
	}

	if chapSelections := doc.Find(attr1).Find(attr2); chapSelections.Size() < 3 {
		return errors.Errorf("%s has spoiler chapter", name)

	}

	return nil
}

func (ch crawlHelper) getPageSource(pageURL string) (doc *goquery.Document, err error) {

	pageBody, err := util.MakeGetRequest(pageURL, nil)
	if err != nil {
		return
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(pageBody))
	if err != nil {
		return
	}
	return
}
