// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"context"

	"Wavelet/core/contracts"

	"github.com/gin-gonic/gin"
)

// StubAuth is a contracts.AuthService that admits every request.
type StubAuth struct {
	User    *contracts.UserDTO
	Sources []contracts.AuthSourceViewDTO
}

var _ contracts.AuthService = StubAuth{}

func passThrough() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func (s StubAuth) RequireAuthMiddleware() any  { return passThrough() }
func (s StubAuth) RequireAdminMiddleware() any { return passThrough() }
func (s StubAuth) DisallowTokenAuthMiddleware() any {
	return passThrough()
}

func (s StubAuth) GetCurrentUser(context.Context) (*contracts.UserDTO, error) {
	return s.User, nil
}
func (s StubAuth) GetCurrentUserID(context.Context) (uint64, error) {
	if s.User == nil {
		return 0, nil
	}
	return s.User.ID, nil
}
func (s StubAuth) VerifyToken(context.Context, string) (*contracts.UserDTO, error) {
	return s.User, nil
}
func (s StubAuth) CreateSession(context.Context, uint64, map[string]any) (string, error) {
	return "", nil
}
func (s StubAuth) RevokeToken(context.Context, string) error        { return nil }
func (s StubAuth) RevokeUserSessions(context.Context, uint64) error { return nil }
func (s StubAuth) InvalidateCachedUser(context.Context, uint64)     {}
func (s StubAuth) InvalidateCachedToken(context.Context, string)    {}
func (s StubAuth) ListAuthSources(context.Context) ([]contracts.AuthSourceViewDTO, error) {
	return s.Sources, nil
}
func (s StubAuth) CreateAuthSource(context.Context, contracts.AuthSourceDTO) (*contracts.AuthSourceDTO, error) {
	return nil, nil
}
func (s StubAuth) UpdateAuthSource(context.Context, uint64, contracts.AuthSourceDTO) (*contracts.AuthSourceDTO, error) {
	return nil, nil
}
func (s StubAuth) DeleteAuthSource(context.Context, uint64) error { return nil }
func (s StubAuth) ToggleAuthSource(context.Context, uint64) (*contracts.AuthSourceDTO, error) {
	return nil, nil
}
