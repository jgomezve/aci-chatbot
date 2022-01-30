package webex

// Package webex uses httptest to execute unit tests
import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// HttpClient interface type.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TODO: Change name
type WebexInterface interface {
	SendMessageToRoom(m string, roomId string) error
	GetBotDetails() (WebexPeople, error)
	GetWebHooks() ([]WebexWebhook, error)
	DeleteWebhook(name, tUrl, id string) error
	CreateWebhook(name, url, resource, event string) error
	GetPersonInformation(id string) (WebexPeople, error)
	GetMessages(roomId string, max int, filter ...string) ([]WebexMessage, error)
	GetMessageById(id string) (WebexMessage, error)
	GetRoomById(roomId string) (WebexRoom, error)
}

type WebexClient struct {
	httpClient HttpClient // Webex Client expect an HttpClient interface type
	tkn        string
	baseURL    string
}

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

func (wbx *WebexClient) GetBaseUrl() string {
	return wbx.baseURL
}

func (wbx *WebexClient) CreateWebhook(name, url, resource, event string) error {
	req, err := wbx.makeCall(http.MethodPost, "/v1/webhooks", WebexWebhook{
		Name:      name,
		TargetUrl: url,
		Resource:  resource,
		Event:     event,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	err = wbx.doCall(req, nil)

	if err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (wbx *WebexClient) GetWebHooks() ([]WebexWebhook, error) {

	var result WebexWebhookReply

	req, err := wbx.makeCall(http.MethodGet, "/v1/webhooks", nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return []WebexWebhook{}, err
	}

	err = wbx.doCall(req, &result)

	if err != nil {
		log.Println("Error: ", err)
		return []WebexWebhook{}, err
	}

	return result.Webhooks, nil
}

func (wbx *WebexClient) UpdateWebhook(name, tUrl, id string) error {
	url := "/v1/webhooks/" + id
	req, err := wbx.makeCall(http.MethodPut, url, WebexWebhook{
		Name:      name,
		TargetUrl: tUrl,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	err = wbx.doCall(req, nil)

	if err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (wbx *WebexClient) DeleteWebhook(name, tUrl, id string) error {
	url := "/v1/webhooks/" + id
	req, err := wbx.makeCall(http.MethodDelete, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	err = wbx.doCall(req, nil)

	if err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (wbx *WebexClient) GetBotDetails() (WebexPeople, error) {

	var result WebexPeople
	url := "/v1/people/me"
	req, err := wbx.makeCall(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return WebexPeople{}, err
	}

	err = wbx.doCall(req, &result)
	if err != nil {
		log.Println("Error: ", err)
		return WebexPeople{}, err
	}

	return result, nil
}

func (wbx *WebexClient) GetPersonInformation(id string) (WebexPeople, error) {

	var result WebexPeopleReply
	url := "/v1/people?" + "id=" + id
	req, err := wbx.makeCall(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return WebexPeople{}, err
	}

	err = wbx.doCall(req, &result)
	if err != nil {
		log.Println("Error: ", err)
		return WebexPeople{}, err
	}
	if len(result.People) != 1 {
		return WebexPeople{}, errors.New("id %s belongs to two different persons ?!?")
	}

	return result.People[0], nil
}

func (wbx *WebexClient) GetMessages(roomId string, max int, filter ...string) ([]WebexMessage, error) {
	var result WebexMessagesReply
	url := fmt.Sprintf("/v1/messages?roomId=%s&max=%d", roomId, max)

	for _, f := range filter {
		url = url + fmt.Sprintf("&%s", f)
	}

	req, err := wbx.makeCall(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result.Messages, err
	}

	err = wbx.doCall(req, &result)
	if err != nil {
		log.Println("Error: ", err)
		return result.Messages, err
	}

	return result.Messages, nil
}

func (wbx *WebexClient) GetMessageById(id string) (WebexMessage, error) {
	var result WebexMessage
	url := fmt.Sprintf("/v1/messages/%s", id)

	req, err := wbx.makeCall(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result, err
	}

	err = wbx.doCall(req, &result)
	if err != nil {
		log.Println("Error: ", err)
		return result, err
	}

	return result, nil
}

func (wbx *WebexClient) GetRoomById(roomId string) (WebexRoom, error) {
	var result WebexRoom
	req, err := wbx.makeCall(http.MethodGet, fmt.Sprintf("/v1/rooms/%s", roomId), nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result, err
	}

	err = wbx.doCall(req, &result)

	if err != nil {
		log.Println("Error: ", err)
		return result, err
	}

	return result, nil
}

func (wbx *WebexClient) GetRoomIds() ([]WebexRoom, error) {

	var result WebexRoomsReply

	req, err := wbx.makeCall(http.MethodGet, "/v1/rooms", nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result.Rooms, err
	}

	err = wbx.doCall(req, &result)

	if err != nil {
		log.Println("Error: ", err)
		return result.Rooms, err
	}

	return result.Rooms, nil
}

func (wbx *WebexClient) SendMessageToRoom(m string, roomId string) error {

	req, err := wbx.makeCall(http.MethodPost, "/v1/messages", WebexMessage{
		RoomId:   roomId,
		Markdown: m,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	err = wbx.doCall(req, nil)

	if err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (wbx *WebexClient) makeCall(m string, url string, p interface{}) (*http.Request, error) {

	jp, err := json.Marshal(p)

	if err != nil {
		return nil, errors.New("unable to marshal the paylaod")
	}
	req, err := http.NewRequest(m, wbx.baseURL+url, bytes.NewBuffer(jp))
	if err != nil {
		return nil, errors.New("unable to create a new HTTP request")
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+wbx.tkn)

	return req, nil
}

func (wbx *WebexClient) doCall(req *http.Request, res interface{}) error {

	resp, err := wbx.httpClient.Do(req)
	if err != nil {
		return errors.New("unable to send the HTTP request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return errors.New("unable to read the response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error processing this request %s\n API message %s", req.URL, body)

	}
	if resp.StatusCode != http.StatusNoContent {
		err = json.Unmarshal(body, &res)
	}

	if err != nil {
		return errors.New("unable to serialize response body")
	}
	return nil
}
