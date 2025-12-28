package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	FullName      string    `json:"full_name"`
	AvatarURL     string    `json:"avatar_url"`
	Role          string    `gorm:"default:'driver'" json:"role"`    // driver, admin, etc.
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, active
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Hook para generar UUID antes de crear si no viene
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
