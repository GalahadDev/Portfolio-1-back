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
	r.Get("/my-routes", h.GetMyRoutes)
	r.Patch("/{id}/status", h.UpdateStatus)
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

func (h *Handler) GetMyRoutes(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener ID del conductor desde el Token (Contexto)
	driverID, err := appMiddleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	// 2. Llamar al servicio
	routes, err := h.service.GetMyRoutes(r.Context(), driverID)
	if err != nil {
		http.Error(w, "Error obteniendo rutas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Responder JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

type UpdateRouteStatusInput struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener ID de la ruta desde la URL
	routeID := chi.URLParam(r, "id")
	if routeID == "" {
		http.Error(w, "ID de ruta requerido", http.StatusBadRequest)
		return
	}

	// 2. Decodificar el JSON (ej: { "status": "in_progress" })
	var input UpdateRouteStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Llamar al servicio
	if err := h.service.UpdateRouteStatus(r.Context(), routeID, input.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Responder OK
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Estado actualizado correctamente"})
}
