package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	mathRand "math/rand"
	"strings"
	"sync"
	"time"

	logger "github.com/opentrx/seata-golang/v2/pkg/util/log"

	"github.com/go-netty/go-netty"
)

func NewChatManager() *ChatManager {
	ch := &ChatManager{}
	ch.makeRsa()

	return ch
}

type ChatManager struct {
	db      *ChatDB
	dbPath  string
	ctxData map[int64]netty.ActiveContext
	desKey  map[int64]string
	//	userList    map[string]*User
	//	userData    map[int64]*User
	//	roomData    map[string]map[int64]netty.ActiveContext
	//	roomMessage map[string][]*MessageItem
	lock       sync.RWMutex
	privateKey *rsa.PrivateKey
	publicKey  string
}

const squareId = "000000"

/*
 * CMD
RNM 通知房间号
CHM 切换房间号
CHN    修改昵称
MSG 新消息
JNR     加入房间
OUT     退出房间
LOGIN 登录
CLEAR 清空房间消息
PUBKEY 下发公钥
KEY 协商DES密钥
*/

func getNickname() string {
	return namePool[mathRand.Intn(len(namePool))]
}

func (ch *ChatManager) makeRsa() {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	var err error
	ch.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}

	//生成私钥
	//	pkcs1PrivateKey := x509.MarshalPKCS1PrivateKey(ch.privateKey)
	//	block := &pem.Block{
	//		Type:  "RSA PRIVATE KEY",
	//		Bytes: pkcs1PrivateKey,
	//	}
	//	//写入文件
	//	privateFile := &bytes.Buffer{}
	//	pem.Encode(privateFile, block)
	//	fmt.Println(privateFile.String())

	//产生公钥 主要取地址
	PublicKey := &ch.privateKey.PublicKey
	//公钥从私钥中产生
	pkixPublicKey, err := x509.MarshalPKIXPublicKey(PublicKey)
	block1 := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixPublicKey,
	}
	publicFile := &bytes.Buffer{}
	pem.Encode(publicFile, block1)
	ch.publicKey = strings.ReplaceAll(publicFile.String(), "\n", "")
}

func (ch *ChatManager) dataInit() {
	if ch.ctxData == nil {
		ch.ctxData = make(map[int64]netty.ActiveContext)
		ch.desKey = make(map[int64]string)
	}
	if ch.db == nil {
		if ch.dbPath == "" {
			ch.dbPath = "./chat.db"
		}
		ch.db = NewChatDB(ch.dbPath)
	}
}

func (ch *ChatManager) connectInit(actx netty.ActiveContext) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := actx.Channel().ID()

	ch.ctxData[channelId] = actx

	msg := map[string]interface{}{
		"code":       "PUBKEY",
		"public_key": ch.publicKey,
	}
	actx.Write(encodeMsg(ch.desKey[channelId], msg))
}

func (ch *ChatManager) pollRoomUser(roomId string, fn func(*RoomUserItem, netty.ActiveContext)) {
	userList, _ := ch.db.GetRoomUser(roomId)
	for _, u := range userList {
		if actx := ch.ctxData[u.ChannelId]; actx != nil {
			fn(u, actx)
		}
	}
}

func (ch *ChatManager) pushDesKey(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()
	actx := ch.ctxData[channelId]

	keyCipher := fmt.Sprintf("%v", message["key"])

	body, err := base64.StdEncoding.DecodeString(keyCipher)
	if err != nil {
		logger.Error(err)
		return
	}

	plainBts, err := rsa.DecryptPKCS1v15(rand.Reader, ch.privateKey, body)
	if err != nil {
		logger.Error(err)
		return
	}

	ch.desKey[channelId] = string(plainBts)

	msg := map[string]interface{}{
		"code": "SHAKESUCC",
	}

	actx.Write(encodeMsg(ch.desKey[channelId], msg))
}

// 登录
func (ch *ChatManager) login(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()
	actx := ch.ctxData[channelId]

	if message["user_id"] == nil {
		ctx.Close(fmt.Errorf("user id can't be null"))
		return
	}
	userId := fmt.Sprintf("%v", message["user_id"])
	user, err := ch.db.GetUser(userId)
	if err != nil {
		ctx.Close(err)
		logger.Error(err)
		return
	}
	if user == nil {
		user = &UserModel{
			UserId:       userId,
			Nickname:     getNickname(),
			ChannelId:    channelId,
			LastLiveTime: time.Now(),
			CreateTime:   time.Now(),
		}
		if err = ch.db.RegisterUser(user); err != nil {
			logger.Error(err)
			ctx.Close(err)
			return
		} else {
			// logger.Infof("register user:%+v", user)
		}
	} else {
		user.ChannelId = channelId
		user.LastLiveTime = time.Now()
		ch.db.UpdateUser(user, "channel_id", "last_live_time")
	}

	ch.db.JoinRoom(squareId, user.UserId)

	// logger.Infof("[%d] join room user[%+v]", channelId, user)
	jnrReport := map[string]interface{}{
		"code": "JNR",
		"name": user.Nickname,
	}
	rnmMsg := map[string]interface{}{
		"code":       "RNM",
		"roomId":     squareId,
		"nickname":   user.Nickname,
		"onlineList": ch.getOnlineList(squareId),
	}
	actx.Write(encodeMsg(ch.desKey[channelId], rnmMsg))

	// 加入房间通知
	ch.pollRoomUser(squareId, func(u *RoomUserItem, actx netty.ActiveContext) {
		if u.ChannelId != channelId {
			actx.Write(encodeMsg(ch.desKey[u.ChannelId], jnrReport))
		}
	})

	ch.pushHistoryMessage(squareId, channelId)
}

// 推送历史消息
func (ch *ChatManager) pushHistoryMessage(roomId string, channelId int64) {
	actx := ch.ctxData[channelId]

	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		actx.Close(err)
		return
	}

	lastNum := 30 // 加载最近的30条消息
	msgList, err := ch.db.LoadHistoryMsg(roomId, lastNum)
	if err != nil {
		actx.Close(err)
		return
	}

	for _, tmp := range msgList {
		msg := map[string]interface{}{
			"code":     "MSG",
			"from":     tmp.Nickname,
			"message":  tmp.Content,
			"sendTime": tmp.CreateTime.Format("15:04:05"),
		}
		if tmp.UserId == user.UserId {
			msg["is_self"] = 1
		}

		actx.Write(encodeMsg(ch.desKey[channelId], msg))
	}
}

func (ch *ChatManager) outRoom(channelId int64) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		logger.Error(err)
		return
	}
	outReport := map[string]interface{}{
		"code": "OUT",
		"name": user.Nickname,
	}

	roomId, err := ch.db.GetUserRoomId(channelId)
	if err != nil {
		logger.Error(err)
		return
	}

	roomUserList, err := ch.db.GetRoomUser(roomId)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, user := range roomUserList {
		if user.ChannelId != channelId {
			actx := ch.ctxData[user.ChannelId]
			actx.Write(encodeMsg(ch.desKey[user.ChannelId], outReport))
		}
	}

	ch.db.OutRoom(roomId, user.UserId)
	delete(ch.ctxData, channelId)
	delete(ch.desKey, channelId)
}

func (ch *ChatManager) getOnlineList(roomId string) (onlineList []string) {
	userList, err := ch.db.GetRoomUser(roomId)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, user := range userList {
		onlineList = append(onlineList, user.Nickname)
	}

	return
}

func (ch *ChatManager) changeRoom(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()
	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		ctx.Close(err)
		return
	}

	srcRoomId, _ := ch.db.GetUserRoomId(channelId)
	dstRoomId, ok := message["roomId"].(string)
	if !ok {
		logger.Errorf("changeRoom message error")
		return
	}

	srcRoom := ch.db.GetRoom(srcRoomId)
	if srcRoom != nil {
		ch.db.OutRoom(srcRoomId, user.UserId)
	}

	dstRoom := ch.db.GetRoom(dstRoomId)
	if dstRoom == nil {
		dstRoom = &RoomModel{
			RoomId:     dstRoomId,
			RoomName:   "",
			CreateTime: time.Now(),
		}
		ch.db.CreateRoom(dstRoom)
	}
	ch.db.JoinRoom(dstRoomId, user.UserId)

	rnmMsg := map[string]interface{}{
		"code":       "RNM",
		"roomId":     dstRoomId,
		"nickname":   user.Nickname,
		"onlineList": ch.getOnlineList(dstRoomId),
	}
	ctx.Write(encodeMsg(ch.desKey[channelId], rnmMsg))

	jnrReport := map[string]interface{}{
		"code": "JNR",
		"name": user.Nickname,
	}
	outReport := map[string]interface{}{
		"code": "OUT",
		"name": user.Nickname,
	}

	// 退出房间通知
	srcUserList, _ := ch.db.GetRoomUser(srcRoomId)
	for _, u := range srcUserList {
		if actx := ch.ctxData[u.ChannelId]; actx != nil {
			actx.Write(encodeMsg(ch.desKey[u.ChannelId], outReport))
		}
	}
	// 加入房间通知
	dstUserList, _ := ch.db.GetRoomUser(dstRoomId)
	for _, u := range dstUserList {
		if u.ChannelId != channelId {
			if actx := ch.ctxData[u.ChannelId]; actx != nil {
				actx.Write(encodeMsg(ch.desKey[u.ChannelId], jnrReport))
			}
		}
	}

	// 推送新房间历史消息
	ch.pushHistoryMessage(dstRoomId, channelId)
}

func (ch *ChatManager) clearRoomMessage(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()
	roomId, err := ch.db.GetUserRoomId(channelId)
	if err != nil {
		logger.Error(err)
		ctx.Close(err)
		return
	}

	ch.db.ClearRoomMessage(roomId)
}

func (ch *ChatManager) changeNickname(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	name, ok := message["name"].(string)
	if !ok {
		logger.Errorf("changeNickname message error")
		return
	}

	user, _ := ch.db.GetUserByChannelId(ctx.Channel().ID())
	user.Nickname = name

	ch.db.UpdateUser(user, "nickname")
}

// 检查心跳是否过期
// func (ch *ChatManager) checkLiveTimeout(roomId string) {
// 	room := ch.db.GetRoom(roomId)
// 	if room == nil {
// 		return
// 	}
//
// 	var outList []string
// 	now := time.Now()
// 	for id := range room {
// 		user, exist := ch.userData[id]
// 		if !exist {
// 			continue
// 		}
// 		if now.Sub(user.LastLiveTime) > LiveTimeout {
// 			outList = append(outList, user.Nickname)
// 			delete(ch.userData, id)
// 			delete(room, id)
// 		}
// 	}
// 	for _, name := range outList {
// 		outReport := map[string]interface{}{
// 			"code": "OUT",
// 			"name": name,
// 		}
// 		for _, actx := range room {
// 			actx.Write(encodeMsg(outReport))
// 		}
// 	}
//
// }

func (ch *ChatManager) message(ctx netty.InboundContext, message map[string]interface{}) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()

	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		logger.Errorf("channel id %d not exist", channelId)
		return
	}
	roomId, _ := ch.db.GetUserRoomId(channelId)

	msgContent, ok := message["message"]
	if !ok {
		logger.Errorf("channel id %d message content not found", channelId)
		return
	}
	msg := map[string]interface{}{
		"code":     "MSG",
		"from":     user.Nickname,
		"message":  msgContent,
		"sendTime": time.Now().Format("15:04:05"),
	}
	userList, _ := ch.db.GetRoomUser(roomId)
	for _, u := range userList {
		if actx := ch.ctxData[u.ChannelId]; actx != nil {
			if u.ChannelId != channelId {
				actx.Write(encodeMsg(ch.desKey[u.ChannelId], msg))
			}
		}
	}

	if err = ch.db.NewMessage(&MessageModel{
		RoomId:     roomId,
		UserId:     user.UserId,
		Content:    fmt.Sprintf("%v", msgContent),
		CreateTime: time.Now(),
	}); err != nil {
		logger.Error(err)
	}
}

func (ch *ChatManager) offline(roomId string, channelId int64) {
	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		logger.Errorf("channel id %d not exist", channelId)
		return
	}

	ch.db.OutRoom(roomId, user.UserId)

	outReport := map[string]interface{}{
		"code": "OUT",
		"name": user.Nickname,
	}

	userList, _ := ch.db.GetRoomUser(roomId)

	for _, c := range userList {
		if actx := ch.ctxData[c.ChannelId]; actx != nil {
			actx.Write(encodeMsg(ch.desKey[c.ChannelId], outReport))
		}
	}
}

func (ch *ChatManager) live(ctx netty.InboundContext, msg map[string]interface{}) {
	ch.lock.RLock()
	defer ch.lock.RUnlock()
	ch.dataInit()

	channelId := ctx.Channel().ID()
	user, err := ch.db.GetUserByChannelId(channelId)
	if err != nil {
		logger.Errorf("channel id %d message content not found", channelId)
		return
	}
	now := time.Now()
	user.LastLiveTime = now

	ch.db.UpdateUser(user, "last_live_time")
}
