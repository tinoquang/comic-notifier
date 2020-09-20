package model

// Page model
type Page struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

// PageList contains multiple pages
type PageList struct {
	Page []Page `json:"pages"`
}
