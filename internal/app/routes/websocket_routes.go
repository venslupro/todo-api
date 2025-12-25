package routes

import (
	"net/http"

	"github.com/venslupro/todo-api/internal/app/handlers"
)

// RegisterWebSocketRoutes registers WebSocket routes
func RegisterWebSocketRoutes(mux *http.ServeMux, websocketHandler *handlers.WebSocketHandler) {
	mux.HandleFunc("/ws", websocketHandler.HandleWebSocket)
	mux.HandleFunc("/websocket", websocketHandler.HandleWebSocket)
}
