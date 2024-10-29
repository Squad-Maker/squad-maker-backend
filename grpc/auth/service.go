package auth

import (
	pb "squad-maker/generated/auth"
)

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func NewAuthServiceServer() *AuthServiceServer {
	return &AuthServiceServer{}
}
