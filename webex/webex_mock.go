//go:build !webex
// +build !webex

package webex

import (
	"log"
)

type WebexClientMocks struct {
	LastMsgSent           string
	CreateWebhookF        func(name, url, resource, event string) error
	GetBotDetailsF        func() (WebexPeople, error)
	GetWebHooksF          func() ([]WebexWebhook, error)
	DeleteWebhookF        func(id string) error
	SendMessageToRoomF    func(m string, roomId string) error
	GetPersonInformationF func(id string) (WebexPeople, error)
	GetMessageByIdF       func(id string) (WebexMessage, error)
	GetRoomByIdF          func(roomId string) (WebexRoom, error)
}

var (
	WebexMockClient WebexClientMocks
)

// Mock functions default values
func (wbx *WebexClientMocks) SetDefaultFunctions() {
	wbx.LastMsgSent = ""
	wbx.GetBotDetailsF = func() (WebexPeople, error) {
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

	wbx.GetWebHooksF = func() ([]WebexWebhook, error) {
		log.Println("Mock: Getting Webhook")
		return []WebexWebhook{{
			Name: "Test",
		}}, nil
	}

	wbx.CreateWebhookF = func(name, url, resource, event string) error {
		log.Println("Mock: Creating Webhook")
		return nil
	}

	wbx.DeleteWebhookF = func(id string) error {
		log.Println("Mock: Deleting Webhook")
		return nil
	}

	wbx.SendMessageToRoomF = func(m, roomId string) error {
		log.Printf("Mock: Sending Message to Webex Room %s\n%s\n", roomId, m)
		wbx.LastMsgSent = m
		return nil
	}

	wbx.GetPersonInformationF = func(id string) (WebexPeople, error) {
		return WebexPeople{DisplayName: "ARandomPerson"}, nil
	}

	wbx.GetMessageByIdF = func(id string) (WebexMessage, error) {
		return WebexMessage{Text: "This is a mocked webex test", PersonId: "ARandomId"}, nil
	}
}

func (wbx *WebexClientMocks) GetBotDetails() (WebexPeople, error) {
	return wbx.GetBotDetailsF()
}

func (wbx *WebexClientMocks) SendMessageToRoom(m string, roomId string) error {
	return wbx.SendMessageToRoomF(m, roomId)
}

func (wbx *WebexClientMocks) DeleteWebhook(id string) error {
	return wbx.DeleteWebhookF(id)
}

func (wbx *WebexClientMocks) CreateWebhook(name, url, resource, event string) error {
	return wbx.CreateWebhookF(name, url, resource, event)
}

func (wbx *WebexClientMocks) GetWebHooks() ([]WebexWebhook, error) {
	return wbx.GetWebHooksF()
}

func (wbx *WebexClientMocks) GetPersonInformation(id string) (WebexPeople, error) {
	return wbx.GetPersonInformationF(id)
}

func (wbx *WebexClientMocks) GetMessageById(id string) (WebexMessage, error) {
	return wbx.GetMessageByIdF(id)
}

func (wbx *WebexClientMocks) GetRoomById(roomId string) (WebexRoom, error) {
	return wbx.GetRoomByIdF(roomId)
}
