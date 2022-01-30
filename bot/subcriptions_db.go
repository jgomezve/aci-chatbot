package bot

import (
	"log"
)

// struct to represent the Websocket subscriptions

type webSocketDb struct {
	wss map[string]SocketSubscription
}

// struct to represent the Websocket subscriptions
type SocketSubscription struct {
	SubscriptionID string
	RoomsId        []string
}

// Create new DB Instance
func NewWsDb() *webSocketDb {
	mapDb := make(map[string]SocketSubscription)
	db := webSocketDb{wss: mapDb}
	return &db
}

// Get the Class/MO Name based on the SubscriptionID
func (wsDb *webSocketDb) getClassNamebySubId(subId string) string {
	className := ""
	for class, sub := range wsDb.wss {
		if sub.SubscriptionID == subId {
			className = class
		}
	}
	return className
}

// Get the Subscribed RoomsID by Class/MO name
func (wsDb *webSocketDb) getRoomsIdbyClass(class string) []string {
	return wsDb.wss[class].RoomsId
}

// Get the list of Claas/MO : SubscriptionId
func (wsDb *webSocketDb) getActiveSubscriptions() map[string]string {
	subs := make(map[string]string)
	for class, sub := range wsDb.wss {
		subs[class] = sub.SubscriptionID
	}
	return subs
}

// Add a new subscription from a room
func (wsDb *webSocketDb) addSubcription(class string, subId string, roomId string) {

	if entry, ok := wsDb.wss[class]; !ok {
		log.Printf("New entry for MO/Class %s\n", class)
		wsDb.wss[class] = SocketSubscription{SubscriptionID: subId, RoomsId: []string{roomId}}
	} else {
		//TODO: Fix this update
		log.Printf("Existing entry for MO/Class %s\n", class)
		log.Printf("Adding Room ID %s\n", roomId)
		entry.RoomsId = append(entry.RoomsId, roomId)
	}

}

// Get the classes subscribed in a Room
func (wsDb *webSocketDb) getClassesbyRoomId(roomId string) []string {
	classes := []string{}
	for class, sub := range wsDb.wss {
		for _, room := range sub.RoomsId {
			if room == roomId {
				classes = append(classes, class)
			}
		}
	}
	return classes
}
