package domains

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	FullName      string    `json:"full_name"`
	AvatarURL     string    `json:"avatar_url"`
	Role          string    `gorm:"default:'driver'" json:"role"`     // super_admin, admin, driver
	Status        string    `gorm:"default:'inactive'" json:"status"` // active, inactive
	EmailVerified bool      `json:"email_verified"`

	// FleetCode: Código único que el Admin comparte
	FleetCode string `gorm:"uniqueIndex;default:null" json:"fleet_code,omitempty"`

	// ManagerID: Quién es mi jefe (Self-Referential Foreign Key)
	ManagerID *uuid.UUID `gorm:"type:uuid;default:null" json:"manager_id,omitempty"`

	// Relaciones de GORM
	Manager *User  `gorm:"foreignKey:ManagerID" json:"-"`       // Mi Jefe
	Drivers []User `gorm:"foreignKey:ManagerID" json:"drivers"` // Mis Conductores

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}
