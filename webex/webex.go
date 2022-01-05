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
	RoomId string `json:"roomId"`
	Text   string `json:"text"`
}

type WebexWebhook struct {
	Id        string           `json:"id"`
	Name      string           `json:"name"`
	TargetUrl string           `json:"targetUrl"`
	Resource  string           `json:"resource"`
	Event     string           `json:"event"`
	OrgId     string           `json:"orgId"`
	CreatedBy string           `json:"createdBy"`
	AppId     string           `json:"appId"`
	OwnerId   string           `json:"OwnerId"`
	Status    string           `json:"status"`
	Created   string           `json:"created"`
	ActorId   string           `json:"actorId"`
	Data      WebexWebhookData `json:"data"`
}

type WebexWebhookData struct {
	Id          string `json:"id"`
	RoomId      string `json:"roomId"`
	RoomType    string `json:"roomType"`
	PersonId    string `json:"personId"`
	PersonEmail string `json:"personEmail"`
	Created     string `json:"created"`
}

// HttpClient interface type
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
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

func (wbx *WebexClient) GetMessages(roomId string, max int) ([]WebexMessageR, error) {
	var result WebexMessagesReply
	url := "/v1/messages?" + "roomId=" + roomId + "&max=" + fmt.Sprint(max)

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
		RoomId: roomId,
		Text:   m,
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
	fmt.Printf("-- %s ---", jp)

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

	fmt.Printf("-- %s ---", req.Body)
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error processing this request %s\n API message %s", req.URL, body)

	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return errors.New("unable to read the response body")
	}
	return nil
}
