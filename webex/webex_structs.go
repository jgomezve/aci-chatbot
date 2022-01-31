package webex

// People URI
type WebexPeopleReply struct {
	People []WebexPeople `json:"items"`
}

type WebexPeople struct {
	Id          string   `json:"id"`
	Emails      []string `json:"emails"`
	DisplayName string   `json:"displayName"`
	NickName    string   `json:"nickName"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	UserName    string   `json:"userName"`
}

// Room URI
type WebexRoomsReply struct {
	Rooms []WebexRoom `json:"items"`
}

type WebexRoom struct {
	Id           string `json:"id"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	IsLocked     bool   `json:"isLocked"`
	LastActivity string `json:"lastActivity"`
	CreatorId    string `json:"creatorId"`
	Created      string `json:"created"`
	OwnerId      string `json:"ownerId"`
}

// Message URI
type WebexMessagesReply struct {
	Messages []WebexMessage `json:"items"`
}

type WebexMessage struct {
	Id          string `json:"id,omitempty"`
	RoomId      string `json:"roomId,omitempty"`
	RoomType    string `json:"roomType,omitempty"`
	Text        string `json:"text,omitempty"`
	PersonId    string `json:"personId,omitempty"`
	PersonEmail string `json:"personEmail,omitempty"`
	Created     string `json:"created,omitempty"`
	Markdown    string `json:"markdown,omitempty"`
}

// Webhook URI
type WebexWebhookReply struct {
	Webhooks []WebexWebhook `json:"items"`
}

type WebexWebhook struct {
	Id        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	TargetUrl string            `json:"targetUrl,omitempty"`
	Resource  string            `json:"resource,omitempty"`
	Event     string            `json:"event,omitempty"`
	OrgId     string            `json:"orgId,omitempty"`
	CreatedBy string            `json:"createdBy,omitempty"`
	AppId     string            `json:"appId,omitempty"`
	OwnerId   string            `json:"OwnerId,omitempty"`
	Status    string            `json:"status,omitempty"`
	Created   string            `json:"created,omitempty"`
	ActorId   string            `json:"actorId,omitempty"`
	Data      *WebexWebhookData `json:"data,omitempty"`
}

type WebexWebhookData struct {
	Id              string   `json:"id,omitempty"`
	RoomId          string   `json:"roomId,omitempty"`
	RoomType        string   `json:"roomType,omitempty"`
	PersonId        string   `json:"personId,omitempty"`
	PersonEmail     string   `json:"personEmail,omitempty"`
	MentionedPeople []string `json:"mentionedPeople"`
	Created         string   `json:"created,omitempty"`
}
