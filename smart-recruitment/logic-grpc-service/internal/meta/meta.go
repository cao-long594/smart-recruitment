package meta

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	MDUserID   = "x-user-id"
	MDUserRole = "x-user-role"
)

type User struct {
	ID   int64
	Role string // hr | candidate
}

func IncomingUser(ctx context.Context) (User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return User{}, status.Error(codes.Unauthenticated, "missing metadata")
	}
	ids := md.Get(MDUserID)
	roles := md.Get(MDUserRole)
	if len(ids) == 0 || len(roles) == 0 {
		return User{}, status.Error(codes.Unauthenticated, "missing user context")
	}
	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return User{}, status.Error(codes.Unauthenticated, "invalid user id")
	}
	return User{ID: id, Role: roles[0]}, nil
}

func MustHR(u User) error {
	if u.Role != "hr" {
		return status.Error(codes.PermissionDenied, "hr only")
	}
	return nil
}

func MustCandidate(u User) error {
	if u.Role != "candidate" {
		return status.Error(codes.PermissionDenied, "candidate only")
	}
	return nil
}
