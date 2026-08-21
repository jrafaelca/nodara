package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentStatus struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	AgentVersion    string    `json:"agent_version"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
	FirstSeenAt     time.Time `json:"first_seen_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Sequence        uint64    `json:"sequence"`
}

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL, migrationsDir string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	store := &Store{pool: pool}
	if err := store.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := store.ApplyMigrations(ctx, migrationsDir); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) ApplyMigrations(ctx context.Context, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, string(contents)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

func (s *Store) UpsertAgent(ctx context.Context, agent AgentStatus) (AgentStatus, error) {
	const query = `
		INSERT INTO agents (id, name, hostname, agent_version, status, last_heartbeat_at, first_seen_at, updated_at, sequence)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now(), $7)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name,
			hostname=EXCLUDED.hostname,
			agent_version=EXCLUDED.agent_version,
			status=EXCLUDED.status,
			last_heartbeat_at=EXCLUDED.last_heartbeat_at,
			updated_at=now(),
			sequence=EXCLUDED.sequence
		RETURNING id, name, hostname, agent_version, status, last_heartbeat_at, first_seen_at, updated_at, sequence`
	return s.scanAgent(s.pool.QueryRow(ctx, query, agent.ID, agent.Name, agent.Hostname, agent.AgentVersion, agent.Status, agent.LastHeartbeatAt, agent.Sequence))
}

func (s *Store) ListAgents(ctx context.Context) ([]AgentStatus, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, hostname, agent_version, status, last_heartbeat_at, first_seen_at, updated_at, sequence FROM agents ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()
	result := make([]AgentStatus, 0)
	for rows.Next() {
		agent, err := s.scanAgent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return result, nil
}

func (s *Store) MarkDisconnected(ctx context.Context, cutoff time.Time) ([]AgentStatus, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE agents SET status='disconnected', updated_at=now()
		WHERE status <> 'disconnected' AND last_heartbeat_at < $1
		RETURNING id, name, hostname, agent_version, status, last_heartbeat_at, first_seen_at, updated_at, sequence`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("mark disconnected: %w", err)
	}
	defer rows.Close()
	result := make([]AgentStatus, 0)
	for rows.Next() {
		agent, err := s.scanAgent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, agent)
	}
	return result, rows.Err()
}

type rowScanner interface {
	Scan(...any) error
}

func (s *Store) scanAgent(row rowScanner) (AgentStatus, error) {
	var agent AgentStatus
	err := row.Scan(&agent.ID, &agent.Name, &agent.Hostname, &agent.AgentVersion, &agent.Status, &agent.LastHeartbeatAt, &agent.FirstSeenAt, &agent.UpdatedAt, &agent.Sequence)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AgentStatus{}, fmt.Errorf("agent not found")
		}
		return AgentStatus{}, fmt.Errorf("scan agent: %w", err)
	}
	return agent, nil
}
