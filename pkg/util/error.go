package util

import "github.com/pkg/errors"

var (
	ErrInvalidURL       = errors.New("Invalid URL")
	ErrComicUpToDate    = errors.Errorf("Comic is up-to-date, no new chapter")
	ErrPageNotSupported = errors.Errorf("Page is not supported yet")
)
