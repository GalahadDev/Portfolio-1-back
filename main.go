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

	// 5. Definir Rutas (Endpoints)
	api := router.Group("/api/v1")
	{
		api.GET("/health", health.HealthCheck) // health chech del servidor

		// ========== ZONA AUTENTICADA (Cualquier rol) ==========
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.SupabaseURL))
		{
			protected.POST("/auth/register", auth.RegisterUserFromGoogle)

			// 1. Rutas Comunes (Conductores y Admins)
			protected.GET("/users/me", users.GetMe)

			// 2. Rutas SOLO ADMINS
			adminOnly := protected.Group("/users")
			adminOnly.Use(middleware.RequireRoles("admin"))
			{
				adminOnly.GET("", users.ListUsers)         // Listar todos
				adminOnly.GET("/:id", users.GetUser)       // Ver otro usuario específico
				adminOnly.PUT("/:id", users.UpdateUser)    // Editar otro usuario
				adminOnly.DELETE("/:id", users.DeleteUser) // Borrar usuario
			}

			// --- Rutas de Negocio ---
			routesGroup := protected.Group("/routes")
			{
				// Crear: SOLO ADMIN
				routesGroup.POST("", middleware.RequireRoles("admin"), routes.CreateRoute)

				// Listar: ADMIN y CONDUCTOR
				routesGroup.GET("", routes.ListRoutes)

				// Asignar: SOLO ADMIN
				routesGroup.PATCH("/:id/assign", middleware.RequireRoles("admin"), routes.AssignDriver)
			}
		}
		// 6. Iniciar Servidor
		log.Printf("🚀 Route Manager API corriendo en puerto %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("❌ Error iniciando servidor: %v", err)
		}
	}

}
