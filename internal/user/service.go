package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/internal/domain"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SyncUser se llama cada vez que alguien se loguea en el frontend.
func (s *Service) SyncUser(ctx context.Context, id uuid.UUID, email, name, avatar string) (*domain.User, error) {
	// 1. Instanciamos un usuario nuevo (con defaults seguros)
	user := domain.NewUser(id, email, name, avatar)

	// Si ya existe en la BD, queremos preservar su Rol y Status actual,
	existing, err := s.repo.GetByID(ctx, id.String())
	if err == nil && existing != nil {
		// Si ya existe, mantenemos su rol y status originales
		user.Role = existing.Role
		user.Status = existing.Status
	}

	// 3. Guardamos/Actualizamos
	if err := s.repo.CreateOrUpdate(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// internal/user/service.go

func (s *Service) ListUsers(ctx context.Context, status string) ([]domain.User, error) {
	return s.repo.List(ctx, domain.AccountStatus(status))
}

func (s *Service) ApproveUser(ctx context.Context, id string, role string) error {
	return s.repo.UpdateStatusAndRole(ctx, id, domain.StatusActive, domain.UserRole(role))
}
