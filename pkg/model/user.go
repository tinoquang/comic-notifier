package model

// User model
type User struct {
	Name       string `json:"name"`
	PSID       string `json:"psid"`
	AppID      string `json:"appid"`
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
