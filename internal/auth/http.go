package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jrafaelca/nodara/internal/storage"
)

const sessionCookieName = "nodara_session"

type HTTP struct {
	Service      *Service
	Logger       *slog.Logger
	PublicURL    string
	CookieSecure bool
	Delivery     string
	SMTPURL      string
	SMTPFrom     string
	limiter      *rateLimiter
}

type contextKey string

const userContextKey contextKey = "auth.user"

func NewHTTP(service *Service, logger *slog.Logger, publicURL string, cookieSecure bool, delivery, smtpURL, smtpFrom string) *HTTP {
	return &HTTP{Service: service, Logger: logger, PublicURL: strings.TrimRight(publicURL, "/"), CookieSecure: cookieSecure, Delivery: delivery, SMTPURL: smtpURL, SMTPFrom: smtpFrom, limiter: newRateLimiter()}
}

func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/auth/login":
		h.login(w, r)
	case "/api/auth/logout":
		h.logout(w, r)
	case "/api/auth/me":
		h.me(w, r)
	case "/api/auth/change-password":
		h.changePassword(w, r)
	case "/api/auth/forgot-password":
		h.forgotPassword(w, r)
	case "/api/auth/reset-password":
		h.resetPassword(w, r)
	default:
		h.respondError(w, http.StatusNotFound, "auth route not found")
	}
}

func (h *HTTP) RequireSession(next http.Handler, allowPasswordChangeRequired bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.currentUser(r)
		if err != nil {
			h.respondError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if user.PasswordChangeRequired && !allowPasswordChangeRequired {
			h.respondError(w, http.StatusForbidden, "password change required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func (h *HTTP) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if !h.limiter.allow("login:"+clientIP(r)+":"+strings.ToLower(strings.TrimSpace(request.Identifier)), 5, time.Minute) {
		h.respondError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	user, token, expiresAt, err := h.Service.Login(r.Context(), request.Identifier, request.Password)
	if err != nil {
		h.Logger.Warn("login_failed", "component", "auth", "event", "login_failed", "identifier", request.Identifier, "ip", clientIP(r))
		h.respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.setSessionCookie(w, token, expiresAt)
	h.Logger.Info("login_succeeded", "component", "auth", "event", "login_succeeded", "user_id", user.ID)
	h.respondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *HTTP) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_ = h.Service.Logout(r.Context(), sessionToken(r))
	h.clearSessionCookie(w)
	h.respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *HTTP) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user, err := h.currentUser(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *HTTP) changePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user, err := h.currentUser(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var request struct {
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Password != request.PasswordConfirmation {
		h.respondError(w, http.StatusBadRequest, "password confirmation does not match")
		return
	}
	if err := h.Service.ChangePassword(r.Context(), user.ID, request.Password); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, token, expiresAt, err := h.Service.Login(r.Context(), user.Email, request.Password)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "could not create replacement session")
		return
	}
	h.setSessionCookie(w, token, expiresAt)
	h.respondJSON(w, http.StatusOK, map[string]any{"user": updated})
}

func (h *HTTP) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request struct {
		Identifier string `json:"identifier"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	key := "reset:" + strings.ToLower(strings.TrimSpace(request.Identifier))
	if h.limiter.allow(key, 1, time.Minute) {
		user, token, err := h.Service.CreatePasswordReset(r.Context(), request.Identifier)
		if err == nil {
			link := h.PublicURL + "/reset-password?token=" + url.QueryEscape(token)
			if err := h.deliverReset(user.Email, link); err != nil {
				h.Logger.Error("password_reset_delivery_failed", "component", "auth", "event", "password_reset_delivery_failed", "error", err)
			}
		}
	}
	h.respondJSON(w, http.StatusOK, map[string]string{"message": "If the account exists, a reset link has been sent."})
}

func (h *HTTP) resetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request struct {
		Token                string `json:"token"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Password != request.PasswordConfirmation {
		h.respondError(w, http.StatusBadRequest, "password confirmation does not match")
		return
	}
	if err := h.Service.ResetPassword(r.Context(), request.Token, request.Password); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *HTTP) currentUser(r *http.Request) (storage.User, error) {
	return h.Service.SessionUser(r.Context(), sessionToken(r))
}

func (h *HTTP) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", Expires: expiresAt, HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode})
}

func (h *HTTP) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode})
}

func (h *HTTP) deliverReset(email, link string) error {
	if h.Delivery == "smtp" {
		return sendSMTP(h.SMTPURL, h.SMTPFrom, email, link)
	}
	h.Logger.Info("password_reset_link", "component", "auth", "event", "password_reset_link", "email", email, "url", link)
	return nil
}

func sendSMTP(rawURL, from, to, link string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || parsed.User == nil || from == "" {
		return fmt.Errorf("invalid SMTP configuration")
	}
	host, _, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return err
	}
	password, _ := parsed.User.Password()
	message := []byte("To: " + to + "\r\nSubject: Nodara password reset\r\n\r\nReset your password: " + link + "\r\n")
	return smtp.SendMail(parsed.Host, smtp.PlainAuth("", parsed.User.Username(), password, host), from, []string{to}, message)
}

func (h *HTTP) respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *HTTP) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(destination); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return false
	}
	return true
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{entries: make(map[string][]time.Time)}
}

func (r *rateLimiter) allow(key string, limit int, window time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-window)
	entries := r.entries[key][:0]
	for _, entry := range r.entries[key] {
		if entry.After(cutoff) {
			entries = append(entries, entry)
		}
	}
	if len(entries) >= limit {
		r.entries[key] = entries
		return false
	}
	r.entries[key] = append(entries, now)
	return true
}
