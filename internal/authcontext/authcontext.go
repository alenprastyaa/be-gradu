package authcontext

import (
	"context"

	"github.com/google/uuid"
)

const (
	RoleSuperAdmin  = "super_admin"
	RoleSchoolAdmin = "school_admin"
)

type AuthInfo struct {
	AdminID  uuid.UUID
	Email    string
	Role     string
	SchoolID *uuid.UUID
}

type contextKey struct{}

func WithAuth(ctx context.Context, info AuthInfo) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

func FromContext(ctx context.Context) (AuthInfo, bool) {
	info, ok := ctx.Value(contextKey{}).(AuthInfo)
	return info, ok
}

func SchoolID(ctx context.Context) (*uuid.UUID, bool) {
	info, ok := FromContext(ctx)
	if !ok || info.SchoolID == nil || *info.SchoolID == uuid.Nil {
		return nil, false
	}
	return info.SchoolID, true
}

func WithSchoolID(ctx context.Context, schoolID uuid.UUID) context.Context {
	info, _ := FromContext(ctx)
	info.SchoolID = &schoolID
	if info.Role == "" {
		info.Role = RoleSchoolAdmin
	}
	return WithAuth(ctx, info)
}

func IsSuperAdmin(ctx context.Context) bool {
	info, ok := FromContext(ctx)
	return ok && info.Role == RoleSuperAdmin
}
