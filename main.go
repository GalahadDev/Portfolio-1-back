package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/tu-usuario/route-manager/api/config"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/handlers/auth"
	"github.com/tu-usuario/route-manager/api/handlers/health"
	"github.com/tu-usuario/route-manager/api/handlers/routes"
	"github.com/tu-usuario/route-manager/api/handlers/users"
	"github.com/tu-usuario/route-manager/api/handlers/waypoints"
	"github.com/tu-usuario/route-manager/api/middleware"
)

func main() {
	// 1. Cargar Configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	// 2. Inicializar Base de Datos
	database.InitDB(cfg.DatabaseURL)

	// 3. Configurar Gin
	if os.Getenv("PORT") != "" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// 4. Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://mi-frontend.vercel.app", "http://127.0.0.1:5500"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 5. Definir Rutas
	api := router.Group("/api/v1")
	{
		api.GET("/health", health.HealthCheck) // health check del servidor

		// ========== NIVEL 1: AUTENTICACIÓN ==========
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.SupabaseURL))
		{
			// A. REGISTRO
			protected.POST("/auth/register", auth.RegisterUserFromGoogle)

			// B. UNIRSE A FLOTA
			protected.POST("/users/join-fleet", users.JoinFleet)

			// ========== NIVEL 2: AUTORIZACIÓN ==========

			activeUsers := protected.Group("/")
			activeUsers.Use(middleware.RequireActiveUser())
			{
				// --- USUARIOS ---
				activeUsers.GET("/users/me", users.GetMe)

				// Rutas SOLO ADMINS (y SUPER ADMINS)
				adminOnly := activeUsers.Group("/users")
				// Permitimos que el Super Admin también gestione usuarios
				adminOnly.Use(middleware.RequireRoles("admin", "super_admin"))
				{
					adminOnly.GET("", users.ListUsers)         // Listar todos (Filtrado por lógica de negocio)
					adminOnly.GET("/:id", users.GetUser)       // Ver otro usuario específico
					adminOnly.PUT("/:id", users.UpdateUser)    // Editar/Promover usuario
					adminOnly.DELETE("/:id", users.DeleteUser) // Borrar usuario
				}

				// --- RUTAS (ROUTES) ---
				routesGroup := activeUsers.Group("/routes")
				{
					// Crear (Admin/SuperAdmin)
					routesGroup.POST("", middleware.RequireRoles("admin", "super_admin"), routes.CreateRoute)

					// Listar (Admin y Conductor)
					// Sin middleware de rol: la lógica interna filtra "Mis Rutas" vs "Todas"
					routesGroup.GET("", routes.ListRoutes)

					// Ver Detalle
					routesGroup.GET("/:id", routes.GetRouteByID)

					// Editar (Admin/SuperAdmin)
					routesGroup.PUT("/:id", middleware.RequireRoles("admin", "super_admin"), routes.UpdateRoute)

					// Eliminar (Admin/SuperAdmin)
					routesGroup.DELETE("/:id", middleware.RequireRoles("admin", "super_admin"), routes.DeleteRoute)

					// Operaciones
					routesGroup.PATCH("/:id/assign", middleware.RequireRoles("admin", "super_admin"), routes.AssignDriver)
					routesGroup.PATCH("/:id/status", routes.UpdateRouteStatus)
				}

				// --- WAYPOINTS ---
				waypointsGroup := activeUsers.Group("/waypoints")
				{
					// Completar entrega (Conductor)
					waypointsGroup.PATCH("/:id/complete", waypoints.MarkWaypointComplete)

					// Editar dirección (Admin/SuperAdmin)
					waypointsGroup.PUT("/:id", middleware.RequireRoles("admin", "super_admin"), waypoints.UpdateWaypoint)
				}
			}
		}
	}

	// 6. Iniciar Servidor
	log.Printf("🚀 Route Manager API corriendo en puerto %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}
