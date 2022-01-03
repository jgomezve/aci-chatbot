package main

import (
	"chatbox/apic"
	"chatbox/webex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

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

type MessageClient interface {
	GetMessages(roomId string, max int) ([]webex.WebexMessageR, error)
	SendMessageToRoom(m string, roomId string) error
}

func messageHandler(client MessageClient) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload WebexWebhook
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error: ", err)
		}
		err = json.Unmarshal(body, &payload)
		if err != nil {
			log.Println("Error: ", err)
		}
		messages, _ := client.GetMessages(payload.Data.RoomId, 1)
		if messages[0].PersonEmail != "cx-germany-bot@webex.bot" {
			client.SendMessageToRoom("RE: "+messages[0].Text, payload.Data.RoomId)
		}
	})
}

func main() {

	_, err := apic.NewApicClient("https://10.49.208.146/", "jgomezve", "Colombia5a", apic.SetTimeout(5))
	if err == nil {
		fmt.Println("Logged in successfully")
	}
	wbx := webex.NewWebexClient("MDRiNDRjN2EtNWQxOS00MmU1LWE4ZDAtODIwMmI2MjUxMzY0NWE4OGFmOTItODk5_PF84_1eb65fdf-9643-417f-9974-ad72cae0e10f")
	http.HandleFunc("/message", messageHandler(&wbx))
	http.ListenAndServe(":7001", nil)
	wbx.SendMessageToRoom("Testing", "Y2lzY29zcGFyazovL3VzL1JPT00vZjRmZWZjZDAtNjI3NS0xMWVjLThiMTQtMDEyYWYxZGQ1M2Vl")

}
