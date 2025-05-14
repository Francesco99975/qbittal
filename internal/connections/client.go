package connections

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
)

type Client struct {
	id string

	ip string

	socket *websocket.Conn

	egress chan Event

	manager *ConnectionManager

	room string

	sauce string

	agent string
}

func (client *Client) read() {
	defer func() {
		client.manager.disconnect <- client
	}()

	client.socket.SetReadLimit(messageBufferSize)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := client.socket.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Error(err)
		return
	}

	client.socket.SetPongHandler(client.pongHandler)

	for {
		_, payload, err := client.socket.ReadMessage()
		log.Debugf("payload: %v", string(payload))
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error reading message: %v", err)
			}
			return
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Errorf("error marshalling message: %v", err)
			return // Breaking the connection here might be harsh xD
		}

		log.Debugf("event received: %v", request)

		if err := client.manager.routeEvent(request, client); err != nil {
			log.Errorf("Error handeling Message: ", err)
		}
	}
}

func (client *Client) write() {
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		client.manager.disconnect <- client
	}()

	for {
		select {
		case message, ok := <-client.egress:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := client.socket.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Infof("connection closed: %v", err)
				}
				// Return to close the goroutine
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Error(err)
				return // closes the connection, should we really
			}
			// Write a Regular text message to the connection
			if err := client.socket.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Error(err)
			}
			log.Debug("sent message")
		case <-ticker.C:
			log.Debug("Ping")
			// Send the Ping
			if err := client.socket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Errorf("writemsg: ", err)
				return // return to break this goroutine triggeing cleanup
			}
		}
	}
}

// pongHandler is used to handle PongMessages for the Client
func (client *Client) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	log.Debug("pong")
	return client.socket.SetReadDeadline(time.Now().Add(pongWait))
}
