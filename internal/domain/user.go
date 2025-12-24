package domain

import (
	"time"

	"github.com/google/uuid"
)

// Definimos los tipos para los Enums de la BD, así evitamos errores de tipeo
type UserRole string
type AccountStatus string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleFleetAdmin UserRole = "fleet_admin"
	RoleDriver     UserRole = "driver"

	StatusPending  AccountStatus = "pending"
	StatusActive   AccountStatus = "active"
	StatusRejected AccountStatus = "rejected"
)

// User representa a cualquier usuario del sistema (Admin o Conductor)
type User struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	Email     string        `json:"email" db:"email"`
	FullName  string        `json:"full_name" db:"full_name"`
	AvatarURL string        `json:"avatar_url" db:"avatar_url"`
	Role      UserRole      `json:"role" db:"role"`
	Status    AccountStatus `json:"status" db:"status"`

	// ManagerID es opcional, por eso usamos puntero
	ManagerID *uuid.UUID `json:"manager_id,omitempty" db:"manager_id"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser crea una instancia con valores por defecto seguros
func NewUser(id uuid.UUID, email string, name string, avatar string) *User {
	return &User{
		ID:        id,
		Email:     email,
		FullName:  name,
		AvatarURL: avatar,
		Role:      RoleDriver,    // Por defecto todos entran como conductores
		Status:    StatusPending, // Por defecto nadie entra pending
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
