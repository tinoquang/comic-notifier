package model

// Subscriber model
type Subscriber struct {
	ID      int    `json:"id"`
	PSID    string `json:"user_psid"`
	ComicID int    `json:"comicid"`
}

// SubscriberList contains multiple subscribers
type SubscriberList struct {
	Subscribers []Subscriber `json:"subscribers"`
}
