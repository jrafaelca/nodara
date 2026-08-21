package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jrafaelca/nodara/api/heartbeat"
	"github.com/jrafaelca/nodara/internal/transport/grpcjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type config struct {
	ServerAddr string
	ServerName string
	AgentID    string
	Hostname   string
	Version    string
	CertFile   string
	KeyFile    string
	CAFile     string
	Interval   time.Duration
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := loadConfig()
	if err != nil {
		logger.Error("configuration_failed", "component", "agent", "event", "configuration_failed", "error", err)
		os.Exit(1)
	}
	tlsConfig, err := clientTLSConfig(cfg)
	if err != nil {
		logger.Error("tls_configuration_failed", "component", "agent", "event", "tls_configuration_failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	sequence := uint64(0)
	for {
		if err := runStream(ctx, cfg, tlsConfig, logger, &sequence); err != nil && ctx.Err() == nil {
			logger.Error("stream_failed", "component", "agent", "event", "stream_failed", "error", err)
			timer := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
			}
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func runStream(ctx context.Context, cfg config, tlsConfig *tls.Config, logger *slog.Logger, sequence *uint64) error {
	conn, err := grpc.DialContext(ctx, cfg.ServerAddr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcjson.Codec{})), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("dial core: %w", err)
	}
	defer conn.Close()
	client := heartbeat.NewAgentHeartbeatClient(conn)
	stream, err := client.Stream(ctx)
	if err != nil {
		return fmt.Errorf("open heartbeat stream: %w", err)
	}
	logger.Info("connected", "component", "agent", "event", "connected", "agent_id", cfg.AgentID, "server", cfg.ServerAddr)
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			(*sequence)++
			message := &heartbeat.AgentMessage{Heartbeat: &heartbeat.Heartbeat{AgentID: cfg.AgentID, Hostname: cfg.Hostname, AgentVersion: cfg.Version, SentAt: now.UTC().Format(time.RFC3339Nano), Sequence: *sequence, Capabilities: []string{"heartbeat"}}}
			if err := stream.Send(message); err != nil {
				return fmt.Errorf("send heartbeat: %w", err)
			}
			ack, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("receive heartbeat ack: %w", err)
			}
			logger.Info("heartbeat_sent", "component", "agent", "event", "heartbeat_sent", "agent_id", cfg.AgentID, "sequence", ack.HeartbeatAck.Sequence)
		}
	}
}

func loadConfig() (config, error) {
	cfg := config{
		ServerAddr: env("NODARA_CORE_ADDR", "nodara-core:9090"),
		ServerName: env("NODARA_TLS_SERVER_NAME", "nodara-core"),
		AgentID:    env("NODARA_AGENT_ID", "agent-local"),
		Hostname:   env("NODARA_AGENT_HOSTNAME", "demo-host"),
		Version:    env("NODARA_AGENT_VERSION", "0.1.0-dev"),
		CertFile:   env("NODARA_TLS_CERT_FILE", "/certs/agent.crt"),
		KeyFile:    env("NODARA_TLS_KEY_FILE", "/certs/agent.key"),
		CAFile:     env("NODARA_TLS_CA_FILE", "/certs/ca.crt"),
	}
	interval, err := time.ParseDuration(env("NODARA_HEARTBEAT_INTERVAL", "5s"))
	if err != nil || interval <= 0 {
		return config{}, fmt.Errorf("invalid NODARA_HEARTBEAT_INTERVAL")
	}
	cfg.Interval = interval
	return cfg, nil
}

func clientTLSConfig(cfg config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load agent certificate: %w", err)
	}
	caBytes, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("parse CA")
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: pool, ServerName: cfg.ServerName, MinVersion: tls.VersionTLS13}, nil
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
