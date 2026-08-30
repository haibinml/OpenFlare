// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	adminmodel "Wavelet/plugins/domain/admin/model"
	authmodel "Wavelet/plugins/domain/auth"
	usermodel "Wavelet/plugins/domain/user"
	uploadmodels "Wavelet/plugins/domain/upload/models"
)

const (
	tokenByteLength = 24
	maskThreshold   = 8
)

// User is the Wavelet w_users entity.
type User = usermodel.User

// AccessToken is the Wavelet w_access_tokens entity.
type AccessToken = usermodel.AccessToken

// AuthSource is the Wavelet w_auth_sources entity.
type AuthSource = authmodel.AuthSource

// ExternalAccount is the Wavelet w_external_accounts entity.
type ExternalAccount = authmodel.ExternalAccount

// TaskExecution is the Wavelet w_task_executions entity.
type TaskExecution = adminmodel.TaskExecution

// Template is the Wavelet w_templates entity.
type Template = adminmodel.Template

// Schedule is the Wavelet w_schedules entity.
type Schedule = adminmodel.Schedule

// Upload is the Wavelet w_uploads entity.
type Upload = uploadmodels.Upload

// UploadMetadata is the Wavelet upload metadata JSON.
type UploadMetadata = uploadmodels.UploadMetadata

// UploadStatus is the Wavelet upload status.
type UploadStatus = uploadmodels.UploadStatus

// UploadStat is the Wavelet w_upload_stats entity.
type UploadStat = uploadmodels.UploadStat

const (
	// UploadStatusPending is a newly stored unused upload.
	UploadStatusPending = uploadmodels.UploadStatusPending
	// UploadStatusUsed is an in-use upload.
	UploadStatusUsed = uploadmodels.UploadStatusUsed
	// UploadStatusDeleted is a soft-deleted upload.
	UploadStatusDeleted = uploadmodels.UploadStatusDeleted

	// UploadStatDimensionTotal is the total stats dimension.
	UploadStatDimensionTotal = uploadmodels.UploadStatDimensionTotal
	// UploadStatDimensionType is the type stats dimension.
	UploadStatDimensionType = uploadmodels.UploadStatDimensionType
	// UploadStatDimensionCategory is the category stats dimension.
	UploadStatDimensionCategory = uploadmodels.UploadStatDimensionCategory
	// UploadStatDimensionTrend is the trend stats dimension.
	UploadStatDimensionTrend = uploadmodels.UploadStatDimensionTrend
)

// GenerateTokenString 生成加密安全的随机 Token 值
func GenerateTokenString() (string, error) {
	bytes := make([]byte, tokenByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "at_" + hex.EncodeToString(bytes), nil
}

// HashToken 计算 Token 的 SHA-256 哈希值用于数据库存储与查询
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// MaskTokenString 生成脱敏显示的 Token，仅保留前缀和最后四位
func MaskTokenString(token string) string {
	if len(token) <= maskThreshold {
		return "at_****"
	}
	return fmt.Sprintf("%s...%s", token[:7], token[len(token)-4:])
}
