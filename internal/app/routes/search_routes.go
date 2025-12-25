package routes

import (
	"net/http"

	"github.com/venslupro/todo-api/internal/app/handlers"
)

// RegisterSearchRoutes registers advanced search routes
func RegisterSearchRoutes(mux *http.ServeMux, todoHandler *handlers.TODOHandler) {
	mux.HandleFunc("/v1/todos/search", todoHandler.HandleAdvancedSearch)
}
