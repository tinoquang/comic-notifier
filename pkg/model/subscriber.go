package model

// Subscriber model
type Subscriber struct {
	ID        int    `json:"id"`
	Page      string `json:"page"`
	UserID    int    `json:"userid"`
	UserName  string `json:"username"`
	ComicID   int    `json:"comicid"`
	ComicName string `json:"comicname"`
}

// SubscriberList contains multiple subscribers
type SubscriberList struct {
	Subscribers []Subscriber `json:"subscribers"`
}
