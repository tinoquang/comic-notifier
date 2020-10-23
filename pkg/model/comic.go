package model

import (
	"errors"
)

// Comic model
type Comic struct {
	ID         int        `json:"id"`
	Page       string     `json:"page"`
	Name       string     `json:"name"`
	URL        string     `json:"url"`
	ImageURL   string     `json:"-"`
	ImgurID    NullString `json:"-"`
	ImgurLink  NullString `json:"-"`
	LatestChap string     `json:"latest"`
	ChapURL    string     `json:"chap-url"`
	Date       string     `json:"-"`
	DateFormat string     `json:"-"`
}

// NullString represent empty string for database
type NullString string

// Scan method of Nullstring
func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return errors.New("Column is not a string")
	}
	*s = NullString(strVal)
	return nil
}

// Value return string
func (s NullString) Value() string {
	if len(s) == 0 { // if nil or empty string
		return ""
	}
	return string(s)
}
