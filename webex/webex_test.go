package webex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func notOk(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestClient(t *testing.T) {
	wbx := NewWebexClient("FAKETOKEN")

	if wbx.tkn != "FAKETOKEN" {
		t.Errorf("Wrong token")
	}
}

func TestGetRoomsId(t *testing.T) {

	roomsM := []WebexRoom{{Id: "AAVVDD"}}
	roomReplyM := WebexRoomsReply{
		Rooms: roomsM,
	}
	jsonBytes, _ := json.Marshal(roomReplyM)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		equals(t, req.URL.String(), "/v1/rooms")
		rw.Write(jsonBytes)
	}))
	defer server.Close()

	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	rooms, err := client.GetRooms()
	ok(t, err)
	equals(t, roomReplyM.Rooms, rooms)
}

func TestSendMessageToRoom(t *testing.T) {
	messageM := WebexMessage{Text: "First Message"}

	jsonBytes, _ := json.Marshal(messageM)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		equals(t, req.URL.String(), "/v1/messages")
		rw.WriteHeader(200)
		rw.Write(jsonBytes)
	}))
	defer server.Close()

	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.SendMessageToRoom("MyMessage", "ABCD")

	ok(t, err)
}

func TestSendMessageToRoomFail(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		equals(t, req.URL.String(), "/v1/messages")
		rw.WriteHeader(400)
	}))
	defer server.Close()

	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.SendMessageToRoom("MyMessage", "ABCD")

	notOk(t, err)
}

func TestGetMessages(t *testing.T) {
	messageM := []WebexMessage{{Text: "First Message"}}
	meesageReplyM := WebexMessagesReply{
		Messages: messageM,
	}
	jsonBytes, _ := json.Marshal(meesageReplyM)

	roomId := "123AbC"
	maxMsgs := 1
	url := "/v1/messages?" + "roomId=" + roomId + "&max=" + fmt.Sprint(maxMsgs)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		equals(t, req.URL.String(), url)
		rw.Write(jsonBytes)
	}))
	defer server.Close()

	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	messages, err := client.GetMessages(roomId, maxMsgs)
	ok(t, err)
	equals(t, meesageReplyM.Messages, messages)
}
