package optimization

import (
	"math"

	"github.com/tu-usuario/route-manager/api/domains"
)

const EarthRadiusKm = 6371.0

// HaversineDistance calcula la distancia en Km entre dos puntos (lat/lng)
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// CalculateRouteDistance suma la distancia total de una secuencia de waypoints
// Asumimos que la ruta empieza en el primer punto de la lista
func CalculateRouteDistance(waypoints []domains.Waypoint) float64 {
	totalDist := 0.0
	for i := 0; i < len(waypoints)-1; i++ {
		totalDist += HaversineDistance(
			waypoints[i].Latitude, waypoints[i].Longitude,
			waypoints[i+1].Latitude, waypoints[i+1].Longitude,
		)
	}
	return totalDist
}
