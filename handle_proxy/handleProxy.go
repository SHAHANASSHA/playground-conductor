package handle_proxy

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getDestination(path string, rdb *redis.Client, ctx context.Context) string {
	parts := strings.Split(path, "/")
	id := parts[1]
	vm := parts[2]
	ip := rdb.Get(ctx, id).Val()
	if ip == "" {
		return ""
	}
	return "ws://" + ip + ":8080/" + vm
}

func HandleProxy(w http.ResponseWriter, r *http.Request, rdb *redis.Client, ctx context.Context) {
	dest := getDestination(r.URL.Path, rdb, ctx)
	if dest == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Client upgrade error:", err)
		return
	}
	defer clientConn.Close()

	targetURL, _ := url.Parse(dest)
	targetConn, _, err := websocket.DefaultDialer.Dial(targetURL.String(), nil)
	if err != nil {
		log.Println("Dial target error:", err)
		return
	}
	defer targetConn.Close()

	errCh := make(chan error, 2)

	go proxyWebSocket(clientConn, targetConn, errCh)
	go proxyWebSocket(targetConn, clientConn, errCh)

	<-errCh
}

func proxyWebSocket(src, dest *websocket.Conn, errCh chan error) {
	for {
		mt, message, err := src.ReadMessage()
		if err != nil {
			errCh <- err
			break
		}
		err = dest.WriteMessage(mt, message)
		if err != nil {
			errCh <- err
			break
		}
	}
}
