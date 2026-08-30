// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package publicconfig implements contracts.PublicConfigProvider for OpenFlare.
package publicconfig

import (
	"context"

	"Wavelet/core"
	adminrepo "Wavelet/plugins/domain/admin/repository"
	"Wavelet/plugins/infra/database"
)

// Provider returns visibility=1 system configs as a flat key/value map,
// matching gold GetPublicConfig.
type Provider struct{}

// New constructs a PublicConfigProvider. ctx is accepted for future binding
// but the payload is loaded from the shared system config store.
func New(_ *core.Context) *Provider {
	return &Provider{}
}

// PublicConfig returns visibility=1 keys as map[string]string.
func (p *Provider) PublicConfig(ctx context.Context) (any, error) {
	if adminrepo.GetDB(ctx) != nil {
		return listVisible(ctx)
	}
	return listVisibleFromGORM(ctx)
}

func listVisible(ctx context.Context) (map[string]string, error) {
	configs, err := adminrepo.ListVisibleSystemConfigs(ctx)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]string, len(configs))
	for _, config := range configs {
		resp[config.Key] = config.Value
	}
	return resp, nil
}

func listVisibleFromGORM(ctx context.Context) (map[string]string, error) {
	conn := database.DB(ctx)
	if conn == nil {
		return map[string]string{}, nil
	}
	type row struct {
		Key   string
		Value string
	}
	var rows []row
	if err := conn.Table("w_system_configs").
		Select("key, value").
		Where("visibility = ?", 1).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	resp := make(map[string]string, len(rows))
	for _, r := range rows {
		resp[r.Key] = r.Value
	}
	return resp, nil
}
