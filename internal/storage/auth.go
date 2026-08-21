package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID                     string     `json:"id"`
	Username               string     `json:"username"`
	Email                  string     `json:"email"`
	PasswordHash           string     `json:"-"`
	Active                 bool       `json:"active"`
	PasswordChangeRequired bool       `json:"password_change_required"`
	PasswordChangedAt      *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (s *Store) SeedAdmin(ctx context.Context, passwordHash string) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&exists); err != nil {
		return fmt.Errorf("check users: %w", err)
	}
	if exists {
		return nil
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO users (id, username, email, password_hash, password_change_required) VALUES ('user_admin', 'admin', 'admin@nodara.dev', $1, true)`, passwordHash)
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

func (s *Store) FindUserByIdentifier(ctx context.Context, identifier string) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT id, username, email, password_hash, active, password_change_required, password_changed_at, created_at, updated_at FROM users WHERE lower(email)=lower($1) OR lower(username)=lower($1)`, identifier))
}

func (s *Store) GetUser(ctx context.Context, id string) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT id, username, email, password_hash, active, password_change_required, password_changed_at, created_at, updated_at FROM users WHERE id=$1`, id))
}

func (s *Store) CreateSession(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`, id, userID, tokenHash, expiresAt)
	return err
}

func (s *Store) UserForSession(ctx context.Context, tokenHash string, now time.Time) (User, error) {
	user, err := scanUser(s.pool.QueryRow(ctx, `SELECT u.id, u.username, u.email, u.password_hash, u.active, u.password_change_required, u.password_changed_at, u.created_at, u.updated_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at > $2`, tokenHash, now))
	if err != nil {
		return User{}, err
	}
	_, _ = s.pool.Exec(ctx, `UPDATE sessions SET last_seen_at=$1 WHERE token_hash=$2`, now, tokenHash)
	return user, nil
}

func (s *Store) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at=$1 WHERE token_hash=$2 AND revoked_at IS NULL`, now, tokenHash)
	return err
}

func (s *Store) RevokeUserSessions(ctx context.Context, userID string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at=$1 WHERE user_id=$2 AND revoked_at IS NULL`, now, userID)
	return err
}

func (s *Store) UpdatePassword(ctx context.Context, userID, passwordHash string, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash=$1, password_change_required=false, password_changed_at=$2, updated_at=$2 WHERE id=$3`, passwordHash, now, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE sessions SET revoked_at=$1 WHERE user_id=$2 AND revoked_at IS NULL`, now, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateResetToken(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id=$1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`, id, userID, tokenHash, expiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ResetPassword(ctx context.Context, tokenHash, passwordHash string, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID string
	if err := tx.QueryRow(ctx, `SELECT user_id FROM password_reset_tokens WHERE token_hash=$1 AND used_at IS NULL AND expires_at>$2 FOR UPDATE`, tokenHash, now).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("invalid or expired reset token")
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash=$1, password_change_required=false, password_changed_at=$2, updated_at=$2 WHERE id=$3`, passwordHash, now, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE sessions SET revoked_at=$1 WHERE user_id=$2 AND revoked_at IS NULL`, now, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at=$1 WHERE token_hash=$2`, now, tokenHash); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func scanUser(row rowScanner) (User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Active, &user.PasswordChangeRequired, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, err
	}
	return user, nil
}
