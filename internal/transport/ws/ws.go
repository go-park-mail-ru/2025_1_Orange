package ws

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"net/http"
)

type WebsocketHandler struct {
	auth  usecase.Auth
	wsHub *Hub
}

func NewWebsocketHandler(auth usecase.Auth, wsHub *Hub) *WebsocketHandler {
	return &WebsocketHandler{auth: auth, wsHub: wsHub}
}

func (h *WebsocketHandler) Configure(r *http.ServeMux) {
	wsMux := http.NewServeMux()

	wsMux.HandleFunc("GET /connect", h.HandleWebsocket)

	r.Handle("/ws/", http.StripPrefix("/ws", wsMux))
}

func (h *WebsocketHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}
	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}
	ServeWs(h.wsHub, w, r, userID, role)
}

//
//import (
//	"ResuMatch/internal/entity"
//	"ResuMatch/internal/transport/http/utils"
//	"ResuMatch/internal/usecase"
//	l "ResuMatch/pkg/logger"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/gorilla/websocket"
//	"net/http"
//	"sync"
//	"time"
//)
//
//type WebsocketPool struct {
//	connections map[ConnectionKey][]*websocket.Conn
//	upgrader    websocket.Upgrader
//	mu          *sync.Mutex
//	auth        usecase.Auth
//}
//
//func NewWebsocketPool(authUC usecase.Auth) *WebsocketPool {
//	return &WebsocketPool{
//		connections: make(map[ConnectionKey][]*websocket.Conn),
//		upgrader: websocket.Upgrader{
//			ReadBufferSize:  1024,
//			WriteBufferSize: 1024,
//			CheckOrigin: func(r *http.Request) bool {
//				return true
//			},
//		},
//		mu:   new(sync.Mutex),
//		auth: authUC,
//	}
//}
//
//func (wsp *WebsocketPool) Connect(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//
//	cookie, err := r.Cookie("session_id")
//	if err != nil {
//		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
//		return
//	}
//	userID, role, err := wsp.auth.GetUserIDBySession(ctx, cookie.Value)
//	if err != nil {
//		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
//		return
//	}
//
//	conn, err := wsp.upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
//		l.Log.Errorf("websocket upgrade failed: %v", err)
//		return
//	}
//
//	var userRole entity.UserRole
//	switch role {
//	case "employer":
//		userRole = entity.EmployerRole
//	case "applicant":
//		userRole = entity.ApplicantRole
//	default:
//		http.Error(w, "invalid user role", http.StatusForbidden)
//		return
//	}
//
//	userKey := ConnectionKey{UserID: userID, Type: userRole}
//	defer wsp.cleanupConnection(userKey, conn)
//
//	wsp.AddConn(userKey, conn)
//	wsp.handleConnection(conn)
//}
//
//func (wsp *WebsocketPool) cleanupConnection(key ConnectionKey, conn *websocket.Conn) {
//	if err := conn.Close(); err != nil {
//		l.Log.Errorf("connection close error: %v", err)
//	}
//	if err := wsp.RemoveConn(key, conn); err != nil {
//		l.Log.Errorf("connection delete error: %v", err)
//	}
//}
//
//func (wsp *WebsocketPool) handleConnection(conn *websocket.Conn) {
//	pingInterval := 30 * time.Second
//	pongWait := 60 * time.Second
//
//	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
//		l.Log.Errorf("initial set read deadline error: %v", err)
//	}
//
//	conn.SetPongHandler(func(appData string) error {
//		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
//			l.Log.Errorf("set read deadline (pong) error: %v", err)
//		}
//		return nil
//	})
//
//	ticker := time.NewTicker(pingInterval)
//	defer ticker.Stop()
//
//	go func() {
//		for range ticker.C {
//			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
//				l.Log.Warnf("ping failed: %v", err)
//				if err = conn.Close(); err != nil {
//					l.Log.Errorf("websocket close error: %v", err)
//				}
//				return
//			}
//		}
//	}()
//
//	for {
//		_, _, err := conn.ReadMessage()
//		if err != nil {
//			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
//				l.Log.Errorf("read error: %v", err)
//			}
//			break
//		}
//	}
//}
//
//func (wsp *WebsocketPool) AddConn(key ConnectionKey, conn *websocket.Conn) {
//	wsp.mu.Lock()
//	defer wsp.mu.Unlock()
//	wsp.connections[key] = append(wsp.connections[key], conn)
//	l.Log.Infof("websocket connection added for notification key: %v", key)
//}
//
//func (wsp *WebsocketPool) RemoveConn(key ConnectionKey, conn *websocket.Conn) error {
//	wsp.mu.Lock()
//	defer wsp.mu.Unlock()
//
//	conns, ok := wsp.connections[key]
//	if !ok {
//		l.Log.Warnf("websocket connection not found for notification key: %s %d %v", key.Type, key.UserID, key)
//		return errors.New("connection not found")
//	}
//
//	for i, c := range conns {
//		if c == conn {
//			conns[i] = conns[len(conns)-1]
//			wsp.connections[key] = conns[:len(conns)-1]
//
//			if len(wsp.connections[key]) == 0 {
//				delete(wsp.connections, key)
//			}
//			return nil
//		}
//	}
//	return nil
//}
//
//func (wsp *WebsocketPool) send(key ConnectionKey, data []byte) error {
//	wsp.mu.Lock()
//	defer wsp.mu.Unlock()
//
//	conns, ok := wsp.connections[key]
//	if !ok {
//		l.Log.Warnf("websocket connection not found for notification key: %v", key)
//		return errors.New("connection not found")
//	}
//
//	for _, conn := range conns {
//		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
//			l.Log.Errorf("Send error to %v: %v", key, err)
//			return err
//		}
//	}
//	return nil
//}
//
//func (wsp *WebsocketPool) SendNotification(notification *entity.NotificationPreview) error {
//	if notification == nil {
//		return errors.New("nil notification")
//	}
//
//	data, err := json.Marshal(notification)
//	if err != nil {
//		return fmt.Errorf("marshal error: %w", err)
//	}
//
//	var receiverRole entity.UserRole
//	switch notification.Type {
//	case entity.DownloadResumeType:
//		receiverRole = entity.ApplicantRole
//	case entity.ApplyNotificationType:
//		receiverRole = entity.EmployerRole
//	}
//
//	receiver := ConnectionKey{
//		UserID: notification.ReceiverID,
//		Type:   receiverRole,
//	}
//
//	return wsp.send(receiver, data)
//}
