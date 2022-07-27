package main

import (
	"encoding/json"
	"fmt"

	logger "github.com/opentrx/seata-golang/v2/pkg/util/log"

	"github.com/go-netty/go-netty"
)

type chatHandler struct {
}

func (chatHandler) HandleActive(ctx netty.ActiveContext) {
	fmt.Printf("child connection from: %s\n", ctx.Channel().RemoteAddr())
	ch.connectInit(ctx)
}

func (chatHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
	str, ok := message.(string)
	if !ok {
		logger.Error("read message error")
		return
	}
	if str == "" {
		return
	}
	channelId := ctx.Channel().ID()

	desKey := ""
	ch.lock.Lock()
	desKey = ch.desKey[channelId]
	ch.lock.Unlock()

	str = decode(desKey, str)

	cmd := make(map[string]interface{})
	if err := json.Unmarshal([]byte(str), &cmd); err != nil {
		logger.Errorf("read message error: %v", err)
		return
	}
	code, ok := cmd["code"]
	// logger.Infof("[%d] message from: %s code[%s]\n", ctx.Channel().ID(), ctx.Channel().RemoteAddr(), code)
	if !ok {
		logger.Errorf("read message error: ", str)
		return
	}
	switch code {
	case "KEY": // 协商des密钥
		ch.pushDesKey(ctx, cmd)
	case "LOGIN": // 登录
		ch.login(ctx, cmd)
	case "CHM": // 切换房间号
		logger.Infof("[%d] message from: %s code[%s]\n", ctx.Channel().ID(), ctx.Channel().RemoteAddr(), code)
		ch.changeRoom(ctx, cmd)
	case "CHN": // 修改昵称
		logger.Infof("[%d] message from: %s code[%s]\n", ctx.Channel().ID(), ctx.Channel().RemoteAddr(), code)
		ch.changeNickname(ctx, cmd)
	case "MSG": // 新消息
		ch.message(ctx, cmd)
	case "LIV": // 心跳包
		ch.live(ctx, cmd)
	case "OUT": // 退出
		logger.Infof("[%d] message from: %s code[%s]\n", ctx.Channel().ID(), ctx.Channel().RemoteAddr(), code)
		ch.outRoom(ctx.Channel().ID())
		ctx.Close(nil)
	case "CLEAR": // 清空房间消息
		ch.clearRoomMessage(ctx, cmd)
	}
}

func (*chatHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Printf("child connection closed: %s %s\n", ctx.Channel().RemoteAddr(), ex.Error())

	ch.outRoom(ctx.Channel().ID())
	ctx.HandleInactive(ex)
}
