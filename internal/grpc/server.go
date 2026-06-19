// Package grpc provides the gRPC server implementation for the Messenger service.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xoxa-message-gateway/internal/service"
	"xoxa-message-gateway/pkg/dto"
	pb "xoxa-message-gateway/proto"
)

// MessengerServer adapts the application Service to the generated gRPC interface.
type MessengerServer struct {
	pb.UnimplementedMessengerServer
	svc *service.Service
}

// NewMessengerServer builds a MessengerServer backed by svc.
func NewMessengerServer(svc *service.Service) *MessengerServer {
	return &MessengerServer{svc: svc}
}

func (s *MessengerServer) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	resp, err := s.svc.SendMessage(ctx, dto.SendRequest{
		Channel: req.GetChannel(),
		To:      req.GetTo(),
		Text:    req.GetText(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "send message: %v", err)
	}
	return &pb.SendResponse{Id: resp.ID, Status: resp.Status, Error: resp.Error}, nil
}

func (s *MessengerServer) GetMessage(ctx context.Context, req *pb.GetMessageRequest) (*pb.SendResponse, error) {
	msg, err := s.svc.GetMessage(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "message %q not found", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "get message: %v", err)
	}
	return &pb.SendResponse{Id: msg.ID, Status: msg.Status, Error: msg.Error}, nil
}

// Serve starts the gRPC server on addr and blocks until ctx is cancelled or
// an unrecoverable error occurs.
func Serve(ctx context.Context, addr string, svc *service.Service) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc: listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessengerServer(grpcServer, NewMessengerServer(svc))

	errCh := make(chan error, 1)
	go func() {
		log.Printf("grpc: listening on %s", addr)
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
