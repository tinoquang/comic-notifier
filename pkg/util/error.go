package util

import "github.com/pkg/errors"

var (
	ErrImgUpToDate       = errors.New("Image is up-to-date")
	ErrAlreadySubscribed = errors.New("Already subscribed")
	ErrNotFound          = errors.New("Not found")
	ErrInvalidURL        = errors.New("Invalid URL")
	ErrCrawlTimeout      = errors.New("Time out when crawl comic")
	ErrDownloadFile      = errors.New("Cant' download file")
	ErrCrawlFailed       = errors.New("Crawl failed")
	ErrComicUpToDate     = errors.New("Comic is up-to-date, no new chapter")
	ErrPageNotSupported  = errors.New("Page is not supported yet")
)
