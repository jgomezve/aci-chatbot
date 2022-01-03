package webex

// Package webex uses httptest to execute unit tests
import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (wbx *WebexClient) GetMessages(roomId string, max int) ([]WebexMessageR, error) {
	var result WebexMessagesReply
	url := "/v1/messages?" + "roomId=" + roomId + "&max=" + fmt.Sprint(max)

	_, content, err := wbx.makeCall(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result.Messages, err
	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		log.Println("Error: ", err)
		return result.Messages, err
	}

	return result.Messages, nil

}

func (wbx *WebexClient) GetRoomIds() ([]WebexRoom, error) {

	var result WebexRoomsReply

	_, content, err := wbx.makeCall(http.MethodGet, "/v1/rooms", nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return result.Rooms, err
	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		log.Println("Error: ", err)
		return result.Rooms, err
	}

	return result.Rooms, nil
}

func (wbx *WebexClient) SendMessageToRoom(m string, roomId string) error {
	var payload WebexMessage
	payload.RoomId = roomId
	payload.Text = m

	jsonValue, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	code, _, err := wbx.makeCall(http.MethodPost, "/v1/messages", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	if code == 400 {
		return fmt.Errorf("unknown room id %s", roomId)
	}
	return nil
}

func (wbx *WebexClient) makeCall(m string, url string, p io.Reader) (int, []byte, error) {
	req, err := http.NewRequest(m, wbx.baseURL+url, p)
	if err != nil {
		return 0, nil, errors.New("unable to create a new HTTP request")
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+wbx.tkn)

	resp, err := wbx.httpClient.Do(req)
	if err != nil {
		return 0, nil, errors.New("unable to send the HTTP request")
	}

	// Why defer ?
	defer resp.Body.Close()
	print(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, nil, errors.New("unable to read the response body")
	}
	return resp.StatusCode, body, nil

}
