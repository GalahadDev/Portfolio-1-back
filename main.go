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

		// ========== NIVEL 1: AUTENTICACIÓN (Tener Token Válido) ==========
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.SupabaseURL))
		{
			// A. REGISTRO: Accesible aunque el status sea 'inactive'
			protected.POST("/auth/register", auth.RegisterUserFromGoogle)

			// ========== NIVEL 2: AUTORIZACIÓN (Tener Status 'active') ==========
			// Todo lo que esté dentro de 'activeUsers' requiere que el admin te haya aprobado
			activeUsers := protected.Group("/")
			activeUsers.Use(middleware.RequireActiveUser())
			{
				// --- USUARIOS ---
				activeUsers.GET("/users/me", users.GetMe)

				// Rutas SOLO ADMINS
				adminOnly := activeUsers.Group("/users")
				adminOnly.Use(middleware.RequireRoles("admin"))
				{
					adminOnly.GET("", users.ListUsers)         // Listar todos
					adminOnly.GET("/:id", users.GetUser)       // Ver otro usuario específico
					adminOnly.PUT("/:id", users.UpdateUser)    // Editar otro usuario (Aquí se activan cuentas)
					adminOnly.DELETE("/:id", users.DeleteUser) // Borrar usuario
				}

				// --- RUTAS ---
				routesGroup := activeUsers.Group("/routes")
				{
					// Crear (Admin)
					routesGroup.POST("", middleware.RequireRoles("admin"), routes.CreateRoute)

					// Listar (Admin y Conductor)
					// NOTA: Quité el middleware "admin" aquí para que los conductores puedan ver SU lista
					routesGroup.GET("", routes.ListRoutes)

					// Ver Detalle
					routesGroup.GET("/:id", routes.GetRouteByID)

					// Editar (Admin)
					routesGroup.PUT("/:id", middleware.RequireRoles("admin"), routes.UpdateRoute)

					// Eliminar (Admin)
					routesGroup.DELETE("/:id", middleware.RequireRoles("admin"), routes.DeleteRoute)

					// Operaciones
					routesGroup.PATCH("/:id/assign", middleware.RequireRoles("admin"), routes.AssignDriver)
					routesGroup.PATCH("/:id/status", routes.UpdateRouteStatus)
				}

				// --- WAYPOINTS (PUNTOS DE ENTREGA) ---
				waypointsGroup := activeUsers.Group("/waypoints")
				{
					// Completar entrega
					waypointsGroup.PATCH("/:id/complete", waypoints.MarkWaypointComplete)

					// Editar dirección (Admin)
					waypointsGroup.PUT("/:id", middleware.RequireRoles("admin"), waypoints.UpdateWaypoint)
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
