package bot

// struct to represent a DB storing the active subscription
// This could be improved by storgin the information in a external DB
//  This would make the application stateless
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

// Add a new subscription to a room
func (wsDb *webSocketDb) addSubcription(class string, subId string, roomId string) {
	if entry, ok := wsDb.wss[class]; !ok {
		wsDb.wss[class] = SocketSubscription{SubscriptionID: subId, RoomsId: []string{roomId}}
	} else {
		entry.RoomsId = append(entry.RoomsId, roomId)
		wsDb.wss[class] = entry
	}
}

// Remove a new subscription from a room
func (wsDb *webSocketDb) removeSubcription(class string, roomId string) {

	if entry, ok := wsDb.wss[class]; ok {
		for idx, room := range entry.RoomsId {
			if room == roomId {
				entry.RoomsId[idx] = entry.RoomsId[len(entry.RoomsId)-1]
				entry.RoomsId = entry.RoomsId[:len(entry.RoomsId)-1]
				wsDb.wss[class] = entry
			}
		}
		if len(wsDb.wss[class].RoomsId) == 0 {
			delete(wsDb.wss, class)
		}
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

// Get the classes subscribed in a Room
func (wsDb *webSocketDb) checkSubsciption(class string, roomId string) bool {
	for _, room := range wsDb.wss[class].RoomsId {
		if room == roomId {
			return true
		}
	}
	return false
}
