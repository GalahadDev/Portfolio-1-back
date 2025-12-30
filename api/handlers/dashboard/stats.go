package dashboard

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"gorm.io/gorm"
)

// Estructuras para la respuesta JSON bonita para el Front
type KPI struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type ChartData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type DashboardResponse struct {
	Cards     []KPI       `json:"cards"`
	ChartData []ChartData `json:"chart_data"`
}

func GetDashboardStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	// 1. Identificar quién pide los datos
	var currentUser domains.User
	if err := database.DB.First(&currentUser, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// 2. Preparar consultas base
	usersQuery := database.DB.Model(&domains.User{})

	// --- KPI 1: USUARIOS / CONDUCTORES ---
	var kpi1 int64
	if currentUser.Role == "super_admin" {
		// Super Admin: Ve el total de usuarios en la plataforma
		usersQuery.Where("status = 'active'").Count(&kpi1)
	} else {
		// Admin: Solo ve a SUS conductores activos
		usersQuery.Where("manager_id = ? AND status = 'active'", currentUser.ID).Count(&kpi1)
	}

	// --- KPI 2: RUTAS PENDIENTES ---
	var kpi2 int64
	database.DB.Model(&domains.Route{}).
		Where("status = ?", "draft").
		Scopes(filterByCreator(currentUser)).
		Count(&kpi2)

	// --- KPI 3: COMPLETADAS HOY ---
	var kpi3 int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	database.DB.Model(&domains.Route{}).
		Where("status = ? AND updated_at >= ?", "completed", startOfDay).
		Scopes(filterByCreator(currentUser)).
		Count(&kpi3)

	// --- GRÁFICO: ÚLTIMOS 7 DÍAS ---
	var chartRaw []struct {
		Date  string
		Total int
	}

	// Query agrupadora
	database.DB.Model(&domains.Route{}).
		Select("TO_CHAR(scheduled_date, 'YYYY-MM-DD') as date, count(*) as total").
		Where("status = 'completed'").
		Where("scheduled_date >= ?", now.AddDate(0, 0, -7)). // Últimos 7 días
		Scopes(filterByCreator(currentUser)).
		Group("date").
		Order("date ASC").
		Scan(&chartRaw)

	// Formatear para el frontend
	chartData := make([]ChartData, 0)
	for _, item := range chartRaw {
		chartData = append(chartData, ChartData{Date: item.Date, Count: item.Total})
	}

	// Títulos Dinámicos según rol
	labelKPI1 := "Mis Conductores"
	if currentUser.Role == "super_admin" {
		labelKPI1 = "Usuarios Totales"
	}

	// 3. Enviar Respuesta
	c.JSON(http.StatusOK, DashboardResponse{
		Cards: []KPI{
			{Label: labelKPI1, Value: kpi1, Color: "blue", Icon: "users"},
			{Label: "Rutas Pendientes", Value: kpi2, Color: "orange", Icon: "clock"},
			{Label: "Entregas Hoy", Value: kpi3, Color: "green", Icon: "check"},
		},
		ChartData: chartData,
	})
}

// --- HELPER PARA FILTRAR ---
func filterByCreator(user domains.User) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user.Role == "super_admin" {
			return db
		}
		return db.Where("creator_id = ?", user.ID)
	}
}
