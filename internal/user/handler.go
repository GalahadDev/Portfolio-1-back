package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	appMiddleware "github.com/tu-usuario/route-manager/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes define las URLs del módulo
func (h *Handler) RegisterRoutes(r chi.Router) {

	// 1. Ruta de Auth
	r.Post("/auth/sync", h.SyncUser)

	// 2. Grupo de Usuarios
	r.Route("/users", func(r chi.Router) {

		// URL final: GET /api/v1/users/me
		r.Get("/me", h.GetMe)

		// URL final: GET /api/v1/users (Listar con filtros)
		r.Get("/", h.ListUsers)

		// URL final: PATCH /api/v1/users/{id}/approve
		r.Patch("/{id}/approve", h.ApproveUser)
	})
}

// Request Body para aprobar
type approveRequest struct {
	Role string `json:"role"` // ej: "driver", "fleet_admin"
}

// ApproveUser: PATCH /api/v1/users/{id}/approve
func (h *Handler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req approveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validar que el rol sea válido
	if req.Role == "" {
		req.Role = "driver" // Default
	}

	if err := h.service.ApproveUser(r.Context(), userID, req.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Usuario actualizado exitosamente"}`))
}

// ListUsers: GET /api/v1/users?status=pending
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	users, err := h.service.ListUsers(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetMe devuelve el perfil del usuario autenticado actualmente
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener ID del Token (gracias al Middleware)
	userIDStr, err := appMiddleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Usuario no identificado en el contexto", http.StatusUnauthorized)
		return
	}

	// 2. Buscar en BD
	user, err := h.service.repo.GetByID(r.Context(), userIDStr)
	if err != nil {
		// Si no lo encontramos, es un 404 (probablemente no hizo el Sync antes)
		http.Error(w, "Usuario no encontrado (¿Hiciste el Sync?)", http.StatusNotFound)
		return
	}

	// 3. Responder JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Request Body para la sincronización
type syncUserRequest struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

// SyncUser recibe los datos del usuario logueado en Next.js y los guarda en nuestra BD
func (h *Handler) SyncUser(w http.ResponseWriter, r *http.Request) {
	var req syncUserRequest

	// 1. Decodificar el JSON del cuerpo de la petición
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Validar ID (UUID)
	uid, err := uuid.Parse(req.ID)
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	// 3. Llamar al servicio
	user, err := h.service.SyncUser(r.Context(), uid, req.Email, req.FullName, req.AvatarURL)
	if err != nil {
		// En un caso real, no devolveríamos el error crudo por seguridad, pero para dev sirve
		http.Error(w, "Error al sincronizar usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Responder con el usuario creado/actualizado en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
