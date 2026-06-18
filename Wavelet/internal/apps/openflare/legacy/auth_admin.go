// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"fmt"
	"strconv"
	"strings"

	ofauth "github.com/Rain-kl/Wavelet/internal/apps/openflare/auth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/gin-gonic/gin"
)

type legacyUserPayload struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        int    `json:"role"`
	Email       string `json:"email"`
}

type manageUserRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

type authSourcePayload struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	DisplayName        string `json:"display_name"`
	IsActive           bool   `json:"is_active"`
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	OpenIDDiscoveryURL string `json:"openid_discovery_url"`
	Scopes             string `json:"scopes"`
	IconURL            string `json:"icon_url"`
}

type authSourceTogglePayload struct {
	IsActive bool `json:"is_active"`
}

func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("p"))
	users, err := ofauth.ListUsers(c.Request.Context(), page)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, users)
}

func SearchUsers(c *gin.Context) {
	users, err := ofauth.SearchUsers(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, users)
}

func GetUser(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	user, err := ofauth.GetUserByID(c.Request.Context(), callerRole(c), uint64(id))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, user)
}

func CreateUser(c *gin.Context) {
	var req legacyUserPayload
	if !compat.BindJSON(c, &req) {
		return
	}
	if err := ofauth.CreateUser(c.Request.Context(), callerRole(c), ofauth.CreateUserInput{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Role:        req.Role,
		Email:       req.Email,
	}); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func UpdateUser(c *gin.Context) {
	var req legacyUserPayload
	if !compat.BindJSON(c, &req) {
		return
	}
	if err := ofauth.UpdateUser(c.Request.Context(), callerRole(c), ofauth.UpdateUserInput{
		ID:          req.ID,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Role:        req.Role,
		Email:       req.Email,
	}); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func DeleteUser(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := ofauth.DeleteUserByID(c.Request.Context(), callerRole(c), uint64(id)); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func ManageUser(c *gin.Context) {
	var req manageUserRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	user, err := ofauth.ManageUser(c.Request.Context(), callerRole(c), ofauth.ManageUserInput{
		Username: req.Username,
		Action:   req.Action,
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, user)
}

func ListAuthSources(c *gin.Context) {
	sources, err := model.GetAuthSources(c.Request.Context())
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, sources)
}

func CreateAuthSource(c *gin.Context) {
	var payload authSourcePayload
	if !compat.BindJSON(c, &payload) {
		return
	}
	source := payload.toModel()
	if err := model.CreateAuthSource(c.Request.Context(), &source); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	source.Sanitize()
	compat.OK(c, source)
}

func UpdateAuthSource(c *gin.Context) {
	id, err := parseAuthSourceID(c)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	var payload authSourcePayload
	if !compat.BindJSON(c, &payload) {
		return
	}
	source := payload.toModel()
	source.ID = id
	keepSecret := strings.TrimSpace(source.ClientSecret) == ""
	if err := model.UpdateAuthSource(c.Request.Context(), &source, keepSecret); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	updated, err := model.GetAuthSourceByID(c.Request.Context(), id)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	updated.Sanitize()
	compat.OK(c, updated)
}

func DeleteAuthSource(c *gin.Context) {
	id, err := parseAuthSourceID(c)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	if err := model.DeleteAuthSource(c.Request.Context(), id); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func ToggleAuthSource(c *gin.Context) {
	id, err := parseAuthSourceID(c)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	var payload authSourceTogglePayload
	if !compat.BindJSON(c, &payload) {
		return
	}
	if err := model.ToggleAuthSource(c.Request.Context(), id, payload.IsActive); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func (payload authSourcePayload) toModel() model.AuthSource {
	return model.AuthSource{
		Name:               payload.Name,
		Type:               payload.Type,
		DisplayName:        payload.DisplayName,
		IsActive:           payload.IsActive,
		ClientID:           payload.ClientID,
		ClientSecret:       payload.ClientSecret,
		OpenIDDiscoveryURL: payload.OpenIDDiscoveryURL,
		Scopes:             payload.Scopes,
		IconURL:            payload.IconURL,
	}
}

func parseAuthSourceID(c *gin.Context) (uint64, error) {
	raw := strings.TrimSpace(c.Param("id"))
	if raw == "" {
		return 0, fmt.Errorf("认证源 ID 无效")
	}
	source, err := model.GetAuthSourceByName(c.Request.Context(), raw)
	if err == nil {
		return source.ID, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("认证源 ID 无效")
	}
	return parsed, nil
}
