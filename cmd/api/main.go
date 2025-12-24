package main

import (
	"log"
	"net/http"

	"github.com/tu-usuario/route-manager/internal/config"
	"github.com/tu-usuario/route-manager/internal/platform/database"
	"github.com/tu-usuario/route-manager/internal/route"
	"github.com/tu-usuario/route-manager/internal/user"
)

func main() {
	// 1. Configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando config: %v", err)
	}

	// 2. Base de Datos
	dbPool, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a BD: %v", err)
	}
	defer dbPool.Close()

	// 3. Inyección de Dependencias (Wiring)

	// -- Feature: User --
	userRepo := user.NewRepository(dbPool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// -- Feature: Route  --
	routeRepo := route.NewRepository(dbPool)
	routeService := route.NewService(routeRepo)
	routeHandler := route.NewHandler(routeService)

	// 4. Inicializar Router
	r := NewRouter(cfg, userHandler, routeHandler)

	// 5. Arrancar Servidor
	log.Printf("🚀 Route Manager corriendo en puerto %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("❌ Error fatal en servidor: %v", err)
	}
}
