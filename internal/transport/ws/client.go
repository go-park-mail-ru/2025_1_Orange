package ws

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	l "ResuMatch/pkg/logger"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Message
	Key  ConnectionKey
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID int, role string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	var userRole entity.UserRole
	switch role {
	case "employer":
		userRole = entity.EmployerRole
	case "applicant":
		userRole = entity.ApplicantRole
	default:
		http.Error(w, "неверная роль пользователя", http.StatusForbidden)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan Message, 256),
		Key: ConnectionKey{
			UserID: userID,
			Type:   userRole,
		},
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения: %v", err)
		}
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		l.Log.Errorf("initial set read deadline error: %v", err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			l.Log.Errorf("set read deadline (pong) error: %v", err)
		}
		return nil
	})

	for {
		var msg struct {
			Type    MessageType `json:"type"`
			ChatID  int         `json:"chat_id"`
			Payload string      `json:"payload"`
		}

		l.Log.Infof("Чтение сообщения: %v", msg)
		if err := c.conn.ReadJSON(&msg); err != nil {
			l.Log.Error("Ошибка при чтении сообщения:", err)
			break
		}

		if msg.Type == MessageTypeChat {
			newMsg := Message{
				Type: MessageTypeChat,
				Payload: dto.MessageRequest{
					ChatID:     msg.ChatID,
					SenderID:   c.Key.UserID,
					SenderRole: c.Key.Type,
					Payload:    msg.Payload,
				},
			}
			c.hub.Broadcast <- newMsg
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения: %v", err)
		}
	}()

	for {
		select {
		case msg, ok := <-c.send:
			l.Log.Infof("Отправка сообщения: payload=%v, type=%v", msg.Payload, msg.Type)

			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				l.Log.Errorf("Ошибка при установке write deadline: %v", err)
				return
			}

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					l.Log.Errorf("Ошибка при отправке CloseMessage: %v", err)
				}
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				l.Log.Error("Ошибка при отправке сообщения:", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				l.Log.Errorf("Ошибка при установке write deadline перед ping: %v", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				l.Log.Error("Ошибка при отправке ping:", err)
				return
			}
		}
	}
}
