package bot

import (
	"fmt"
	"testing"
)

func TestCreateDb(t *testing.T) {
	db := NewWsDb()
	fmt.Println("Ok")
	equals(t, len(db.wss), 0)
}

func TestAddAndRemoveItem(t *testing.T) {
	db := NewWsDb()
	t.Run("Add single item", func(t *testing.T) {
		db.addSubcription("myClass", "12345", "ABC123")
		equals(t, db.wss["myClass"].SubscriptionID, "12345")
		equals(t, db.wss["myClass"].RoomsId[0], "ABC123")
		equals(t, len(db.wss["myClass"].RoomsId), 1)
	})

	t.Run("Add two items to same class", func(t *testing.T) {
		db.addSubcription("myClass", "12345", "DEF456")
		equals(t, db.wss["myClass"].SubscriptionID, "12345")
		equals(t, db.wss["myClass"].RoomsId[0], "ABC123")
		equals(t, len(db.wss["myClass"].RoomsId), 2)
	})
	t.Run("Remove item from a class", func(t *testing.T) {
		db.addSubcription("myClass", "12345", "GHI789")
		db.removeSubcription("myClass", "DEF456")
		equals(t, db.wss["myClass"].SubscriptionID, "12345")
		equals(t, db.wss["myClass"].RoomsId[0], "ABC123")
		equals(t, db.wss["myClass"].RoomsId[1], "GHI789")
		equals(t, len(db.wss["myClass"].RoomsId), 2)
	})
	t.Run("Remove all items from a class", func(t *testing.T) {
		db.removeSubcription("myClass", "ABC123")
		db.removeSubcription("myClass", "GHI789")
		actSub := db.getActiveSubscriptions()
		equals(t, actSub, map[string]string{})

	})
}

func TestCheckState(t *testing.T) {
	db := NewWsDb()
	t.Run("Check valid item", func(t *testing.T) {
		db.addSubcription("myClass", "12345", "ABC123")
		exists := db.checkSubsciption("myClass", "ABC123")
		equals(t, exists, true)
	})
	db = NewWsDb()
	t.Run("Check invalid item - empty list", func(t *testing.T) {
		exists := db.checkSubsciption("myClass", "ABC123")
		equals(t, exists, false)
	})
	t.Run("Check invalid item - non-empty list", func(t *testing.T) {
		db.addSubcription("myClass", "12345", "ABC123")
		exists := db.checkSubsciption("myClass", "DEF456")
		equals(t, exists, false)
	})
}

func TestGetFunctions(t *testing.T) {
	db := NewWsDb()
	t.Run("getClassesbyRoomId", func(t *testing.T) {
		db.addSubcription("fvTenant", "11", "ABC123")
		db.addSubcription("fvBD", "22", "ABC123")
		db.addSubcription("fvBD", "33", "DEF456")
		db.addSubcription("fvCtx", "44", "ABC123")
		classes := db.getClassesbyRoomId("ABC123")
		equals(t, len(classes), 3)
	})

	t.Run("getActiveSubscriptions", func(t *testing.T) {
		subs := db.getActiveSubscriptions()
		equals(t, subs["fvTenant"], "11")
		equals(t, subs["fvBD"], "22")
		equals(t, subs["fvCtx"], "44")
	})

	t.Run("getRoomsIdbyClass", func(t *testing.T) {
		rooms := db.getRoomsIdbyClass("fvBD")
		equals(t, len(rooms), 2)
		rooms = db.getRoomsIdbyClass("fvTenant")
		equals(t, len(rooms), 1)
		equals(t, rooms[0], "ABC123")
	})
	t.Run("getClassNamebySubId", func(t *testing.T) {
		class := db.getClassNamebySubId("44")
		equals(t, class, "fvCtx")
		class = db.getClassNamebySubId("22")
		equals(t, class, "fvBD")
	})
}
