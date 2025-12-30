package optimization

import (
	"math"
	"math/rand"
	"time"

	"github.com/tu-usuario/route-manager/api/domains"
)

// OptimizeRoute aplica la estrategia híbrida: Nearest Neighbor + Simulated Annealing
func OptimizeRoute(waypoints []domains.Waypoint) []domains.Waypoint {
	if len(waypoints) <= 2 {
		return waypoints // No hay nada que optimizar
	}

	// Paso 1: Solución Inicial Rápida (Greedy / Nearest Neighbor)
	initialSolution := nearestNeighbor(waypoints)

	// Paso 2: Refinamiento (Simulated Annealing)
	finalSolution := simulatedAnnealing(initialSolution)

	return finalSolution
}

// nearestNeighbor: Algoritmo voraz. Desde el punto actual, va al más cercano disponible.
func nearestNeighbor(points []domains.Waypoint) []domains.Waypoint {
	if len(points) == 0 {
		return points
	}

	// Copiamos para no mutar el original
	pending := make([]domains.Waypoint, len(points))
	copy(pending, points)

	// El primer punto (depósito/inicio) se queda fijo
	solution := []domains.Waypoint{pending[0]}
	pending = pending[1:] // Quitamos el primero de la lista de pendientes

	current := solution[0]

	for len(pending) > 0 {
		closestIndex := -1
		minDist := math.MaxFloat64

		// Buscar el más cercano de los pendientes
		for i, p := range pending {
			dist := HaversineDistance(current.Latitude, current.Longitude, p.Latitude, p.Longitude)
			if dist < minDist {
				minDist = dist
				closestIndex = i
			}
		}

		// Añadir a la solución y actualizar el actual
		current = pending[closestIndex]
		solution = append(solution, current)

		// Eliminar de pendientes (Truco rápido de slice)
		pending = append(pending[:closestIndex], pending[closestIndex+1:]...)
	}

	return solution
}

// simulatedAnnealing: Intenta mejorar la ruta intercambiando pares aleatorios
func simulatedAnnealing(route []domains.Waypoint) []domains.Waypoint {
	rand.Seed(time.Now().UnixNano())

	// Configuración del "Horno"
	currentSolution := make([]domains.Waypoint, len(route))
	copy(currentSolution, route)

	currentDist := CalculateRouteDistance(currentSolution)
	bestSolution := currentSolution
	bestDist := currentDist

	temp := 10000.0      // Temperatura inicial (alta probabilidad de aceptar cambios malos)
	coolingRate := 0.995 // Velocidad de enfriamiento (cuanto más cerca de 1, más lento y preciso)

	// Iteramos hasta que se "enfríe" el sistema
	for temp > 1 {
		// 1. Crear una solución vecina (intercambiar 2 ciudades al azar)
		// OJO: Nunca tocamos el índice 0 (Punto de partida)
		newSolution := make([]domains.Waypoint, len(currentSolution))
		copy(newSolution, currentSolution)

		// Elegir dos índices aleatorios (entre 1 y len-1)
		pos1 := rand.Intn(len(route)-1) + 1
		pos2 := rand.Intn(len(route)-1) + 1

		// Swap
		newSolution[pos1], newSolution[pos2] = newSolution[pos2], newSolution[pos1]

		// 2. Calcular energía (Distancia)
		newDist := CalculateRouteDistance(newSolution)

		// 3. Decidir si aceptamos la nueva solución
		// Si es mejor, la aceptamos siempre.
		// Si es peor, la aceptamos con una probabilidad basada en la temperatura actual.
		if newDist < currentDist || math.Exp((currentDist-newDist)/temp) > rand.Float64() {
			currentSolution = newSolution
			currentDist = newDist

			// ¿Es la mejor histórica?
			if currentDist < bestDist {
				bestSolution = currentSolution
				bestDist = currentDist
			}
		}

		// 4. Enfriar
		temp *= coolingRate
	}

	return bestSolution
}
