package main

import (
	"flag"
	"fmt"
	mathRand "math/rand"
	"net/http"
	"time"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-transport/websocket"
	logger "github.com/opentrx/seata-golang/v2/pkg/util/log"
)

const (
	LiveTimeout = 10 * time.Second // 在线超时，超出这个时间没有收到心跳包，则认为已掉线
)

func init() {
	mathRand.Seed(time.Now().UnixNano())
}

var ManagerInst = NewManager()
var ch = NewChatManager()

func main() {
	port := "5534"
	dbPath := "./chat.db"
	flag.StringVar(&port, "port", "5534", "-port xxx to set port")
	flag.StringVar(&dbPath, "db", "./chat.db", "-db xxx to set db path")
	flag.Parse()

	// setup websocket params.
	options := &websocket.Options{
		// Timeout:  time.Second * 5,
		ServeMux: http.NewServeMux(),
	}

	// index page.
	options.ServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write(indexHtml)
	})

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// Exceeding maxFrameLength will throw exception handling
			AddLast(PacketCodec(1024)).
			// decode to map[string]interface{}
			AddLast(StringCodec()).
			// session recorder.
			AddLast(ManagerInst).
			// chat handler.
			AddLast(chatHandler{})
	}

	logger.Infof("Server start listen on[%s]", port)
	// setup bootstrap & startup server.
	err := netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(websocket.New())).
		Listen(fmt.Sprintf("0.0.0.0:%s/chat", port), websocket.WithOptions(options)).Sync()
	if err != nil {
		logger.Error(err)
	}
}
