package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jrafaelca/nodara/api/heartbeat"
	"github.com/jrafaelca/nodara/internal/auth"
	"github.com/jrafaelca/nodara/internal/control"
	"github.com/jrafaelca/nodara/internal/storage"
	"github.com/jrafaelca/nodara/internal/transport/grpcjson"
	"github.com/jrafaelca/nodara/internal/transport/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type config struct {
	HTTPAddr      string
	GRPCAddr      string
	DatabaseURL   string
	MigrationsDir string
	TLSCertFile   string
	TLSKeyFile    string
	TLSCAFile     string
	StaleAfter    time.Duration
}

func main() {
	logger := newLogger()
	cfg, err := loadConfig()
	if err != nil {
		logger.Error("configuration_failed", "component", "core", "event", "configuration_failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	store, err := storage.Open(ctx, cfg.DatabaseURL, cfg.MigrationsDir)
	if err != nil {
		logger.Error("storage_start_failed", "component", "core", "event", "storage_start_failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	authService := &auth.Service{Store: store}
	if err := authService.SeedAdmin(ctx); err != nil {
		logger.Error("admin_seed_failed", "component", "core", "event", "admin_seed_failed", "error", err)
		os.Exit(1)
	}

	tlsConfig, err := serverTLSConfig(cfg)
	if err != nil {
		logger.Error("tls_configuration_failed", "component", "core", "event", "tls_configuration_failed", "error", err)
		os.Exit(1)
	}

	hub := websocket.NewHub()
	heartbeatServer := &control.HeartbeatServer{Store: store, Hub: hub, Logger: logger, StaleAfter: cfg.StaleAfter}
	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)), grpc.ForceServerCodec(grpcjson.Codec{}))
	heartbeat.RegisterAgentHeartbeatServer(grpcServer, heartbeatServer)
	grpcListener, err := listen(cfg.GRPCAddr)
	if err != nil {
		logger.Error("grpc_listen_failed", "component", "core", "event", "grpc_listen_failed", "error", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("grpc_started", "component", "core", "event", "grpc_started", "addr", cfg.GRPCAddr)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("grpc_failed", "component", "core", "event", "grpc_failed", "error", err)
			stop()
		}
	}()

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: httpHandler(store, hub, logger), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("http_started", "component", "core", "event", "http_started", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http_failed", "component", "core", "event", "http_failed", "error", err)
			stop()
		}
	}()

	go heartbeatServer.Sweep(ctx)
	<-ctx.Done()
	logger.Info("shutdown_started", "component", "core", "event", "shutdown_started")
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func httpHandler(store *storage.Store, hub *websocket.Hub, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok\n")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(r.Context()); err != nil {
			http.Error(w, "not ready\n", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ready\n")
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		agents, err := store.ListAgents(r.Context())
		if err != nil {
			logger.Error("snapshot_failed", "component", "core", "event", "snapshot_failed", "error", err)
			http.Error(w, "snapshot failed\n", http.StatusInternalServerError)
			return
		}
		hub.ServeHTTP(w, r, control.AgentEvent{Type: "agent.snapshot", OccurredAt: time.Now().UTC(), Agents: agents})
	})
	return mux
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func loadConfig() (config, error) {
	cfg := config{
		HTTPAddr:      env("NODARA_HTTP_ADDR", ":8080"),
		GRPCAddr:      env("NODARA_GRPC_ADDR", ":9090"),
		DatabaseURL:   env("NODARA_DATABASE_URL", "postgres://nodara:nodara@postgres:5432/nodara?sslmode=disable"),
		MigrationsDir: env("NODARA_MIGRATIONS_DIR", "db/migrations"),
		TLSCertFile:   env("NODARA_TLS_CERT_FILE", "/certs/server.crt"),
		TLSKeyFile:    env("NODARA_TLS_KEY_FILE", "/certs/server.key"),
		TLSCAFile:     env("NODARA_TLS_CA_FILE", "/certs/ca.crt"),
	}
	stale, err := time.ParseDuration(env("NODARA_STALE_AFTER", "15s"))
	if err != nil || stale <= 0 {
		return config{}, fmt.Errorf("invalid NODARA_STALE_AFTER")
	}
	cfg.StaleAfter = stale
	return cfg, nil
}

func serverTLSConfig(cfg config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server certificate: %w", err)
	}
	caBytes, err := os.ReadFile(cfg.TLSCAFile)
	if err != nil {
		return nil, fmt.Errorf("read client CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("parse client CA")
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool}, nil
}

func listen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
