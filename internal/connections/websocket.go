package connections

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 4096
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

func checkOrigin(r *http.Request) bool {

	return true

	// // Grab the request origin
	// origin := r.Header.Get("Origin")

	// switch origin {
	// // Update this to HTTPS
	// case os.Getenv("HOST"):
	// 	return true
	// default:
	// 	log.Error(fmt.Sprintf("Origin %s not allowed", origin))
	// 	return false
	// }
}

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin, ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}
