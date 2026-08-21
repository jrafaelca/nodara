package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jrafaelca/nodara/internal/storage"
	"golang.org/x/crypto/argon2"
)

const (
	SessionTTL    = 8 * time.Hour
	ResetTokenTTL = time.Hour
)

type Service struct {
	Store *storage.Store
}

func (s *Service) SeedAdmin(ctx context.Context) error {
	hash, err := HashPassword("password")
	if err != nil {
		return err
	}
	return s.Store.SeedAdmin(ctx, hash)
}

func (s *Service) Login(ctx context.Context, identifier, password string) (storage.User, string, time.Time, error) {
	user, err := s.Store.FindUserByIdentifier(ctx, strings.TrimSpace(identifier))
	if err != nil || !user.Active || !VerifyPassword(user.PasswordHash, password) {
		return storage.User{}, "", time.Time{}, ErrInvalidCredentials
	}
	token, err := randomToken()
	if err != nil {
		return storage.User{}, "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(SessionTTL)
	if err := s.Store.CreateSession(ctx, randomID("session_"), user.ID, hashToken(token), expiresAt); err != nil {
		return storage.User{}, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

func (s *Service) SessionUser(ctx context.Context, token string) (storage.User, error) {
	if token == "" {
		return storage.User{}, ErrUnauthenticated
	}
	user, err := s.Store.UserForSession(ctx, hashToken(token), time.Now().UTC())
	if err != nil || !user.Active {
		return storage.User{}, ErrUnauthenticated
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.Store.RevokeSession(ctx, hashToken(token), time.Now().UTC())
}

func (s *Service) ChangePassword(ctx context.Context, userID, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.Store.UpdatePassword(ctx, userID, hash, time.Now().UTC())
}

func (s *Service) CreatePasswordReset(ctx context.Context, identifier string) (storage.User, string, error) {
	user, err := s.Store.FindUserByIdentifier(ctx, strings.TrimSpace(identifier))
	if err != nil || !user.Active {
		return storage.User{}, "", ErrUnknownUser
	}
	token, err := randomToken()
	if err != nil {
		return storage.User{}, "", err
	}
	if err := s.Store.CreateResetToken(ctx, randomID("reset_"), user.ID, hashToken(token), time.Now().UTC().Add(ResetTokenTTL)); err != nil {
		return storage.User{}, "", err
	}
	return user, token, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.Store.ResetPassword(ctx, hashToken(token), hash, time.Now().UTC())
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUnauthenticated    = fmt.Errorf("unauthenticated")
	ErrUnknownUser        = fmt.Errorf("user not found")
)

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must contain at least 12 characters")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("argon2id$v=19$m=65536,t=3,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" || parts[2] != "m=65536,t=3,p=4" {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, uint32(len(expected)))
	return subtleCompare(actual, expected)
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func randomID(prefix string) string {
	token, _ := randomToken()
	return prefix + token
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func subtleCompare(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var result byte
	for i := range left {
		result |= left[i] ^ right[i]
	}
	return result == 0
}
