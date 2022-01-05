package webex

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

type WebexRoomsReply struct {
	Rooms []WebexRoom `json:"items"`
}

type WebexMessagesReply struct {
	Messages []WebexMessageR `json:"items"`
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

type WebexMessageR struct {
	Id          string `json:"id"`
	RoomId      string `json:"roomId"`
	RoomType    string `json:"roomType"`
	Text        string `json:"text"`
	PersonId    string `json:"personId"`
	PersonEmail string `json:"personEmail"`
	Created     string `json:"created"`
}

type WebexMessage struct {
	RoomId   string `json:"roomId"`
	Markdown string `json:"markdown"`
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
	Id          string `json:"id,omitempty"`
	RoomId      string `json:"roomId,omitempty"`
	RoomType    string `json:"roomType,omitempty"`
	PersonId    string `json:"personId,omitempty"`
	PersonEmail string `json:"personEmail,omitempty"`
	Created     string `json:"created,omitempty"`
}
