package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// User represents a user in the database
type User struct {
	ID        int64          `db:"id" json:"id"`
	UserID    string         `db:"user_id" json:"user_id"`
	Username  sql.NullString `db:"username" json:"username,omitempty"`
	Email     sql.NullString `db:"email" json:"email,omitempty"`
	Status    string         `db:"status" json:"status"`
	LastSeen  sql.NullTime   `db:"last_seen" json:"last_seen,omitempty"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

// UserRepository defines user data access methods
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUserID(ctx context.Context, userID string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateStatus(ctx context.Context, userID string, status string) error
	UpdateLastSeen(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, userID string) (bool, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (user_id, username, email, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowxContext(
		ctx, query,
		user.UserID, user.Username, user.Email, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByUserID retrieves a user by user_id
func (r *userRepository) GetByUserID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, user_id, username, email, status, last_seen, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	return &user, err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, user_id, username, email, status, last_seen, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found with ID: %d", id)
	}
	return &user, err
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, status = $3, last_seen = $4
		WHERE user_id = $5
		RETURNING updated_at
	`

	return r.db.QueryRowxContext(
		ctx, query,
		user.Username, user.Email, user.Status, user.LastSeen, user.UserID,
	).Scan(&user.UpdatedAt)
}

// UpdateStatus updates user status
func (r *userRepository) UpdateStatus(ctx context.Context, userID string, status string) error {
	query := `
		UPDATE users
		SET status = $1, last_seen = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// UpdateLastSeen updates user's last seen timestamp
func (r *userRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_seen = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// List retrieves a paginated list of users
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, user_id, username, email, status, last_seen, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}

// ListByStatus retrieves users by status
func (r *userRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, user_id, username, email, status, last_seen, created_at, updated_at
		FROM users
		WHERE status = $1
		ORDER BY last_seen DESC
		LIMIT $2 OFFSET $3
	`

	var users []*User
	err := r.db.SelectContext(ctx, &users, query, status, limit, offset)
	return users, err
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

// Exists checks if a user exists
func (r *userRepository) Exists(ctx context.Context, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID)
	return exists, err
}
