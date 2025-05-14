package connections

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
)

type ConnectionManager struct {
	clients    map[*Client]bool
	connect    chan *Client
	disconnect chan *Client
	handlers   map[string]EventHandler
	// otps is a map of allowed OTP to accept connections from
	otps RetentionMap
}

func (cm *ConnectionManager) GenerateNewOtp() string {
	return cm.otps.NewOTP().Key
}

func NewManager(ctx context.Context) *ConnectionManager {
	cm := &ConnectionManager{
		connect:    make(chan *Client),
		disconnect: make(chan *Client),
		clients:    make(map[*Client]bool),
		handlers:   make(map[string]EventHandler),
		otps:       NewRetentionMap(ctx, 5*time.Second),
	}

	cm.setupEventHandlers()

	return cm
}

// setupEventHandlers configures and adds all handlers
func (m *ConnectionManager) setupEventHandlers() {
	// m.handlers[EventVisit] = SendVisitHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *ConnectionManager) routeEvent(event Event, c *Client) error {
	// Check if Handler is present in Map
	if handler, ok := m.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("this event type is not supported")
	}
}

func (cm *ConnectionManager) BroadcastEvent(event Event) {
	log.Debugf("Broadcasting event: %v", event)
	for client := range cm.clients {
		client.egress <- event
	}
}

func (cm *ConnectionManager) Run() {
	for {
		select {
		case client := <-cm.connect:
			cm.clients[client] = true
		case client := <-cm.disconnect:
			if _, ok := cm.clients[client]; ok {

				// for cl := range client.manager.clients {
				// 	// Only send to clients inside the same chatroom
				// 	if cl.room == "admin" {
				// 		cl.egress <- Event{Type: EventUpdateVisitsAdmin, Payload: []byte("{}")}
				// 	}

				// }

				close(client.egress)
				// analizer.archiveVisit(client.id)
				client.socket.Close()
				delete(cm.clients, client)
			}
		}
	}
}

func (cm *ConnectionManager) ServeWS(c echo.Context) error {
	socket, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Fatal("Serve HTTP Sockets Error: ", err)
		return err
	}

	log.Info("Connection Received")

	client := &Client{
		id:      uuid.NewV4().String(),
		ip:      c.Request().Header.Get("X-Forwarded-For"),
		socket:  socket,
		egress:  make(chan Event, messageBufferSize),
		manager: cm,
		room:    "base",
		sauce:   c.Request().Header.Get("Referer"),
		agent:   c.Request().Header.Get("User-Agent"),
	}

	log.Infof("Client: %v", client)

	cm.connect <- client

	go client.read()

	go client.write()

	return nil
}
