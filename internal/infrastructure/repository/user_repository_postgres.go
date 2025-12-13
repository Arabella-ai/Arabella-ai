package repository

import (
	"context"
	"errors"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepositoryPostgres implements UserRepository for PostgreSQL
type UserRepositoryPostgres struct {
	pool *pgxpool.Pool
}

// NewUserRepositoryPostgres creates a new UserRepositoryPostgres
func NewUserRepositoryPostgres(pool *pgxpool.Pool) repository.UserRepository {
	return &UserRepositoryPostgres{pool: pool}
}

// Create creates a new user
func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.GoogleID,
		user.Name,
		user.AvatarURL,
		user.Credits,
		user.Tier,
		user.SubscriptionExpiresAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleID,
		&user.Name,
		&user.AvatarURL,
		&user.Credits,
		&user.Tier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryPostgres) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleID,
		&user.Name,
		&user.AvatarURL,
		&user.Credits,
		&user.Tier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByGoogleID retrieves a user by Google ID
func (r *UserRepositoryPostgres) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	query := `
		SELECT id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE google_id = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleID,
		&user.Name,
		&user.AvatarURL,
		&user.Credits,
		&user.Tier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates an existing user
func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $2, google_id = $3, name = $4, avatar_url = $5, credits = $6, 
		    tier = $7, subscription_expires_at = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.GoogleID,
		user.Name,
		user.AvatarURL,
		user.Credits,
		user.Tier,
		user.SubscriptionExpiresAt,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// UpdateCredits updates user credits atomically
func (r *UserRepositoryPostgres) UpdateCredits(ctx context.Context, id uuid.UUID, delta int) error {
	query := `
		UPDATE users
		SET credits = credits + $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL AND credits + $2 >= 0
	`

	result, err := r.pool.Exec(ctx, query, id, delta, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		// Either user not found or insufficient credits
		user, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if user.Credits+delta < 0 {
			return entity.ErrInsufficientCredits
		}
		return entity.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// List lists users with pagination
func (r *UserRepositoryPostgres) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	query := `
		SELECT id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.GoogleID,
			&user.Name,
			&user.AvatarURL,
			&user.Credits,
			&user.Tier,
			&user.SubscriptionExpiresAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetByTier retrieves users by tier
func (r *UserRepositoryPostgres) GetByTier(ctx context.Context, tier entity.UserTier, offset, limit int) ([]*entity.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users WHERE tier = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, tier).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	query := `
		SELECT id, email, google_id, name, avatar_url, credits, tier, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE tier = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tier, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.GoogleID,
			&user.Name,
			&user.AvatarURL,
			&user.Credits,
			&user.Tier,
			&user.SubscriptionExpiresAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

