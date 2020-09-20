package model

// Subscriber model
type Subscriber struct {
	ID      int `json:"id"`
	UserID  int `json:"userid"`
	ComicID int `json:"comicid"`
}

// SubscriberList contains multiple subscribers
type SubscriberList struct {
	Subscribers []Subscriber `json:"subscribers"`
}
