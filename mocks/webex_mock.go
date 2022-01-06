package mocks

import (
	"chatbot/webex"
	"fmt"
)

type WebexClientMocks struct {
	tkn     string
	baseURL string
}

var (
	WebexMockClient WebexClientMocks
)

func (wbx *WebexClientMocks) GetBotDetails() (webex.WebexPeople, error) {

	return webex.WebexPeople{
			NickName: "TestBot",
		},
		nil
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
	fmt.Printf("Creatiing Webhook")
	return nil
}

func (wbx *WebexClientMocks) GetWebHooks() ([]webex.WebexWebhook, error) {
	fmt.Printf("Getting Webhook")
	return []webex.WebexWebhook{{Name: "Test"}}, nil
}

func (wbx *WebexClientMocks) GetPersonInfromation(id string) (webex.WebexPeople, error) {
	return webex.WebexPeople{DisplayName: "Test"}, nil
}

func (wbx *WebexClientMocks) GetMessages(roomId string, max int) ([]webex.WebexMessage, error) {
	return []webex.WebexMessage{{Text: "Test", PersonId: "Test"}}, nil
}
