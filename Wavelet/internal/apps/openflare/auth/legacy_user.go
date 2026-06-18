// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package auth provides OpenFlare legacy auth business logic.
package auth

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/model"
)

// LegacyUser mirrors the old OpenFlare frontend user shape.
type LegacyUser struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        int    `json:"role"`
	Status      int    `json:"status"`
	Token       string `json:"token,omitempty"`
	Email       string `json:"email,omitempty"`
}

const (
	legacyUserStatusEnabled  = 1
	legacyUserStatusDisabled = 2
)

// RoleFromUser maps a Wavelet user to the legacy role value.
func RoleFromUser(user *model.User) int {
	if user == nil {
		return 0
	}
	if user.IsAdmin {
		return compat.RoleRootUser
	}
	return compat.RoleCommonUser
}

// StatusFromUser maps is_active to legacy status.
func StatusFromUser(user *model.User) int {
	if user == nil || !user.IsActive {
		return legacyUserStatusDisabled
	}
	return legacyUserStatusEnabled
}

// ToLegacyUser converts a Wavelet user to the legacy response shape.
func ToLegacyUser(user *model.User, token string) LegacyUser {
	if user == nil {
		return LegacyUser{}
	}
	return LegacyUser{
		ID:          int(user.ID),
		Username:    user.Username,
		DisplayName: displayName(user),
		Role:        RoleFromUser(user),
		Status:      StatusFromUser(user),
		Token:       token,
		Email:       user.Email,
	}
}

// ToLegacyUsers converts a slice of users.
func ToLegacyUsers(users []model.User) []LegacyUser {
	result := make([]LegacyUser, 0, len(users))
	for i := range users {
		result = append(result, ToLegacyUser(&users[i], ""))
	}
	return result
}

func displayName(user *model.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	return user.Username
}

// IsAdminRole reports whether a legacy role has admin privileges.
func IsAdminRole(role int) bool {
	return role >= compat.RoleAdminUser
}
