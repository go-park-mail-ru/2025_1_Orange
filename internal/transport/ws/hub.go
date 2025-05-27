package ws

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/usecase"
	l "ResuMatch/pkg/logger"
	"context"
	"sync"
)

type Hub struct {
	clients    map[ConnectionKey]*Client
	register   chan *Client
	unregister chan *Client
	Broadcast  chan Message
	mu         sync.Mutex
	chatUC     usecase.Chat
}

func NewHub(chat usecase.Chat) *Hub {
	return &Hub{
		clients:    make(map[ConnectionKey]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Broadcast:  make(chan Message),
		chatUC:     chat,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.Key] = client
			h.mu.Unlock()
			l.Log.Infof("Client connected: %+v", client.Key)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.Key]; ok {
				close(client.send)
				delete(h.clients, client.Key)
				l.Log.Infof("Client disconnected: %+v", client.Key)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.Lock()
			switch msgType := message.Type; msgType {
			case MessageTypeChat:
				req := message.Payload.(dto.MessageRequest)

				resp, err := h.chatUC.SendMessage(context.Background(), req.ChatID, req.SenderID, string(req.SenderRole), req.Payload)
				if err != nil {
					l.Log.Warnf("Не удалось сохранить сообщение: %v", err)
					h.mu.Unlock()
					continue
				}

				receiverKey := ConnectionKey{
					UserID: resp.ReceiverID,
					Type:   h.getReceiverRole(resp.FromApplicant),
				}

				if receiver, ok := h.clients[receiverKey]; ok {
					receiver.send <- Message{
						Type:    MessageTypeChat,
						Payload: resp,
					}
				}

				senderKey := ConnectionKey{
					UserID: req.SenderID,
					Type:   req.SenderRole,
				}

				if sender, ok := h.clients[senderKey]; ok {
					sender.send <- Message{
						Type:    MessageTypeChat,
						Payload: resp,
					}
				}
			case MessageTypeNotification:
				notificationMsg := message.Payload.(*entity.NotificationPreview)

				var receiverRole entity.UserRole
				switch notificationMsg.Type {
				case entity.DownloadResumeType:
					receiverRole = entity.ApplicantRole
				case entity.ApplyNotificationType:
					receiverRole = entity.EmployerRole
				}

				key := ConnectionKey{
					UserID: notificationMsg.ReceiverID,
					Type:   receiverRole,
				}
				if client, ok := h.clients[key]; ok {
					client.send <- message
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) getReceiverRole(fromApplicant bool) entity.UserRole {
	if fromApplicant {
		return entity.EmployerRole
	}
	return entity.ApplicantRole
}
