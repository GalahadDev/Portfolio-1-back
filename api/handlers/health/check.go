package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
)

// CheckResponse define la estructura de nuestra respuesta de salud
type CheckResponse struct {
	Status    string    `json:"status"`    // "ok" o "error"
	Database  string    `json:"database"`  // "connected" o "disconnected"
	Timestamp time.Time `json:"timestamp"` // Hora del servidor (útil para debug)
	Service   string    `json:"service"`   // Nombre del servicio
}

// HealthCheck verifica la conectividad del servidor y la base de datos
func HealthCheck(c *gin.Context) {
	dbStatus := "connected"
	httpStatus := http.StatusOK

	// 1. Verificar conexión a BD
	sqlDB, err := database.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "disconnected"
		httpStatus = http.StatusServiceUnavailable // 503 Service Unavailable
	}

	// 2. Construir respuesta
	response := CheckResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now(),
		Service:   "route-manager-api",
	}

	// Si la BD falla, cambiamos el estado general a error
	if dbStatus == "disconnected" {
		response.Status = "error"
	}

	c.JSON(httpStatus, response)
}
