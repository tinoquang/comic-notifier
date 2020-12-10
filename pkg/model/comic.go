package model

// Comic model
type Comic struct {
	ID           int    `json:"id"`
	Page         string `json:"page"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	OriginImgURL string `json:"-"`
	CloudImg     string `json:"-"`
	LatestChap   string `json:"latest"`
	ChapURL      string `json:"chap-url"`
}
