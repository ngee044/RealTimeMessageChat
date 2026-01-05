package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")

	t.Cleanup(func() {
		sqlxDB.Close()
	})

	return sqlxDB, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		user := &User{
			UserID:   "test_user_123",
			Username: sql.NullString{String: "testuser", Valid: true},
			Email:    sql.NullString{String: "test@example.com", Valid: true},
			Status:   "offline",
		}

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(1), now, now)

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.UserID, user.Username, user.Email, user.Status).
			WillReturnRows(rows)

		err := repo.Create(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		user := &User{
			UserID:   "test_user_456",
			Username: sql.NullString{String: "testuser2", Valid: true},
			Email:    sql.NullString{String: "test2@example.com", Valid: true},
			Status:   "offline",
		}

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.UserID, user.Username, user.Email, user.Status).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("user found", func(t *testing.T) {
		userID := "test_user_123"
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at"}).
			AddRow(int64(1), userID, "testuser", "test@example.com", "online", now, now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE user_id`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.UserID)
		assert.Equal(t, "testuser", user.Username.String)
		assert.Equal(t, "online", user.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent_user"

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE user_id`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByUserID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful status update", func(t *testing.T) {
		userID := "test_user_123"
		status := "online"

		mock.ExpectExec(`UPDATE users SET status`).
			WithArgs(status, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateStatus(ctx, userID, status)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent_user"
		status := "online"

		mock.ExpectExec(`UPDATE users SET status`).
			WithArgs(status, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateStatus(ctx, userID, status)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		userID := "test_user_123"

		mock.ExpectExec(`DELETE FROM users WHERE user_id`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, userID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent_user"

		mock.ExpectExec(`DELETE FROM users WHERE user_id`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_List(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		limit, offset := 10, 0
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at"}).
			AddRow(int64(1), "user1", "name1", "email1@test.com", "online", now, now, now).
			AddRow(int64(2), "user2", "name2", "email2@test.com", "offline", now, now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users ORDER BY created_at DESC`).
			WithArgs(limit, offset).
			WillReturnRows(rows)

		users, err := repo.List(ctx, limit, offset)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1", users[0].UserID)
		assert.Equal(t, "user2", users[1].UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		limit, offset := 10, 0

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT (.+) FROM users ORDER BY created_at DESC`).
			WithArgs(limit, offset).
			WillReturnRows(rows)

		users, err := repo.List(ctx, limit, offset)

		assert.NoError(t, err)
		assert.Len(t, users, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Exists(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("user exists", func(t *testing.T) {
		userID := "test_user_123"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID).
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, userID)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user does not exist", func(t *testing.T) {
		userID := "nonexistent_user"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID).
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, userID)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Count(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("count users", func(t *testing.T) {
		expectedCount := int64(42)

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

		mock.ExpectQuery(`SELECT COUNT`).
			WillReturnRows(rows)

		count, err := repo.Count(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_ListByStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("list online users", func(t *testing.T) {
		status := "online"
		limit, offset := 10, 0
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at"}).
			AddRow(int64(1), "user1", "name1", "email1@test.com", "online", now, now, now).
			AddRow(int64(2), "user2", "name2", "email2@test.com", "online", now, now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE status`).
			WithArgs(status, limit, offset).
			WillReturnRows(rows)

		users, err := repo.ListByStatus(ctx, status, limit, offset)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "online", users[0].Status)
		assert.Equal(t, "online", users[1].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("user found by ID", func(t *testing.T) {
		userID := int64(1)
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at"}).
			AddRow(userID, "test_user", "testuser", "test@example.com", "online", now, now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test_user", user.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found by ID", func(t *testing.T) {
		userID := int64(9999)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		now := time.Now()
		user := &User{
			UserID:   "test_user_123",
			Username: sql.NullString{String: "updated_name", Valid: true},
			Email:    sql.NullString{String: "updated@example.com", Valid: true},
			Status:   "online",
			LastSeen: sql.NullTime{Time: now, Valid: true},
		}

		rows := sqlmock.NewRows([]string{"updated_at"}).AddRow(now)

		mock.ExpectQuery(`UPDATE users SET`).
			WithArgs(user.Username, user.Email, user.Status, user.LastSeen, user.UserID).
			WillReturnRows(rows)

		err := repo.Update(ctx, user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on update", func(t *testing.T) {
		now := time.Now()
		user := &User{
			UserID:   "test_user_456",
			Username: sql.NullString{String: "name", Valid: true},
			Email:    sql.NullString{String: "email@example.com", Valid: true},
			Status:   "offline",
			LastSeen: sql.NullTime{Time: now, Valid: true},
		}

		mock.ExpectQuery(`UPDATE users SET`).
			WithArgs(user.Username, user.Email, user.Status, user.LastSeen, user.UserID).
			WillReturnError(sql.ErrConnDone)

		err := repo.Update(ctx, user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_UpdateLastSeen(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successful last seen update", func(t *testing.T) {
		userID := "test_user_123"

		mock.ExpectExec(`UPDATE users SET last_seen`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateLastSeen(ctx, userID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		userID := "test_user_456"

		mock.ExpectExec(`UPDATE users SET last_seen`).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateLastSeen(ctx, userID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
