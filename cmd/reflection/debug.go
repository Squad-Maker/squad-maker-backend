//go:build debug

package reflection

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func MaybeEnableReflection(s *grpc.Server) {
	reflection.Register(s)
}
