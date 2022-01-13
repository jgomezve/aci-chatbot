package webex

import (
	"log"
)

type WebexClientMocks struct {
	tkn     string
	baseURL string
}

// TODO: check if this approach is valid
var (
	WebexMockClient       WebexClientMocks
	GetBotDetailsF        func() (WebexPeople, error)
	GetWebHooksF          func() ([]WebexWebhook, error)
	CreateWebhookF        func(name, url, resource, event string) error
	DeleteWebhookF        func(name, tUrl, id string) error
	GetMessagesF          func(roomId string, max int) ([]WebexMessage, error)
	SendMessageToRoomF    func(m string, roomId string) error
	GetPersonInformationF func(id string) (WebexPeople, error)
)

// Mock functions default values
func (wbx *WebexClientMocks) SetDefaultFunctions() {
	GetBotDetailsF = func() (WebexPeople, error) {
		return WebexPeople{
			Id:          "ABC123",
			Emails:      []string{"test@bot.com"},
			DisplayName: "Test-Bot",
			NickName:    "Test",
			FirstName:   "Test",
			LastName:    "Bot",
			UserName:    "testbot",
		}, nil
	}

	GetWebHooksF = func() ([]WebexWebhook, error) {
		log.Println("Mock: Getting Webhook")
		return []WebexWebhook{{
			Name: "Test",
		}}, nil
	}

	CreateWebhookF = func(name, url, resource, event string) error {
		log.Println("Mock: Creating Webhook")
		return nil
	}

	DeleteWebhookF = func(name, tUrl, id string) error {
		log.Println("Mock: Deleting Webhook")
		return nil
	}

	GetMessagesF = func(roomId string, max int) ([]WebexMessage, error) {
		log.Println("Mock: Getting Webex Room Message")
		return []WebexMessage{{Text: "This is a mocked webex test", PersonId: "ARandomId"}}, nil
	}

	SendMessageToRoomF = func(m, roomId string) error {
		log.Println("Mock: Sending Message to Webex Room")
		return nil
	}

	GetPersonInformationF = func(id string) (WebexPeople, error) {
		return WebexPeople{DisplayName: "ARandomPerson"}, nil
	}
}

func (wbx *WebexClientMocks) GetBotDetails() (WebexPeople, error) {
	return GetBotDetailsF()
}

func (wbx *WebexClientMocks) SendMessageToRoom(m string, roomId string) error {
	return SendMessageToRoomF(m, roomId)
}

func (wbx *WebexClientMocks) DeleteWebhook(name, tUrl, id string) error {
	return DeleteWebhookF(name, tUrl, id)
}

func (wbx *WebexClientMocks) CreateWebhook(name, url, resource, event string) error {
	return CreateWebhookF(name, url, resource, event)
}

func (wbx *WebexClientMocks) GetWebHooks() ([]WebexWebhook, error) {
	return GetWebHooksF()
}

func (wbx *WebexClientMocks) GetPersonInformation(id string) (WebexPeople, error) {
	return GetPersonInformationF(id)
}

func (wbx *WebexClientMocks) GetMessages(roomId string, max int) ([]WebexMessage, error) {
	return GetMessagesF(roomId, max)
}
