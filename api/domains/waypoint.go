package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Waypoint struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RouteID uuid.UUID `gorm:"type:uuid;column:route_id" json:"route_id"`

	SequenceOrder int     `gorm:"not null" json:"sequence_order"`
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	CustomerName  string  `json:"customer_name"`
	Notes         string  `json:"notes"`

	IsCompleted bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (w *Waypoint) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return
}
