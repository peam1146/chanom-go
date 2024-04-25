package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type subscription struct {
	conn   *connection
	sender int
}

type MessagePayload struct {
	Event     string `json:"event"`
	Data      string `json:"data"`
	RoomID    string `json:"roomID"`
	CreatedBy int    `json:"createdBy"`
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
}

func (s *subscription) readPump() {
	c := s.conn
	defer func() {
		// Unregister
		H.unregister <- *s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// Reading incoming message...
		_, payload, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		payload = bytes.TrimSpace(bytes.Replace(payload, newline, space, -1))

		var msgPayload MessagePayload
		if err := json.Unmarshal(payload, &msgPayload); err != nil {
			log.Println(err)
			continue
		}

		m := message{
			Event:     msgPayload.Event,
			Data:      msgPayload.Data,
			Room:      msgPayload.RoomID,
			CreatedAt: time.Now(),
			CreatedBy: s.sender,
		}
		fmt.Println(m)
		H.broadcast <- m
	}
}

func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		// Listerning message when it comes will write it into writer and then send it to the client
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	urlQuery := r.URL.Query()
	sessionID, err := strconv.Atoi(urlQuery.Get("session_id"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	userID, ok := FindUserBySessionID(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	// Get room's id from client...
	if err != nil {
		log.Println(err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{
		conn:   c,
		sender: userID,
	}
	H.register <- s
	go s.writePump()
	go s.readPump()
}
