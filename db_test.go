package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDB(t *testing.T) {
	db := NewChatDB("./chat.db")

	tabList := []string{"user", "room", "room_user", "message"}

	for _, tab := range tabList {
		var msgList []map[string]interface{}
		db.conn.SQL(fmt.Sprintf("select * from %s", tab)).Find(&msgList)
		fmt.Printf("==========%s=========\n", tab)
		for _, msg := range msgList {
			bts, _ := json.Marshal(msg)
			fmt.Printf("===> %s\n", string(bts))
		}
	}
}
