package model

// Comic model
type Comic struct {
	ID         int    `json:"id"`
	Page       string `json:"page"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	ImageURL   string `json:"-"`
	LatestChap string `json:"latest"`
	ChapURL    string `json:"chap-url"`
	Date       string `json:"-"`
	DateFormat string `json:"-"`
}
