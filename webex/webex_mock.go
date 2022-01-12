package webex

import (
	"fmt"
	"log"
)

type WebexClientMocks struct {
	tkn     string
	baseURL string
}

var (
	WebexMockClient WebexClientMocks
	GetBotDetailsF  func() (WebexPeople, error)
	GetWebHooksF    func() ([]WebexWebhook, error)
	CreateWebhookF  func(name, url, resource, event string) error
	DeleteWebhookF  func(name, tUrl, id string) error
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
}

func (wbx *WebexClientMocks) GetBotDetails() (WebexPeople, error) {
	return GetBotDetailsF()
}

func (wbx *WebexClientMocks) SendMessageToRoom(m string, roomId string) error {
	fmt.Printf("Sending message")
	return nil
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

func (wbx *WebexClientMocks) GetPersonInfromation(id string) (WebexPeople, error) {
	return WebexPeople{DisplayName: "Test"}, nil
}

func (wbx *WebexClientMocks) GetMessages(roomId string, max int) ([]WebexMessage, error) {
	return []WebexMessage{{Text: "Test", PersonId: "Test"}}, nil
}
