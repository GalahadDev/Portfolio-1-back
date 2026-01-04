package dashboard

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// --- ESTRUCTURAS ---

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

type RouteProgress struct {
	ID                 string  `json:"id"`
	RouteName          string  `json:"route_name"`
	DriverName         string  `json:"driver_name"`
	TotalWaypoints     int     `json:"total_waypoints"`
	CompletedWaypoints int     `json:"completed_waypoints"`
	Percentage         float64 `json:"percentage"`
	Status             string  `json:"status"`
}

type DashboardResponse struct {
	Cards        []KPI           `json:"cards"`
	ChartData    []ChartData     `json:"chart_data"`
	ActiveRoutes []RouteProgress `json:"active_routes"`
}

// --- HANDLER PRINCIPAL ---

func GetDashboardStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	// 1. Identificar al usuario actual
	var currentUser domains.User
	if err := database.DB.First(&currentUser, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Inicializamos slices vacíos
	cards := []KPI{}
	activeRoutes := make([]RouteProgress, 0)
	chartData := make([]ChartData, 0)

	// 2. Lógica Diferenciada por ROL
	switch currentUser.Role {

	//  SUPER ADMIN (Gestión y Soporte)

	case "super_admin":
		// KPI 1: Usuarios Pendientes (Inactivos)
		var pendingUsers int64
		database.DB.Model(&domains.User{}).Where("status = ?", "inactive").Count(&pendingUsers)
		cards = append(cards, KPI{Label: "Usuarios Pendientes", Value: pendingUsers, Color: "red", Icon: "user-plus"})

		// KPI 2: Total de Empresas (Admins)
		var totalAdmins int64
		database.DB.Model(&domains.User{}).Where("role = ?", "admin").Count(&totalAdmins)
		cards = append(cards, KPI{Label: "Empresas Registradas", Value: totalAdmins, Color: "purple", Icon: "building"})

		// KPI 3: Volumen Global (Todas las rutas históricas)
		var totalRoutes int64
		database.DB.Model(&domains.Route{}).Count(&totalRoutes)
		cards = append(cards, KPI{Label: "Rutas Totales en Plataforma", Value: totalRoutes, Color: "blue", Icon: "globe"})

	//  DRIVER (Operativo Personal)

	case "driver":
		// KPI 1: Rutas Completadas (Histórico personal)
		var myCompletedRoutes int64
		database.DB.Model(&domains.Route{}).
			Where("driver_id = ? AND status = ?", currentUser.ID, "completed").
			Count(&myCompletedRoutes)
		cards = append(cards, KPI{Label: "Mis Rutas Completadas", Value: myCompletedRoutes, Color: "green", Icon: "flag"})

		// KPI 2: Total de Paradas/Paquetes Entregados (Histórico personal)
		var myTotalDeliveries int64
		database.DB.Model(&domains.Waypoint{}).
			Joins("JOIN routes ON waypoints.route_id = routes.id").
			Where("routes.driver_id = ? AND waypoints.is_completed = ?", currentUser.ID, true).
			Count(&myTotalDeliveries)
		cards = append(cards, KPI{Label: "Entregas Totales", Value: myTotalDeliveries, Color: "blue", Icon: "box"})

		// KPI 3: Asignaciones Pendientes (Lo que tengo que hacer hoy/mañana)
		var myPendingRoutes int64
		database.DB.Model(&domains.Route{}).
			Where("driver_id = ? AND status IN ?", currentUser.ID, []string{"pending", "in_progress"}).
			Count(&myPendingRoutes)
		cards = append(cards, KPI{Label: "Rutas Activas", Value: myPendingRoutes, Color: "orange", Icon: "map-pin"})

	//  ADMIN (Gestión de Flota - El original)

	default: // admin
		// KPI 1: Mis Conductores Activos
		var activeDrivers int64
		database.DB.Model(&domains.User{}).
			Where("manager_id = ? AND status = 'active'", currentUser.ID).
			Count(&activeDrivers)
		cards = append(cards, KPI{Label: "Mis Conductores", Value: activeDrivers, Color: "blue", Icon: "users"})

		// KPI 2: Rutas en Borrador/Planificación
		var draftRoutes int64
		database.DB.Model(&domains.Route{}).
			Where("creator_id = ? AND status = ?", currentUser.ID, "draft").
			Count(&draftRoutes)
		cards = append(cards, KPI{Label: "Planificación (Borradores)", Value: draftRoutes, Color: "orange", Icon: "clipboard"})

		// KPI 3: Productividad del Día (Entregas Hoy)
		var deliveriesToday int64
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		database.DB.Model(&domains.Waypoint{}).
			Joins("JOIN routes ON waypoints.route_id = routes.id").
			Where("routes.creator_id = ? AND waypoints.is_completed = ? AND waypoints.completed_at >= ?", currentUser.ID, true, startOfDay).
			Count(&deliveriesToday)
		cards = append(cards, KPI{Label: "Entregas Realizadas Hoy", Value: deliveriesToday, Color: "green", Icon: "check-circle"})
	}

	// 3. TABLA DE PROGRESO (Común para todos, filtrada por permisos)
	var activeRoutesDB []domains.Route
	query := database.DB.
		Preload("Waypoints").
		Preload("Driver").
		Where("status IN ?", []string{"pending", "in_progress"})

	// Filtros de la tabla según rol
	switch currentUser.Role {
	case "admin":
		query = query.Where("creator_id = ?", currentUser.ID)
	case "driver":
		query = query.Where("driver_id = ?", currentUser.ID)
	}
	// Super Admin ve todas las activas (sin where adicional)

	query.Find(&activeRoutesDB)

	for _, r := range activeRoutesDB {
		total := len(r.Waypoints)
		completed := 0
		for _, wp := range r.Waypoints {
			if wp.IsCompleted {
				completed++
			}
		}
		percentage := 0.0
		if total > 0 {
			percentage = (float64(completed) / float64(total)) * 100
			percentage = math.Round(percentage*10) / 10
		}

		driverName := "Sin Asignar"
		if r.Driver != nil {
			driverName = r.Driver.FullName
		}

		activeRoutes = append(activeRoutes, RouteProgress{
			ID:                 r.ID.String(),
			RouteName:          r.Name,
			DriverName:         driverName,
			TotalWaypoints:     total,
			CompletedWaypoints: completed,
			Percentage:         percentage,
			Status:             r.Status,
		})
	}

	// 4. GRÁFICO
	var chartRaw []struct {
		Date  string
		Total int
	}
	chartQuery := database.DB.Model(&domains.Route{}).
		Select("TO_CHAR(scheduled_date, 'YYYY-MM-DD') as date, count(*) as total").
		Where("status = 'completed'").
		Where("scheduled_date >= ?", time.Now().AddDate(0, 0, -7))

	// Filtros del gráfico
	switch currentUser.Role {
	case "admin":
		chartQuery = chartQuery.Where("creator_id = ?", currentUser.ID)
	case "driver":
		chartQuery = chartQuery.Where("driver_id = ?", currentUser.ID)
	}

	chartQuery.Group("date").Order("date ASC").Scan(&chartRaw)

	for _, item := range chartRaw {
		chartData = append(chartData, ChartData{Date: item.Date, Count: item.Total})
	}

	// 5. RESPUESTA FINAL
	c.JSON(http.StatusOK, DashboardResponse{
		Cards:        cards,
		ChartData:    chartData,
		ActiveRoutes: activeRoutes,
	})
}
