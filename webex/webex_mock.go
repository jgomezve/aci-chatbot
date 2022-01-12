package webex

import (
	"fmt"
)

type WebexClientMocks struct {
	tkn     string
	baseURL string
}

var (
	WebexMockClient WebexClientMocks
	GetBotDetailsF  = func() (WebexPeople, error) {
		return WebexPeople{
			Id:          "ABC123",
			Emails:      []string{"test@bot.com"},
			DisplayName: "Test Bot",
			NickName:    "Test",
			FirstName:   "Test",
			LastName:    "Bot",
			UserName:    "testbot",
		}, nil
	}
)

func (wbx *WebexClientMocks) GetBotDetails() (WebexPeople, error) {
	return GetBotDetailsF()
}

func (wbx *WebexClientMocks) SendMessageToRoom(m string, roomId string) error {
	fmt.Printf("Sending message")
	return nil
}

func (wbx *WebexClientMocks) DeleteWebhook(name, tUrl, id string) error {
	fmt.Printf("Deleting Webhook")
	return nil
}

func (wbx *WebexClientMocks) CreateWebhook(name, url, resource, event string) error {
	fmt.Printf("Creating Webhook")
	return nil
}

func (wbx *WebexClientMocks) GetWebHooks() ([]WebexWebhook, error) {
	fmt.Printf("Getting Webhook")
	return []WebexWebhook{{Name: "Test"}}, nil
}

func (wbx *WebexClientMocks) GetPersonInfromation(id string) (WebexPeople, error) {
	return WebexPeople{DisplayName: "Test"}, nil
}

func (wbx *WebexClientMocks) GetMessages(roomId string, max int) ([]WebexMessage, error) {
	return []WebexMessage{{Text: "Test", PersonId: "Test"}}, nil
}
