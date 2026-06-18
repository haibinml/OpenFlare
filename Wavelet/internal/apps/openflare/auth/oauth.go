// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/db"
	"github.com/Rain-kl/Wavelet/internal/listener"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/pkg/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const pendingExternalAccountSessionKey = "pending_external_account"

// OAuthCallbackResult is the legacy OAuth callback payload.
type OAuthCallbackResult struct {
	Status string      `json:"status"`
	User   *LegacyUser `json:"user,omitempty"`
}

// PendingExternalAccount stores OAuth bind-pending state in session.
type PendingExternalAccount struct {
	AuthSourceID     uint64 `json:"auth_source_id"`
	ExternalID       string `json:"external_id"`
	ExternalUsername string `json:"external_username"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
}

// OAuthAuthorize builds an authorize URL for a legacy auth source route param.
func OAuthAuthorize(ctx context.Context, c *gin.Context, sourceKey string) (string, error) {
	source, err := resolveAuthSourceByRoute(ctx, sourceKey)
	if err != nil {
		return "", err
	}
	if !source.IsActive {
		return "", errors.New(errAuthSourceDisabled)
	}
	if err := source.Validate(); err != nil {
		return "", err
	}

	state := uuid.NewString()
	session := sessions.Default(c)
	session.Set(oauthStateSessionKey(source.ID), state)
	if err := session.Save(); err != nil {
		return "", errors.New(errSaveSessionFailed)
	}

	redirectURL := legacyOAuthCallbackURL(c, source)
	return buildLegacyAuthorizeURL(ctx, source, redirectURL, state)
}

// OAuthCallback handles GET /oauth/:source/callback for the legacy frontend.
func OAuthCallback(ctx context.Context, c *gin.Context, sourceKey string) (OAuthCallbackResult, error) {
	source, err := resolveAuthSourceByRoute(ctx, sourceKey)
	if err != nil {
		return OAuthCallbackResult{}, err
	}
	if !source.IsActive {
		return OAuthCallbackResult{}, errors.New(errAuthSourceDisabled)
	}

	session := sessions.Default(c)
	expectedState, _ := session.Get(oauthStateSessionKey(source.ID)).(string)
	state := c.Query("state")
	if expectedState == "" || state == "" || state != expectedState {
		return OAuthCallbackResult{}, errors.New("授权状态无效，请重新登录")
	}
	session.Delete(oauthStateSessionKey(source.ID))
	if err := session.Save(); err != nil {
		return OAuthCallbackResult{}, errors.New(errSaveSessionFailed)
	}
	if oauthError := c.Query("error"); oauthError != "" {
		description := c.Query("error_description")
		if description == "" {
			description = oauthError
		}
		return OAuthCallbackResult{}, errors.New(description)
	}

	redirectURL := legacyOAuthCallbackURL(c, source)
	userInfo, err := exchangeLegacyOAuthProfile(ctx, source, c.Query("code"), state, redirectURL)
	if err != nil {
		return OAuthCallbackResult{}, err
	}

	var currentUserID *uint64
	if current := currentUserFromLegacyToken(ctx, c); current != nil {
		currentUserID = &current.ID
	}

	result, pending, err := completeLegacyOAuthLogin(ctx, source, userInfo, currentUserID)
	if err != nil {
		return OAuthCallbackResult{}, err
	}
	if pending != nil {
		raw, marshalErr := json.Marshal(pending)
		if marshalErr != nil {
			return OAuthCallbackResult{}, marshalErr
		}
		session.Set(pendingExternalAccountSessionKey, string(raw))
		if err := session.Save(); err != nil {
			return OAuthCallbackResult{}, errors.New(errSaveSessionFailed)
		}
		return result, nil
	}
	if result.User != nil {
		var dbUser model.User
		if err := db.DB(ctx).Where("id = ?", result.User.ID).First(&dbUser).Error; err != nil {
			return OAuthCallbackResult{}, err
		}
		if err := setLoginSession(ctx, c, &dbUser); err != nil {
			return OAuthCallbackResult{}, errors.New(errSaveSessionFailed)
		}
		token, tokenErr := issueLegacyAccessToken(ctx, &dbUser)
		if tokenErr != nil {
			return OAuthCallbackResult{}, tokenErr
		}
		legacy := ToLegacyUser(&dbUser, token)
		result.User = &legacy
		listener.EmitAdminLoggedIn(ctx, &dbUser, c.ClientIP())
	}
	return result, nil
}

// LinkExistingOAuthAccount binds a pending external account to an existing user.
func LinkExistingOAuthAccount(ctx context.Context, c *gin.Context, input LinkExistingInput) (OAuthCallbackResult, error) {
	session := sessions.Default(c)
	raw, _ := session.Get(pendingExternalAccountSessionKey).(string)
	if raw == "" {
		return OAuthCallbackResult{}, errors.New(errPendingOAuthExpired)
	}
	var pending PendingExternalAccount
	if err := json.Unmarshal([]byte(raw), &pending); err != nil {
		return OAuthCallbackResult{}, errors.New(errPendingOAuthInvalid)
	}

	user, err := linkPendingExternalAccount(ctx, &pending, input)
	if err != nil {
		return OAuthCallbackResult{}, err
	}

	session.Delete(pendingExternalAccountSessionKey)
	if err := session.Save(); err != nil {
		return OAuthCallbackResult{}, errors.New(errSaveSessionFailed)
	}
	if err := setLoginSession(ctx, c, user); err != nil {
		return OAuthCallbackResult{}, errors.New(errSaveSessionFailed)
	}
	token, err := issueLegacyAccessToken(ctx, user)
	if err != nil {
		return OAuthCallbackResult{}, err
	}
	legacy := ToLegacyUser(user, token)
	return OAuthCallbackResult{Status: "linked", User: &legacy}, nil
}

func resolveAuthSourceByRoute(ctx context.Context, raw string) (*model.AuthSource, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("认证源不能为空")
	}
	if parsed, err := parseUint64(raw); err == nil && parsed > 0 {
		return model.GetAuthSourceByID(ctx, parsed)
	}
	return model.GetAuthSourceByName(ctx, raw)
}

func oauthStateSessionKey(sourceID uint64) string {
	return fmt.Sprintf("oauth_state_%d", sourceID)
}

func legacyOAuthCallbackURL(c *gin.Context, source *model.AuthSource) string {
	ctx := c.Request.Context()
	base := ""
	if sc, err := repository.GetSystemConfigByKey(ctx, model.ConfigKeyServerAddress); err == nil {
		base = strings.TrimRight(sc.Value, "/")
	}
	if base == "" {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
			host = forwardedHost
		}
		base = scheme + "://" + host
	}
	sourceName := source.Name
	if sourceName == "" {
		sourceName = fmt.Sprintf("%d", source.ID)
	}
	callback, _ := url.JoinPath(base, "oauth", sourceName)
	return callback
}

func buildLegacyAuthorizeURL(ctx context.Context, source *model.AuthSource, redirectURL, state string) (string, error) {
	payloadValue, err := encodeLegacyOAuthState(source.Name, state)
	if err != nil {
		return "", err
	}
	stateKey := fmt.Sprintf("of_oauth_state:%s", state)
	if err := db.Redis.Set(ctx, db.PrefixedKey(stateKey), payloadValue, 10*time.Minute).Err(); err != nil {
		return "", err
	}
	return oauthBuildAuthorizeURL(ctx, source, redirectURL, state)
}

func encodeLegacyOAuthState(sourceName, state string) (string, error) {
	payload := map[string]string{
		"source_name": sourceName,
		"state":       state,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func legacyFrontendLoginRedirectURL(ctx context.Context, source *model.AuthSource) (string, error) {
	sc, err := repository.GetSystemConfigByKey(ctx, model.ConfigKeyServerAddress)
	if err != nil || strings.TrimSpace(sc.Value) == "" {
		return "", errors.New("server_address 未配置")
	}
	base := strings.TrimRight(sc.Value, "/")
	name := source.Name
	if name == "" {
		name = fmt.Sprintf("%d", source.ID)
	}
	return base + "/oauth/" + url.PathEscape(name), nil
}

func exchangeLegacyOAuthProfile(ctx context.Context, source *model.AuthSource, code, state, redirectURL string) (*model.OAuthUserInfo, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("授权 code 不能为空")
	}
	// Validate state from Redis cache written during authorize.
	stateKey := fmt.Sprintf("of_oauth_state:%s", state)
	payloadRaw, err := db.Redis.Get(ctx, db.PrefixedKey(stateKey)).Result()
	if err != nil {
		return nil, errors.New("授权状态无效，请重新登录")
	}
	_ = db.Redis.Del(ctx, db.PrefixedKey(stateKey)).Err()

	var payload map[string]string
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		return nil, err
	}
	if payload["source_name"] != source.Name {
		return nil, errors.New("授权状态无效，请重新登录")
	}

	userInfo, err := buildOAuthUserInfo(ctx, source, code, state, redirectURL)
	if err != nil {
		return nil, err
	}
	if err := normalizeOAuthUserInfo(userInfo); err != nil {
		return nil, err
	}
	if userInfo.Sub == "" {
		userInfo.Sub = userInfo.Username
	}
	return userInfo, nil
}

func completeLegacyOAuthLogin(ctx context.Context, source *model.AuthSource, profile *model.OAuthUserInfo, currentUserID *uint64) (OAuthCallbackResult, *PendingExternalAccount, error) {
	if source == nil || profile == nil || strings.TrimSpace(profile.Sub) == "" {
		return OAuthCallbackResult{}, nil, errors.New("第三方账号资料不完整")
	}

	account, err := model.FindExternalAccount(ctx, source.ID, profile.Sub)
	if err == nil {
		var user model.User
		if err := db.DB(ctx).Where("id = ?", account.UserID).First(&user).Error; err != nil {
			return OAuthCallbackResult{}, nil, err
		}
		if !user.IsActive {
			return OAuthCallbackResult{}, nil, errors.New(errBannedAccount)
		}
		legacy := ToLegacyUser(&user, "")
		return OAuthCallbackResult{Status: "logged_in", User: &legacy}, nil, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return OAuthCallbackResult{}, nil, err
	}

	if currentUserID != nil && *currentUserID > 0 {
		var user model.User
		if err := db.DB(ctx).Where("id = ?", *currentUserID).First(&user).Error; err != nil {
			return OAuthCallbackResult{}, nil, err
		}
		if !user.IsActive {
			return OAuthCallbackResult{}, nil, errors.New(errBannedAccount)
		}
		if err := model.BindExternalAccount(ctx, &model.ExternalAccount{
			AuthSourceID:     source.ID,
			UserID:           user.ID,
			ExternalID:       profile.Sub,
			ExternalUsername: profile.Username,
			Email:            profile.Email,
		}); err != nil {
			return OAuthCallbackResult{}, nil, err
		}
		legacy := ToLegacyUser(&user, "")
		return OAuthCallbackResult{Status: "linked", User: &legacy}, nil, nil
	}

	registrationEnabled, regErr := repository.GetBoolByKey(ctx, model.ConfigKeyRegistrationEnabled)
	if regErr != nil {
		registrationEnabled = true
	}
	if !registrationEnabled {
		pending := &PendingExternalAccount{
			AuthSourceID:     source.ID,
			ExternalID:       profile.Sub,
			ExternalUsername: profile.Username,
			DisplayName:      profile.Name,
			Email:            profile.Email,
		}
		return OAuthCallbackResult{Status: "link_required"}, pending, nil
	}

	user, err := createUserFromOAuthProfile(ctx, source, profile)
	if err != nil {
		return OAuthCallbackResult{}, nil, err
	}
	legacy := ToLegacyUser(&user, "")
	return OAuthCallbackResult{Status: "logged_in", User: &legacy}, nil, nil
}

func createUserFromOAuthProfile(ctx context.Context, source *model.AuthSource, profile *model.OAuthUserInfo) (model.User, error) {
	username, err := uniqueLegacyUsername(ctx, profile.Username)
	if err != nil {
		return model.User{}, err
	}
	profile.Username = username

	var user model.User
	if err := user.CreateUser(ctx, db.DB(ctx), profile); err != nil {
		return model.User{}, err
	}
	if err := model.BindExternalAccount(ctx, &model.ExternalAccount{
		AuthSourceID:     source.ID,
		UserID:           user.ID,
		ExternalID:       profile.Sub,
		ExternalUsername: profile.Username,
		Email:            profile.Email,
	}); err != nil {
		return model.User{}, err
	}
	logger.InfoF(ctx, "[LoginAudit] successful legacy OAuth registration via source: %s, user: %s, ID: %d", source.Name, user.Username, user.ID)
	return user, nil
}

func linkPendingExternalAccount(ctx context.Context, pending *PendingExternalAccount, input LinkExistingInput) (*model.User, error) {
	if pending == nil || pending.AuthSourceID == 0 || pending.ExternalID == "" {
		return nil, errors.New(errPendingOAuthExpired)
	}
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || input.Password == "" {
		return nil, errors.New(errInvalidParams)
	}

	var user model.User
	if err := db.DB(ctx).Where("username = ? OR email = ?", input.Username, input.Username).First(&user).Error; err != nil {
		return nil, errors.New(errUsernameOrPasswordWrong)
	}
	if !user.IsActive {
		return nil, errors.New(errBannedAccount)
	}
	if !user.CheckPassword(input.Password) {
		return nil, errors.New(errUsernameOrPasswordWrong)
	}

	if existing, err := model.FindExternalAccount(ctx, pending.AuthSourceID, pending.ExternalID); err == nil {
		if existing.UserID != user.ID {
			return nil, errors.New("该第三方账号已绑定其他用户")
		}
		return &user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := model.BindExternalAccount(ctx, &model.ExternalAccount{
		AuthSourceID:     pending.AuthSourceID,
		UserID:           user.ID,
		ExternalID:       pending.ExternalID,
		ExternalUsername: pending.ExternalUsername,
		Email:            pending.Email,
	}); err != nil {
		return nil, err
	}
	return &user, nil
}

func currentUserFromLegacyToken(ctx context.Context, c *gin.Context) *model.User {
	token := strings.TrimSpace(c.GetHeader(compat.OpenFlareTokenHeader()))
	if token == "" {
		return nil
	}
	tokenHash := model.HashToken(token)
	var record model.AccessToken
	if err := db.DB(ctx).Where("token_hash = ?", tokenHash).First(&record).Error; err != nil {
		return nil
	}
	var user model.User
	if err := db.DB(ctx).Where("id = ? AND is_active = ?", record.UserID, true).First(&user).Error; err != nil {
		return nil
	}
	return &user
}

func uniqueLegacyUsername(ctx context.Context, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "user"
	}
	candidate := base
	for i := 0; i <= 1000; i++ {
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}
		count, err := repository.CountUsersByUsername(ctx, candidate)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", errors.New("无法生成唯一用户名")
}

func parseUint64(raw string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(raw, "%d", &id)
	return id, err
}

func isOIDCLoginEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyOIDCLoginEnabled)
	return err != nil || enabled
}

// The following functions mirror oauth package internals for legacy GET callback support.
// They intentionally duplicate minimal logic to avoid modifying the core oauth module.

func buildOAuthUserInfo(ctx context.Context, source *model.AuthSource, code, nonce, redirectURL string) (*model.OAuthUserInfo, error) {
	authConfig, verifier, err := buildOAuthConfig(ctx, source, redirectURL)
	if err != nil {
		return nil, err
	}
	token, err := authConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	userInfo := &model.OAuthUserInfo{Active: true}
	if verifier != nil {
		if verifyErr := verifyIDToken(ctx, verifier, token, nonce, userInfo); verifyErr != nil {
			return nil, verifyErr
		}
	}
	if userInfo.Username == "" && userInfo.PreferredUsername != "" {
		userInfo.Username = userInfo.PreferredUsername
	}
	if userInfo.Username == "" && userInfo.Email != "" {
		userInfo.Username = strings.Split(userInfo.Email, "@")[0]
	}
	if userInfo.Username == "" && userInfo.Sub != "" {
		userInfo.Username = userInfo.Sub
	}
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Username
	}
	return userInfo, nil
}

func normalizeOAuthUserInfo(userInfo *model.OAuthUserInfo) error {
	userInfo.Username = strings.TrimSpace(userInfo.Username)
	userInfo.Email = strings.TrimSpace(userInfo.Email)
	userInfo.Name = strings.TrimSpace(userInfo.Name)
	if userInfo.Username == "" {
		return errors.New("无法从认证源获取用户名")
	}
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Username
	}
	if !userInfo.Active {
		userInfo.Active = true
	}
	return nil
}
