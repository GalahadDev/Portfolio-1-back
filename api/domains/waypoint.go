package domains

import (
	"time"

	"github.com/google/uuid"
)

type Waypoint struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RouteID uuid.UUID `gorm:"type:uuid;column:route_id" json:"route_id"`

	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	SequenceOrder int     `json:"sequence_order"`
	CustomerName  string  `json:"customer_name"`
	Notes         string  `json:"notes"`

	IsCompleted   bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at"`
	ProofPhotoURL *string    `json:"proof_photo_url"`

	// Relaciones
	Route Route `gorm:"foreignKey:RouteID" json:"-"`
}
