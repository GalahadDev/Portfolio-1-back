package waypoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"github.com/tu-usuario/route-manager/api/services/storage"
)

func MarkWaypointComplete(c *gin.Context) {
	waypointID := c.Param("id")
	userID, _ := c.Get("userID")

	// 1. Buscar Waypoint
	var wp domains.Waypoint
	if err := database.DB.Preload("Route").First(&wp, "id = ?", waypointID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Punto no encontrado"})
		return
	}

	// 2. Seguridad
	if wp.Route.DriverID == nil || wp.Route.DriverID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Sin permiso"})
		return
	}

	// 3. Procesar Archivo

	// Form-data key: "proof_file"
	fileHeader, err := c.FormFile("proof_file")
	var storagePath string

	if err == nil {
		// A. Abrir archivo
		file, openErr := fileHeader.Open()
		if openErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error abriendo archivo"})
			return
		}
		defer file.Close()

		// B. Inicializar Servicio Storage
		storageSvc := storage.NewService()

		// C. Subir a Supabase
		contentType := fileHeader.Header.Get("Content-Type")
		path, uploadErr := storageSvc.UploadFile(file, fileHeader.Filename, contentType)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo foto: " + uploadErr.Error()})
			return
		}
		storagePath = path
	}

	// 4. Actualizar BD
	now := time.Now()
	wp.IsCompleted = true
	wp.CompletedAt = &now

	if storagePath != "" {
		wp.ProofPhotoURL = &storagePath
	}

	if err := database.DB.Save(&wp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando cambios"})
		return
	}

	// Para responder al front, firmamos la URL recién creada
	var signedURL string
	if storagePath != "" {
		svc := storage.NewService()
		signedURL, _ = svc.GetSignedURL(storagePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Entrega completada",
		"completed_at":     wp.CompletedAt,
		"proof_signed_url": signedURL,
	})
}
