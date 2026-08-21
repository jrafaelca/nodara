package control

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jrafaelca/nodara/api/heartbeat"
	"github.com/jrafaelca/nodara/internal/health"
	"github.com/jrafaelca/nodara/internal/storage"
	"google.golang.org/grpc/peer"
)

type HeartbeatServer struct {
	heartbeat.UnimplementedAgentHeartbeatServer
	Store      *storage.Store
	Hub        EventBroadcaster
	Logger     *slog.Logger
	StaleAfter time.Duration
}

type EventBroadcaster interface {
	Broadcast(any)
}

func (s *HeartbeatServer) Stream(stream heartbeat.AgentHeartbeat_StreamServer) error {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if message.Heartbeat == nil {
			return fmt.Errorf("heartbeat message is empty")
		}
		hb := message.Heartbeat
		if err := validateHeartbeat(hb); err != nil {
			return err
		}
		if err := verifyAgentIdentity(stream.Context(), hb.AgentID); err != nil {
			return err
		}
		now := time.Now().UTC()
		status := storage.AgentStatus{
			ID:              hb.AgentID,
			Name:            hb.AgentID,
			Hostname:        hb.Hostname,
			AgentVersion:    hb.AgentVersion,
			Status:          "healthy",
			LastHeartbeatAt: now,
			UpdatedAt:       now,
			Sequence:        hb.Sequence,
		}
		stored, err := s.Store.UpsertAgent(stream.Context(), status)
		if err != nil {
			return err
		}
		s.Logger.Info("heartbeat_received", "component", "core", "event", "heartbeat_received", "agent_id", hb.AgentID, "sequence", hb.Sequence)
		s.Hub.Broadcast(AgentEvent{Type: "agent.updated", OccurredAt: now, Agent: &stored})
		if err := stream.Send(&heartbeat.ServerMessage{HeartbeatAck: &heartbeat.HeartbeatAck{ReceivedAt: now.Format(time.RFC3339Nano), Sequence: hb.Sequence}}); err != nil {
			return err
		}
	}
}

func (s *HeartbeatServer) Sweep(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			cutoff := now.Add(-s.StaleAfter)
			agents, err := s.Store.MarkDisconnected(ctx, cutoff)
			if err != nil {
				s.Logger.Error("mark_disconnected_failed", "component", "core", "event", "mark_disconnected_failed", "error", err)
				continue
			}
			for i := range agents {
				s.Logger.Warn("agent_disconnected", "component", "core", "event", "agent_disconnected", "agent_id", agents[i].ID)
				s.Hub.Broadcast(AgentEvent{Type: "agent.disconnected", OccurredAt: now.UTC(), Agent: &agents[i]})
			}
		}
	}
}

func validateHeartbeat(hb *heartbeat.Heartbeat) error {
	if hb.AgentID == "" || hb.Hostname == "" || hb.AgentVersion == "" || hb.Sequence == 0 {
		return fmt.Errorf("heartbeat requires agent_id, hostname, agent_version and sequence")
	}
	return nil
}

func verifyAgentIdentity(ctx context.Context, agentID string) error {
	p, ok := peer.FromContext(ctx)
	if !ok || p.AuthInfo == nil {
		return fmt.Errorf("missing peer certificate")
	}
	tlsInfo, ok := p.AuthInfo.(interface{ State() tls.ConnectionState })
	if ok {
		state := tlsInfo.State()
		if len(state.PeerCertificates) > 0 && state.PeerCertificates[0].Subject.CommonName != agentID {
			return fmt.Errorf("agent certificate identity does not match agent_id")
		}
	}
	return nil
}

func IsAgentDisconnected(lastHeartbeat, now time.Time, staleAfter time.Duration) bool {
	return health.IsDisconnected(lastHeartbeat, now, staleAfter)
}
