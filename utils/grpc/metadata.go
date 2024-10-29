package grpcUtils

import (
	"context"
	"strconv"

	pbAuth "squad-maker/generated/auth"

	"google.golang.org/grpc/metadata"
)

func GetCurrentUserIdFromMetadata(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}

	ids := md["user-id"]
	if len(ids) == 0 {
		return 0
	}

	id, _ := strconv.ParseInt(ids[len(ids)-1], 10, 64)
	return id
}

func GetCurrentSessionTokenFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md["session-token"]
	if len(values) == 0 {
		return ""
	}

	return values[len(values)-1]
}

func GetCurrentSessionIdFromMetadata(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}

	ids := md["session-id"]
	if len(ids) == 0 {
		return 0
	}

	id, _ := strconv.ParseInt(ids[len(ids)-1], 10, 64)
	return id
}

func GetCurrentUserTypeFromMetadata(ctx context.Context) pbAuth.UserType {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return pbAuth.UserType_utStudent
	}

	values := md["user-type"]
	if len(values) == 0 {
		return pbAuth.UserType_utStudent
	}

	v, err := strconv.ParseInt(values[len(values)-1], 10, 64)
	if err != nil {
		return pbAuth.UserType_utStudent
	}

	return pbAuth.UserType(v)
}
