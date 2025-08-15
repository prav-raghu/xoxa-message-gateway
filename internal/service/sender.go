// Package service contains core send logic
package service

import (
	"xoxa-message-gateway/pkg/dto"
)

func SendMessage(req dto.SendRequest) dto.SendResponse {
	// ...core send logic...
	return dto.SendResponse{Status: "sent"}
}

func SendMessageGRPC(ctx context.Context, req *proto.SendRequest) (*proto.SendResponse, error) {
	// ...core send logic for gRPC...
	return &proto.SendResponse{Status: "sent"}, nil
}
