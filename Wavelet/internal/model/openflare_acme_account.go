// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"errors"
	"time"

	"github.com/Rain-kl/Wavelet/internal/db"
	"gorm.io/gorm"
)

// AcmeAccount OpenFlare ACME 账号实体。
type AcmeAccount struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Email      string    `json:"email" gorm:"size:255"`
	URL        string    `json:"url" gorm:"size:255"`
	PrivateKey string    `json:"-" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名。
func (AcmeAccount) TableName() string {
	return "of_acme_accounts"
}

// GetAcmeAccountByID 按 ID 查询 ACME 账号。
func GetAcmeAccountByID(ctx context.Context, id uint) (*AcmeAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var account AcmeAccount
	if err := conn.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetDefaultAcmeAccount 获取默认 ACME 账号，不存在时创建占位记录。
func GetDefaultAcmeAccount(ctx context.Context) (*AcmeAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var account AcmeAccount
	err := conn.Order("id asc").First(&account).Error
	if err == nil {
		return &account, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	account = AcmeAccount{
		Email: "admin@openflare.dev",
	}
	if err = conn.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}
