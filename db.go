package main

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-xorm/xorm"
	"xorm.io/builder"

	_ "github.com/mattn/go-sqlite3"
)

func NewChatDB(path string) *ChatDB {
	db := &ChatDB{
		path: path,
	}
	db.connect(path)

	return db
}

type ChatDB struct {
	path string
	conn *xorm.Engine
}

func (db *ChatDB) connect(path string) {
	var err error
	db.conn, err = xorm.NewEngine(builder.SQLITE, path)
	if err != nil {
		panic(err)
	}

	db.init()
}

func (db *ChatDB) init() {
	tables := []string{`
CREATE TABLE IF NOT EXISTS user (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     user_id VARCHAR(64) NULL,
	 nickname VARCHAR(64) NULL,
	 channel_id int(11) NULL,
     last_live_time DATETIME NULL,
	 create_time DATETIME NULL
);`,
		`
CREATE TABLE IF NOT EXISTS room (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     room_id VARCHAR(64) NULL,
	 room_name VARCHAR(64) NULL,
	 create_time DATETIME NULL
);
`,
		`
CREATE TABLE IF NOT EXISTS room_user (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     room_id VARCHAR(64) NULL,
	 user_id VARCHAR(64) NULL,
	 create_time DATETIME NULL
);
`,

		`
CREATE TABLE IF NOT EXISTS message (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     room_id VARCHAR(64) NULL,
	 user_id VARCHAR(64) NULL,
	 content TEXT NULL,
	 create_time DATETIME NULL
);
`,
		`update user set channel_id=0;`,
		`delete from room_user;`,
	}
	for _, createSql := range tables {
		if _, err := db.conn.Exec(createSql); err != nil {
			fmt.Println(createSql)
			panic(err)
		}
	}

	if db.GetRoom(squareId) == nil {
		db.CreateRoom(&RoomModel{
			RoomId:     squareId,
			RoomName:   "广场",
			CreateTime: time.Now(),
		})
	}
}

func (db *ChatDB) GetUser(userId string) (*UserModel, error) {
	user := new(UserModel)
	exist, err := db.conn.Where("user_id = ?", userId).Select("*").Get(user)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, err
	}
	return user, nil
}

func (db *ChatDB) GetUserByChannelId(channelId int64) (*UserModel, error) {
	user := new(UserModel)
	exist, err := db.conn.Where("channel_id = ?", channelId).Select("*").Get(user)
	if err != nil {
		return nil, err
	}
	if !exist {
		err = errors.New("channel id not exist")
		return nil, err
	}
	return user, nil
}

func (db *ChatDB) RegisterUser(user *UserModel) (err error) {
	_, err = db.conn.InsertOne(user)
	return
}

func (db *ChatDB) UpdateUser(user *UserModel, cols ...string) (err error) {
	_, err = db.conn.Where("user_id = ?", user.UserId).Cols(cols...).Update(user)
	return
}

func (db *ChatDB) GetRoom(roomId string) *RoomModel {
	room := new(RoomModel)
	exist, err := db.conn.Where("room_id = ?", roomId).Select("*").Get(room)
	if err != nil || !exist {
		return nil
	}

	return room
}

func (db *ChatDB) CreateRoom(room *RoomModel) {
	db.conn.InsertOne(room)
}

func (db *ChatDB) IsUserInRoom(userId, roomId string) (bool, error) {
	return db.conn.Where("room_id = ? and user_id = ?", roomId, userId).Table(&RoomUserModel{}).Exist()
}

func (db *ChatDB) JoinRoom(roomId, userId string) (err error) {
	exist, err := db.IsUserInRoom(userId, roomId)
	if err != nil {
		return err
	}
	if !exist {
		mp := &RoomUserModel{
			RoomId:     roomId,
			UserId:     userId,
			CreateTime: time.Now(),
		}
		_, err = db.conn.Insert(mp)
	}

	return
}

func (db *ChatDB) OutRoom(roomId, userId string) (err error) {
	_, err = db.conn.Where("room_id = ? and user_id = ?", roomId, userId).Delete(&RoomUserModel{})
	if err != nil {
		return
	}
	if roomId == squareId {
		return
	}

	list, _ := db.GetRoomUser(roomId)
	if len(list) == 0 {
		db.DeleteRoom(roomId)
	}

	return
}

func (db *ChatDB) LoadHistoryMsg(roomId string, num int) (msgList []*HistoryMessageItem, err error) {
	err = db.conn.Table(&MessageModel{}).Alias("a").
		Join("LEFT", []string{"user", "b"}, "a.user_id = b.user_id").
		Where("room_id = ?", roomId).Desc("a.create_time").Limit(num).Select("a.*, b.nickname").Find(&msgList)
	if err != nil {
		return
	}
	sort.Slice(msgList, func(i, j int) bool {
		return msgList[i].CreateTime.Before(msgList[j].CreateTime)
	})

	return
}

func (db *ChatDB) GetUserRoomId(channelId int64) (roomId string, err error) {
	exist, err := db.conn.Table(&RoomUserModel{}).Alias("a").
		Join("LEFT", []string{"user", "b"}, "a.user_id = b.user_id").
		Where("channel_id = ?", channelId).
		Select("a.room_id").
		Get(&roomId)
	if err != nil {
		return
	}
	if !exist {
		err = fmt.Errorf("not in room")
		return
	}

	return
}

func (db *ChatDB) GetRoomUser(roomId string) (userList []*RoomUserItem, err error) {
	err = db.conn.Table(&RoomUserModel{}).Alias("a").
		Join("LEFT", []string{"user", "b"}, "a.user_id = b.user_id").
		Where("room_id = ?", roomId).
		Asc("a.create_time").
		Select("a.*, b.nickname, b.channel_id").
		Find(&userList)

	return
}

func (db *ChatDB) ClearRoomMessage(roomId string) {
	db.conn.Where("room_id = ?", roomId).Delete(&MessageModel{})
}

func (db *ChatDB) DeleteRoom(roomId string) {
	db.conn.Where("room_id = ?", roomId).Delete(&RoomModel{})
}

func (db *ChatDB) NewMessage(msg *MessageModel) (err error) {
	_, err = db.conn.InsertOne(msg)
	return
}
