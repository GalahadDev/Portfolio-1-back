package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tu-usuario/route-manager/internal/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateOrUpdate realiza un "Upsert": Si el usuario no existe, lo crea. Si existe, actualiza datos básicos.
func (r *Repository) CreateOrUpdate(ctx context.Context, u *domain.User) error {
	// Query de Upsert en PostgreSQL
	query := `
		INSERT INTO public.users (id, email, full_name, avatar_url, role, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO UPDATE 
		SET email = EXCLUDED.email, 
		    full_name = EXCLUDED.full_name,
		    avatar_url = EXCLUDED.avatar_url,
		    updated_at = NOW();
	`

	_, err := r.db.Exec(ctx, query,
		u.ID, u.Email, u.FullName, u.AvatarURL, u.Role, u.Status,
	)

	if err != nil {
		return fmt.Errorf("error haciendo upsert de usuario: %w", err)
	}

	return nil
}

// GetByID busca un usuario por su ID
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, full_name, avatar_url, role, status, manager_id, created_at FROM public.users WHERE id = $1`

	var u domain.User
	// Scan llena la estructura automáticamente
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.Role, &u.Status, &u.ManagerID, &u.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}

	return &u, nil
}

// List devuelve usuarios filtrados por estado
func (r *Repository) List(ctx context.Context, status domain.AccountStatus) ([]domain.User, error) {
	// Query base
	query := `SELECT id, email, full_name, avatar_url, role, status, created_at FROM public.users`
	args := []interface{}{}

	// Si hay filtro, agregamos WHERE
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando usuarios: %w", err)
	}
	defer rows.Close()

	// Construir slice de usuarios
	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// UpdateStatusAndRole actualiza los permisos de un usuario
func (r *Repository) UpdateStatusAndRole(ctx context.Context, id string, status domain.AccountStatus, role domain.UserRole) error {
	query := `UPDATE public.users SET status = $1, role = $2, updated_at = NOW() WHERE id = $3`

	cmd, err := r.db.Exec(ctx, query, status, role, id)
	if err != nil {
		return fmt.Errorf("error actualizando usuario: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("usuario no encontrado")
	}

	return nil
}
