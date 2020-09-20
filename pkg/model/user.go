package model

// User model
type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	AppID      string `json:"appid"`
	PageID     string `json:"pageid"`
	ProfilePic string `json:"profile-pic"`
}

// UserList contains multiple users
type UserList struct {
	Users []User `json:"users"`
}

// Session model
type Session struct {
	ID    int    `json:"id"`
	UUID  string `json:"uuid"`
	AppID string `json:"appid"`
}
