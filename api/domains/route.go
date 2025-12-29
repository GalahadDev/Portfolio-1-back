package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Route struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CreatorID uuid.UUID  `gorm:"type:uuid;column:creator_id" json:"creator_id"`
	DriverID  *uuid.UUID `gorm:"type:uuid;column:driver_id" json:"driver_id"`

	Name                 string     `gorm:"not null" json:"name"`
	Status               string     `gorm:"default:'draft'" json:"status"`
	ScheduledDate        *time.Time `json:"scheduled_date"`
	TotalDistanceKm      int        `json:"total_distance_km"`
	EstimatedDurationMin int        `json:"estimated_duration_min"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relaciones
	Creator   User       `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Driver    *User      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Waypoints []Waypoint `gorm:"foreignKey:RouteID" json:"waypoints,omitempty"`
}

func (r *Route) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
