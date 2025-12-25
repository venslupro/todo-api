package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/venslupro/todo-api/internal/domain"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	Conn    *websocket.Conn
	UserID  string
	Send    chan WebSocketMessage
	TeamIDs []string // Teams the user is part of
}

// WebSocketService manages WebSocket connections and message broadcasting
type WebSocketService struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan WebSocketMessage),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Run starts the WebSocket service
func (s *WebSocketService) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			log.Printf("Client registered: %s", client.UserID)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.Send)
			}
			s.mu.Unlock()
			log.Printf("Client unregistered: %s", client.UserID)

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				// Send message to relevant clients based on user ID or team membership
				if s.shouldSendToClient(client, message) {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(s.clients, client)
					}
				}
			}
			s.mu.RUnlock()
		}
	}
}

// shouldSendToClient determines if a message should be sent to a specific client
func (s *WebSocketService) shouldSendToClient(client *WebSocketClient, message WebSocketMessage) bool {
	// Broadcast to all clients if no specific user/team targeting
	if message.UserID == "" {
		return true
	}

	// Send to specific user
	if message.UserID == client.UserID {
		return true
	}

	// For team-related messages, check if client is in the relevant team
	if teamMessage, ok := message.Payload.(map[string]interface{}); ok {
		if teamID, exists := teamMessage["team_id"]; exists {
			for _, clientTeamID := range client.TeamIDs {
				if clientTeamID == teamID {
					return true
				}
			}
		}
	}

	return false
}

// BroadcastTODOUpdate sends a TODO update to relevant clients
func (s *WebSocketService) BroadcastTODOUpdate(ctx context.Context, todo *domain.TODO, action string) {
	message := WebSocketMessage{
		Type: "todo_update",
		Payload: map[string]interface{}{
			"action":  action,
			"todo":    todo,
			"todo_id": todo.ID,
		},
		Timestamp: time.Now(),
	}

	// If TODO is in a team, send to all team members
	if todo.TeamID != nil && *todo.TeamID != "" {
		message.Payload.(map[string]interface{})["team_id"] = *todo.TeamID
	}

	s.broadcast <- message
}

// BroadcastTeamUpdate sends a team update to team members
func (s *WebSocketService) BroadcastTeamUpdate(ctx context.Context, team *domain.Team, action string) {
	message := WebSocketMessage{
		Type: "team_update",
		Payload: map[string]interface{}{
			"action":  action,
			"team":    team,
			"team_id": team.ID,
		},
		Timestamp: time.Now(),
	}

	s.broadcast <- message
}

// BroadcastUserNotification sends a notification to a specific user
func (s *WebSocketService) BroadcastUserNotification(ctx context.Context, userID, messageType, content string) {
	message := WebSocketMessage{
		Type: "notification",
		Payload: map[string]interface{}{
			"type":    messageType,
			"content": content,
		},
		UserID:    userID,
		Timestamp: time.Now(),
	}

	s.broadcast <- message
}

// WebSocketUpgrader upgrades HTTP connections to WebSocket connections
var WebSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin
		return true
	},
}

// HandleWebSocketConnection handles incoming WebSocket connections
func (s *WebSocketService) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, userID string, teamIDs []string) error {
	conn, err := WebSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	client := &WebSocketClient{
		Conn:    conn,
		UserID:  userID,
		Send:    make(chan WebSocketMessage, 256),
		TeamIDs: teamIDs,
	}

	s.register <- client

	// Start goroutines for reading and writing
	go s.writePump(client)
	go s.readPump(client)

	return nil
}

// writePump sends messages to the WebSocket client
func (s *WebSocketService) writePump(client *WebSocketClient) {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				// Channel closed
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message as JSON
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			err = client.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			err := client.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

// readPump reads messages from the WebSocket client
func (s *WebSocketService) readPump(client *WebSocketClient) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512) // 512 bytes max message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Process message based on type
		s.handleIncomingMessage(client, msg)
	}
}

// handleIncomingMessage processes incoming WebSocket messages
func (s *WebSocketService) handleIncomingMessage(client *WebSocketClient, message WebSocketMessage) {
	switch message.Type {
	case "ping":
		// Respond to ping
		response := WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		client.Send <- response

	case "subscribe":
		// Handle subscription requests
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if teamID, exists := payload["team_id"]; exists {
				// Add team to client's subscription
				if teamIDStr, ok := teamID.(string); ok {
					client.TeamIDs = append(client.TeamIDs, teamIDStr)
				}
			}
		}

	case "unsubscribe":
		// Handle unsubscription requests
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if teamID, exists := payload["team_id"]; exists {
				// Remove team from client's subscription
				if teamIDStr, ok := teamID.(string); ok {
					for i, id := range client.TeamIDs {
						if id == teamIDStr {
							client.TeamIDs = append(client.TeamIDs[:i], client.TeamIDs[i+1:]...)
							break
						}
					}
				}
			}
		}

	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}
