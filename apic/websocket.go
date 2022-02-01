package apic

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type ApicWebSocket struct {
	ip  string
	ws  *websocket.Conn
	tkn string
	dl  *websocket.Dialer
}

func NewApicWebSClient(ip string, token string) (*ApicWebSocket, error) {
	u := fmt.Sprintf("wss://10.49.208.146/socket%s", token)
	d := *websocket.DefaultDialer
	d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	log.Printf("Setting up Websocket...")
	wsc, _, err := d.Dial(u, nil)
	if err != nil {
		log.Printf("Error setting up the websocket connection . Error %s", err)
		return nil, err
	}
	aws := ApicWebSocket{ip: ip, ws: wsc, tkn: token, dl: &d}

	return &aws, nil
}

func (aws *ApicWebSocket) NewDial(token string) error {
	ws, _, err := aws.dl.Dial(fmt.Sprintf("wss://10.49.208.146/socket%s", token), nil)
	if err != nil {
		log.Printf("Error setting up the websocket connection . Error %s", err)
		return err
	}
	aws.ws = ws
	return nil
}

func (aws *ApicWebSocket) readSocket(data interface{}) error {

	_, message, err := aws.ws.ReadMessage()
	if err != nil {
		log.Printf("Error reading the WebSocket data. Error %s", err)
		return err
	}

	if err = json.Unmarshal(message, &data); err != nil {
		log.Printf("Error unmarshalling Websocket. Error %s", err)
		return err
	}
	return nil
}

func (aws *ApicWebSocket) ReadSocketEvent() (string, []map[string]interface{}) {

	var result map[string]interface{}
	var events []map[string]interface{}

	aws.readSocket(&result)

	for _, item := range result["imdata"].([]interface{}) {
		event := make(map[string]interface{})
		for k1, v1 := range item.(map[string]interface{}) {
			event["class"] = k1
			event["dn"] = v1.(map[string]interface{})["attributes"].(map[string]interface{})["dn"].(string)
			event["status"] = v1.(map[string]interface{})["attributes"].(map[string]interface{})["status"].(string)
			changed_attributes := []string{}
			for k2, v2 := range v1.(map[string]interface{})["attributes"].(map[string]interface{}) {
				changed_attributes = append(changed_attributes, fmt.Sprintf("%s:%s", k2, v2))
			}
			event["changed_attributes"] = changed_attributes
			events = append(events, event)
		}
	}

	subId := result["subscriptionId"].([]interface{})[0].(string)
	return subId, events
}
