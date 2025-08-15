// Package grpc provides the gRPC server implementation
package grpc

import (
	"context"
	"google.golang.org/grpc"
	"xoxa-message-gateway/proto"
	"xoxa-message-gateway/internal/service"
)

type MessengerServer struct {
	proto.UnimplementedMessengerServer
}

func (s *MessengerServer) Send(ctx context.Context, req *proto.SendRequest) (*proto.SendResponse, error) {
	return service.SendMessageGRPC(ctx, req)
}

func StartGRPCServer() {
	// ...existing code...
}
