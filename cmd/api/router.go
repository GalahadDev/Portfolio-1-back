package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/tu-usuario/route-manager/internal/config"
	appMiddleware "github.com/tu-usuario/route-manager/internal/middleware"
	"github.com/tu-usuario/route-manager/internal/route"
	"github.com/tu-usuario/route-manager/internal/user"
)

// NewRouter recibe la configuración
func NewRouter(
	cfg *config.Config,
	userHandler *user.Handler,
	routeHandler *route.Handler,
) *chi.Mux {

	r := chi.NewRouter()

	// ---------------------------------------------------------
	// 1. Middlewares Globales (Infraestructura)
	// ---------------------------------------------------------
	r.Use(middleware.Logger)    // Log de cada petición
	r.Use(middleware.Recoverer) // Evita que el server muera por pánico

	// Configuración CORS para Next.js
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ---------------------------------------------------------
	// 2. Rutas Públicas
	// ---------------------------------------------------------
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Route Manager API: Online 🚀"))
	})

	// ---------------------------------------------------------
	// 3. Rutas Protegidas (API v1)
	// ---------------------------------------------------------
	r.Route("/api/v1", func(r chi.Router) {

		// Pasamos la URL de Supabase para el JWKS
		r.Use(appMiddleware.AuthMiddleware(cfg.SupabaseURL))

		// --- Endpoint de prueba de seguridad ---
		r.Get("/check-auth", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := appMiddleware.GetUserID(r.Context())
			w.Write([]byte("✅ Autenticado correctamente. Tu ID es: " + userID))
		})

		// --- Módulo de Usuarios ---
		userHandler.RegisterRoutes(r)

		// --- Módulo de Rutas  ---
		r.Route("/routes", func(r chi.Router) {
			routeHandler.RegisterRoutes(r)
		})
	})

	return r
}
