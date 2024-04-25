package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	// put registered clients into the room.
	rooms map[*connection]struct{}
	// Inbound messages from the clients.
	broadcast chan message

	// Register requests from the clients.
	register chan subscription

	// Unregister requests from clients.
	unregister chan subscription
}

type message struct {
	Event     string    `json:"event"`
	Room      string    `json:"room"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy int       `json:"createdBy"`
}

func (m message) String() string {
	user, ok := FindUserByID(m.CreatedBy)
	fmt.Println(ok)
	return fmt.Sprintf("[%v] [%s] %s: %s\n", m.CreatedAt.Format(time.ANSIC), m.Room, user.Name, string(m.Data))
}

var H = &Hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[*connection]struct{}),
}

func NewSession(userID int) (*session, int) {
	sessionID := rand.Int()
	return &session{UserID: userID}, sessionID
}

func (h *Hub) Run() {
	for {
		select {
		case s := <-h.register:
			_, ok := h.rooms[s.conn]
			if !ok {
				// create a new room
				h.rooms[s.conn] = struct{}{}
			}
		case s := <-h.unregister:
			_, ok := h.rooms[s.conn]
			if !ok {
				delete(h.rooms, s.conn)
				close(s.conn.send)
			}
		case m := <-h.broadcast:
			for c := range h.rooms {
				byteMessage, _ := json.Marshal(m)
				select {
				case c.send <- byteMessage:
				default:
					close(c.send)
					delete(h.rooms, c)
				}
			}
		}
	}
}
