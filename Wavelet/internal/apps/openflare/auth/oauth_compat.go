// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"
)

type legacyOIDCProviderCacheType struct {
	mu      sync.RWMutex
	entries map[string]*oidc.Provider
	sfGroup singleflight.Group
}

var legacyOIDCProviderCache = &legacyOIDCProviderCacheType{
	entries: make(map[string]*oidc.Provider),
}

func (c *legacyOIDCProviderCacheType) get(ctx context.Context, issuer string) (*oidc.Provider, error) {
	c.mu.RLock()
	if p, ok := c.entries[issuer]; ok {
		c.mu.RUnlock()
		return p, nil
	}
	c.mu.RUnlock()

	bg := context.Background()
	if client, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok && client != nil {
		bg = oidc.ClientContext(bg, client)
	}

	v, err, _ := c.sfGroup.Do(issuer, func() (any, error) {
		c.mu.RLock()
		if p, ok := c.entries[issuer]; ok {
			c.mu.RUnlock()
			return p, nil
		}
		c.mu.RUnlock()

		p, err := oidc.NewProvider(bg, issuer)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.entries[issuer] = p
		c.mu.Unlock()
		return p, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*oidc.Provider), nil //nolint:forcetypeassert // singleflight value type is fixed
}

// oauthBuildAuthorizeURL builds an OAuth authorize URL using the same rules as apps/oauth.
func oauthBuildAuthorizeURL(ctx context.Context, source *model.AuthSource, redirectURL, state string) (string, error) {
	authConfig, verifier, err := buildOAuthConfig(ctx, source, redirectURL)
	if err != nil {
		return "", err
	}
	if verifier != nil {
		return authConfig.AuthCodeURL(state, oidc.Nonce(state)), nil
	}
	return authConfig.AuthCodeURL(state), nil
}

func buildOAuthConfig(ctx context.Context, source *model.AuthSource, redirectURL string) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	if source == nil {
		return nil, nil, errors.New("认证源不能为空")
	}
	if source.OpenIDDiscoveryURL == "" {
		return nil, nil, errors.New("认证源未配置 OpenID Discovery URL")
	}

	issuer := strings.TrimSuffix(strings.TrimSpace(source.OpenIDDiscoveryURL), "/")
	issuer = strings.TrimSuffix(issuer, "/.well-known/openid-configuration")
	issuer = strings.TrimSuffix(issuer, "/.well-known/oauth-authorization-server")

	provider, err := legacyOIDCProviderCache.get(ctx, issuer)
	if err != nil {
		return nil, nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: source.ClientID})
	scopes := strings.Fields(source.Scopes)
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	if !containsScope(scopes, oidc.ScopeOpenID) {
		scopes = append([]string{oidc.ScopeOpenID}, scopes...)
	}

	return &oauth2.Config{
		ClientID:     source.ClientID,
		ClientSecret: source.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     provider.Endpoint(),
	}, verifier, nil
}

func containsScope(scopes []string, scope string) bool {
	for _, item := range scopes {
		if item == scope {
			return true
		}
	}
	return false
}

func verifyIDToken(ctx context.Context, verifier *oidc.IDTokenVerifier, token *oauth2.Token, nonce string, userInfo *model.OAuthUserInfo) error {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil
	}
	idToken, verifyErr := verifier.Verify(ctx, rawIDToken)
	if verifyErr != nil {
		return fmt.Errorf("ID Token 验证失败: %w", verifyErr)
	}
	if nonce != "" && idToken.Nonce != nonce {
		return errors.New("nonce 不匹配")
	}
	return idToken.Claims(userInfo)
}
