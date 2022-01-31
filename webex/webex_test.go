package webex

import (
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

// Test creation of the client
func TestClient(t *testing.T) {
	wbx := NewWebexClient("FAKETOKEN")

	equals(t, "https://webexapis.com", wbx.GetBaseUrl())
}

// Test the functions talking to the /room URI
func TestRoomsUriOk(t *testing.T) {
	roomTest := `{
		"id": "AAVVDD",
		"title": "A Test Room",
		"type": "", 
		"isLocked": false, 
		"lastActivity": "", 
		"creatorId": "", 
		"created": "", 
		"ownerId": ""
	}`
	roomsTest := fmt.Sprintf(`{
		"items": [
			%s
		]
	}`, roomTest)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/v1/rooms":
			rw.Write([]byte(roomsTest))
		case "/v1/rooms/AAVVDD":
			rw.Write([]byte(roomTest))
		}
	}))
	defer server.Close()
	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	t.Run("Get Rooms", func(t *testing.T) {
		rooms, err := client.GetRooms()
		ok(t, err)
		equals(t, 1, len(rooms))
	})

	t.Run("Get Room ID", func(t *testing.T) {
		room, err := client.GetRoomById("AAVVDD")
		ok(t, err)
		equals(t, "AAVVDD", room.Id)
		equals(t, "A Test Room", room.Title)
	})

}

// Test the functions talking to the /message URI
func TestMessageUriOk(t *testing.T) {

	messageTest1 := `{
		"id": "A1B2C3",
		"roomId": "AABB",
		"personId": "CCDD"
	}`
	messageTest2 := `{
		"id": "A1B2C3",
		"roomId": "AABB",
		"personId": "CCDD"
	}`
	messagesTest := fmt.Sprintf(`{
		"items": [
			%s,
			%s
		]
	}`, messageTest1, messageTest2)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/v1/messages?roomId=AABB&max=2":
			rw.Write([]byte(messagesTest))
			rw.WriteHeader(200)
		case "/v1/messages":
			rw.WriteHeader(200)
			rw.Write([]byte(messageTest2))
		case "/v1/messages/A1B2C3":
			rw.Write([]byte(messageTest1))
		}
	}))
	defer server.Close()
	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL
	t.Run("Send Message", func(t *testing.T) {
		err := client.SendMessageToRoom("TestMsg", "ABCD")
		ok(t, err)
	})

	t.Run("Get Messages", func(t *testing.T) {
		msg, err := client.GetMessages("AABB", 2)
		ok(t, err)
		equals(t, "AABB", msg[0].RoomId)
		equals(t, "CCDD", msg[0].PersonId)
		equals(t, "A1B2C3", msg[0].Id)
		equals(t, 2, len(msg))
	})

	t.Run("Get Message by ID", func(t *testing.T) {
		msg, err := client.GetMessageById("A1B2C3")
		ok(t, err)
		equals(t, "AABB", msg.RoomId)
		equals(t, "CCDD", msg.PersonId)
		equals(t, "A1B2C3", msg.Id)
	})

}

// Test the functions talking to the /Webhook URI
func TestWebhookUriOk(t *testing.T) {

	webhookTest1 := `{
		"id": "AAAA",
		"name": "Webhook1",
		"data": {
			"id": "AAAA"

		}
	}`
	webhookTest2 := `{
		"id": "BBBB",
		"name": "Webhook2",
		"data": {
			"id": "BBBB"

		}
	}`
	webhooksTest := fmt.Sprintf(`{
		"items": [
			%s,
			%s
		]
	}`, webhookTest1, webhookTest2)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println(req.URL.String())
		if req.URL.String() == "/v1/webhooks" && req.Method == http.MethodGet {
			rw.Write([]byte(webhooksTest))
		} else if req.URL.String() == "/v1/webhooks" && req.Method == http.MethodPost {
			rw.Write([]byte(webhookTest2))
		} else if req.URL.String() == "/v1/webhooks/BBBB" && req.Method == http.MethodDelete {
			rw.Write([]byte(webhookTest2))
		} else if req.URL.String() == "/v1/webhooks/BBBB" && req.Method == http.MethodPut {
			rw.Write([]byte(webhookTest2))
		}
	}))
	defer server.Close()
	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	t.Run("Get Webhooks", func(t *testing.T) {
		whs, err := client.GetWebHooks()
		ok(t, err)
		equals(t, 2, len(whs))
	})

	t.Run("Create Webhook", func(t *testing.T) {
		err := client.CreateWebhook("myWebhooks", "https://test.com", "messages", "created")
		ok(t, err)
	})

	t.Run("Delete Webhook", func(t *testing.T) {
		err := client.DeleteWebhook("BBBB")
		ok(t, err)
	})

	t.Run("Update Webhook", func(t *testing.T) {
		err := client.UpdateWebhook("New Name", "https://test.com", "BBBB")
		ok(t, err)
	})
}

// Test the functions talking to the /People URI
func TestPeopleUriOk(t *testing.T) {

	peopleTest := `{
		"id": "AAAA",
		"displayName": "Person1"
	}`
	people := fmt.Sprintf(`{
		"items": [
			%s
		]
	}`, peopleTest)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/v1/people/me":
			rw.Write([]byte(peopleTest))
		case "/v1/people?id=AAAA":
			rw.Write([]byte(people))
		}
	}))
	defer server.Close()
	client := NewWebexClient("FAKETOKEN")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	t.Run("Get People", func(t *testing.T) {
		_, err := client.GetPersonInformation("AAAA")
		ok(t, err)
	})
	t.Run("Get People Me", func(t *testing.T) {
		_, err := client.GetBotDetails()
		ok(t, err)
	})
}

// func TestSendMessageToRoomFail(t *testing.T) {

// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		equals(t, req.URL.String(), "/v1/messages")
// 		rw.WriteHeader(400)
// 	}))
// 	defer server.Close()

// 	client := NewWebexClient("FAKETOKEN")
// 	client.httpClient = server.Client()
// 	client.baseURL = server.URL

// 	err := client.SendMessageToRoom("MyMessage", "ABCD")

// 	notOk(t, err)
// }
