//go:build !debug

package reflection

import (
	"google.golang.org/grpc"
)

func MaybeEnableReflection(s *grpc.Server) {
	// no
}
