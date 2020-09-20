package msg

/* --------- Facebook message request format -------- */

// UserMessage general message form
type UserMessage struct {
	Object  string  `json:"object,omitempty"`
	Entries []Entry `json:"entry,omitempty"`
}

// Entry message from each page
// meaning each POST message can contains more than one message from each page
type Entry struct {
	ID        string      `json:"id,omitempty"` // Page ID
	Time      int         `json:"time,omitempty"`
	Messaging []Messaging `json:"messaging,omitempty"`
}

// Messaging struct per message
type Messaging struct {
	Sender    *User         `json:"sender,omitempty"`    // Page-scope ID different with Page ID
	Recipient *User         `json:"recipient,omitempty"` // Page ID
	Timestamp int           `json:"timestamp,omitempty"`
	Message   *RecvMsg      `json:"message,omitempty"`
	PostBack  *RecvPostBack `json:"postback,omitempty"`
}

// RecvMsg --> message received from user
type RecvMsg struct {
	Mid        string `json:"mid,omitempty"`
	Text       string `json:"text,omitempty"`
	QuickReply *Reply `json:"quick_reply,omitempty"`
}

// Reply : user send msg with type Quickreply (normally when press button)
type Reply struct {
	Payload string `json:"payload,omitempty"`
}

// RecvPostBack --> message when user press button
type RecvPostBack struct {
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// User struct contains ID of Sender and Recipient
type User struct {
	ID string `json:"id,omitempty"`
}

/* -------------Message response format----------- */

// Response : general response format
type Response struct {
	Type          string   `json:"messaging_type,omitempty"`
	Recipient     *User    `json:"recipient,omitempty"`
	Message       *RespMsg `json:"message,omitempty"`
	Action        string   `json:"sender_action,omitempty"`
	MessagingType string   `json:"messaging_type,omitempty"`
	Tag           string   `json:"tag,omitempty"`
}

// RespMsg : message content, include some type: text, template, quick-reply,..
type RespMsg struct {
	Text     string       `json:"text,omitempty"`
	Template *Attachment  `json:"attachment,omitempty"`
	Options  []QuickReply `json:"quick_replies,omitempty"`
}

// Attachment such as image, link
type Attachment struct {
	Type     string   `json:"type,omitempty"`
	Payloads *Payload `json:"payload,omitemtpy"`
}

// Payload : attachment content, usually image, button, ...
type Payload struct {
	TemplateType string    `json:"template_type,omitempty"`
	Elements     []Element `json:"elements,omitempty"`
}

// Element : template elements
type Element struct {
	Title         string   `json:"title,omitempty"`
	ImageURL      string   `json:"image_url,omitempty"`
	Subtitle      string   `json:"subtitle,omitempty"`
	DefaultAction *Action  `json:"default_action,omitempty"`
	Buttons       []Button `json:"buttons,omitempty"`
}

// Action : contains URL of comic
type Action struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Button : button include in attachment, example: Read button, Unsubscribe button,...
type Button struct {
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// QuickReply : button, link,... generate quick-reply request when user click
type QuickReply struct {
	Type     string `json:"content_type,omitempty"`
	Title    string `json:"title,omitempty"`
	Payload  string `json:"payload,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}
