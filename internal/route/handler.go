package route

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appMiddleware "github.com/tu-usuario/route-manager/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener quién está creando la ruta
	creatorID, err := appMiddleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	// 2. Decodificar JSON
	var input CreateRouteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Llamar al servicio
	route, err := h.service.CreateRoute(r.Context(), creatorID, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Responder
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(route)
}
