CREATE TABLE agents (
    id text PRIMARY KEY,
    name text NOT NULL,
    hostname text NOT NULL,
    agent_version text NOT NULL,
    status text NOT NULL CHECK (status IN ('healthy', 'disconnected')),
    last_heartbeat_at timestamptz NOT NULL,
    first_seen_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    sequence bigint NOT NULL DEFAULT 0
);

CREATE INDEX agents_status_idx ON agents (status);
CREATE INDEX agents_last_heartbeat_idx ON agents (last_heartbeat_at);
