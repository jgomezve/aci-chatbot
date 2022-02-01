package webex

// Package webex uses httptest to execute unit tests
import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Webex interface. Implemented by WebexClient and WebexClientMocks
type WebexInterface interface {
	SendMessageToRoom(m string, roomId string) error
	GetBotDetails() (WebexPeople, error)
	GetWebHooks() ([]WebexWebhook, error)
	DeleteWebhook(id string) error
	CreateWebhook(name, url, resource, event string) error
	GetPersonInformation(id string) (WebexPeople, error)
	GetMessageById(id string) (WebexMessage, error)
	GetRoomById(roomId string) (WebexRoom, error)
}

// Webex Client struct
type WebexClient struct {
	httpClient *http.Client // Webex Client expect an HttpClient interface type
	tkn        string
	baseURL    string
}

// Create a new Webex Client
func NewWebexClient(tkn string) WebexClient {
	wbx := WebexClient{
		tkn: tkn,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: "https://webexapis.com",
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return wbx
}

// Get client base URL
func (wbx *WebexClient) GetBaseUrl() string {
	return wbx.baseURL
}

// Get existing webhooks
func (wbx *WebexClient) GetWebHooks() ([]WebexWebhook, error) {

	var result WebexWebhookReply

	err := wbx.processMessage(http.MethodGet, "/v1/webhooks", nil, &result)
	if err != nil {
		return result.Webhooks, err
	}

	return result.Webhooks, nil
}

// Update existing webhook with new URL
func (wbx *WebexClient) UpdateWebhook(name, tUrl, id string) error {

	err := wbx.processMessage(http.MethodPut, fmt.Sprintf("/v1/webhooks/%s", id), WebexWebhook{Name: name, TargetUrl: tUrl}, nil)
	if err != nil {
		return err
	}

	return nil
}

// Create webhook
func (wbx *WebexClient) CreateWebhook(name, url, resource, event string) error {

	err := wbx.processMessage(http.MethodPost, "/v1/webhooks", WebexWebhook{Name: name, TargetUrl: url, Resource: resource, Event: event}, nil)
	if err != nil {
		return err
	}

	return nil
}

// Delete webhook by id
func (wbx *WebexClient) DeleteWebhook(id string) error {

	err := wbx.processMessage(http.MethodDelete, fmt.Sprintf("/v1/webhooks/%s", id), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Get User detatils
func (wbx *WebexClient) GetBotDetails() (WebexPeople, error) {

	var result WebexPeople
	err := wbx.processMessage(http.MethodGet, "/v1/people/me", nil, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Get Person information by ID
func (wbx *WebexClient) GetPersonInformation(id string) (WebexPeople, error) {

	var result WebexPeopleReply

	err := wbx.processMessage(http.MethodGet, fmt.Sprintf("/v1/people?id=%s", id), nil, &result)
	if err != nil {
		return WebexPeople{}, err
	}
	if len(result.People) > 1 {
		return WebexPeople{}, errors.New("id %s belongs to two different persons ?!?")
	}
	if len(result.People) == 0 {
		return WebexPeople{}, errors.New("person not found")
	}
	return result.People[0], nil
}

// Get message information by  ID
func (wbx *WebexClient) GetMessageById(id string) (WebexMessage, error) {
	var result WebexMessage

	err := wbx.processMessage(http.MethodGet, fmt.Sprintf("/v1/messages/%s", id), nil, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Send a markdown message to a Webex Room
func (wbx *WebexClient) SendMessageToRoom(m, roomId string) error {

	err := wbx.processMessage(http.MethodPost, "/v1/messages", WebexMessage{RoomId: roomId, Markdown: m}, nil)
	if err != nil {
		return err
	}
	return nil
}

// Get room information by ID
func (wbx *WebexClient) GetRoomById(roomId string) (WebexRoom, error) {
	var result WebexRoom

	err := wbx.processMessage(http.MethodGet, fmt.Sprintf("/v1/rooms/%s", roomId), nil, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Get list of subscribed rooms
func (wbx *WebexClient) GetRooms() ([]WebexRoom, error) {

	var result WebexRoomsReply

	err := wbx.processMessage(http.MethodGet, "/v1/rooms", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Rooms, nil
}

/// Create and exectute and HTTP request
func (wbx *WebexClient) processMessage(method, url string, payload interface{}, response interface{}) error {

	req, err := wbx.makeCall(method, url, payload)
	if err != nil {
		return err
	}

	err = wbx.doCall(req, response)
	if err != nil {
		return err
	}

	return nil
}

// Create a HTTP request
func (wbx *WebexClient) makeCall(m, url string, p interface{}) (*http.Request, error) {

	jp, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(m, wbx.baseURL+url, bytes.NewBuffer(jp))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+wbx.tkn)

	return req, nil
}

// Execute a HTTP request
func (wbx *WebexClient) doCall(req *http.Request, res interface{}) error {

	resp, err := wbx.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error processing this request %s\n API message %s", req.URL, body)

	}
	if resp.StatusCode != http.StatusNoContent {
		err = json.Unmarshal(body, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
