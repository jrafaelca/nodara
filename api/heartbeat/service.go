package heartbeat

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentMessage struct {
	Heartbeat *Heartbeat `json:"heartbeat,omitempty"`
}

type Heartbeat struct {
	AgentID      string   `json:"agent_id"`
	Hostname     string   `json:"hostname"`
	AgentVersion string   `json:"agent_version"`
	SentAt       string   `json:"sent_at"`
	Sequence     uint64   `json:"sequence"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type ServerMessage struct {
	HeartbeatAck *HeartbeatAck `json:"heartbeat_ack,omitempty"`
}

type HeartbeatAck struct {
	ReceivedAt string `json:"received_at"`
	Sequence   uint64 `json:"sequence"`
}

type AgentHeartbeatClient interface {
	Stream(ctx context.Context, opts ...grpc.CallOption) (AgentHeartbeat_StreamClient, error)
}

type agentHeartbeatClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentHeartbeatClient(cc grpc.ClientConnInterface) AgentHeartbeatClient {
	return &agentHeartbeatClient{cc: cc}
}

func (c *agentHeartbeatClient) Stream(ctx context.Context, opts ...grpc.CallOption) (AgentHeartbeat_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &AgentHeartbeat_ServiceDesc.Streams[0], "/nodara.agent.v1.AgentHeartbeat/Stream", opts...)
	if err != nil {
		return nil, err
	}
	return &agentHeartbeatStreamClient{ClientStream: stream}, nil
}

type AgentHeartbeat_StreamClient interface {
	Send(*AgentMessage) error
	Recv() (*ServerMessage, error)
	grpc.ClientStream
}

type agentHeartbeatStreamClient struct {
	grpc.ClientStream
}

func (x *agentHeartbeatStreamClient) Send(message *AgentMessage) error {
	return x.ClientStream.SendMsg(message)
}

func (x *agentHeartbeatStreamClient) Recv() (*ServerMessage, error) {
	message := new(ServerMessage)
	if err := x.ClientStream.RecvMsg(message); err != nil {
		return nil, err
	}
	return message, nil
}

type AgentHeartbeatServer interface {
	Stream(AgentHeartbeat_StreamServer) error
	mustEmbedUnimplementedAgentHeartbeatServer()
}

type UnimplementedAgentHeartbeatServer struct{}

func (UnimplementedAgentHeartbeatServer) Stream(AgentHeartbeat_StreamServer) error {
	return status.Error(codes.Unimplemented, "method Stream not implemented")
}

func (UnimplementedAgentHeartbeatServer) mustEmbedUnimplementedAgentHeartbeatServer() {}

type UnsafeAgentHeartbeatServer interface {
	mustEmbedUnimplementedAgentHeartbeatServer()
}

func RegisterAgentHeartbeatServer(registrar grpc.ServiceRegistrar, server AgentHeartbeatServer) {
	registrar.RegisterService(&AgentHeartbeat_ServiceDesc, server)
}

func _AgentHeartbeat_Stream_Handler(server interface{}, stream grpc.ServerStream) error {
	return server.(AgentHeartbeatServer).Stream(&agentHeartbeatStreamServer{ServerStream: stream})
}

type AgentHeartbeat_StreamServer interface {
	Send(*ServerMessage) error
	Recv() (*AgentMessage, error)
	grpc.ServerStream
}

type agentHeartbeatStreamServer struct {
	grpc.ServerStream
}

func (x *agentHeartbeatStreamServer) Send(message *ServerMessage) error {
	return x.ServerStream.SendMsg(message)
}

func (x *agentHeartbeatStreamServer) Recv() (*AgentMessage, error) {
	message := new(AgentMessage)
	if err := x.ServerStream.RecvMsg(message); err != nil {
		return nil, err
	}
	return message, nil
}

var AgentHeartbeat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "nodara.agent.v1.AgentHeartbeat",
	HandlerType: (*AgentHeartbeatServer)(nil),
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _AgentHeartbeat_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}
