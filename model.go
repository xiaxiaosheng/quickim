package main

import "time"

type UserModel struct {
	Id           int64
	UserId       string
	Nickname     string
	ChannelId    int64
	LastLiveTime time.Time
	CreateTime   time.Time
}

func (*UserModel) TableName() string {
	return "user"
}

type RoomModel struct {
	Id         int64
	RoomId     string
	RoomName   string
	CreateTime time.Time
}

func (*RoomModel) TableName() string {
	return "room"
}

type RoomUserModel struct {
	Id         int64
	RoomId     string
	UserId     string
	CreateTime time.Time
}

func (*RoomUserModel) TableName() string {
	return "room_user"
}

type MessageModel struct {
	Id         int64
	RoomId     string
	UserId     string
	Content    string
	CreateTime time.Time
}

func (*MessageModel) TableName() string {
	return "message"
}

type HistoryMessageItem struct {
	*MessageModel `xorm:"extends"`
	Nickname      string
}

type RoomUserItem struct {
	*RoomUserModel `xorm:"extends"`
	Nickname       string
	ChannelId      int64
}
