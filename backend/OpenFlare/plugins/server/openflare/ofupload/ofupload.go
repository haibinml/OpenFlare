// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package ofupload wraps Wavelet upload ingest with Pages-specific helpers.
package ofupload

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	waveletupload "Wavelet/plugins/domain/upload"
	"Wavelet/plugins/domain/upload/cache"
	"Wavelet/plugins/domain/upload/models"
	uploadrepo "Wavelet/plugins/domain/upload/repository"
	uploadstats "Wavelet/plugins/domain/upload/stats"
	uploadstorage "Wavelet/plugins/domain/upload/storage"
	"Wavelet/plugins/infra/database"

	"gorm.io/gorm"
)

// ReservedPagesDeploymentType is managed exclusively by the Pages domain.
const ReservedPagesDeploymentType = "openflare_pages_deployment"

const (
	// PolicyCreate always stores a new object and creates a new upload record.
	PolicyCreate = waveletupload.PolicyCreate
	// PolicyDedupNewRecord reuses an existing object path on hash match but creates a new record.
	PolicyDedupNewRecord = waveletupload.PolicyDedupNewRecord
	// PolicyResolveExisting returns an existing upload on hash match.
	PolicyResolveExisting = waveletupload.PolicyResolveExisting
)

type (
	// IngestRequest is the programmatic upload ingest payload.
	IngestRequest = waveletupload.IngestRequest
	// IngestResult reports ingest side effects.
	IngestResult = waveletupload.IngestResult
	// IngestPolicy controls hash-collision behavior during ingest.
	IngestPolicy = waveletupload.IngestPolicy
)

// IngestFromLocalPath ingests a local regular file through Wavelet upload ingest.
func IngestFromLocalPath(ctx context.Context, localPath string, req IngestRequest) (IngestResult, error) {
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return IngestResult{}, errors.New("local path is required")
	}
	file, err := os.Open(localPath) //nolint:gosec // localPath is resolved from managed Pages artifacts
	if err != nil {
		return IngestResult{}, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return IngestResult{}, err
	}
	if info.IsDir() {
		return IngestResult{}, errors.New("local path must be a regular file")
	}
	if req.Size <= 0 {
		req.Size = info.Size()
	}
	req.Reader = file
	return waveletupload.Ingest(ctx, req)
}

// RemoveLockedTx performs the idempotent active-to-deleted transition for a row
// that the caller has already locked in its surrounding transaction.
func RemoveLockedTx(tx *gorm.DB, upload *models.Upload) (bool, error) {
	if upload == nil {
		return false, nil
	}
	if upload.Status == models.UploadStatusDeleted {
		return false, nil
	}
	snapshot := *upload
	if err := uploadrepo.SoftDeleteUploadTx(tx, upload); err != nil {
		return false, err
	}
	if err := uploadstats.ApplyUploadStatsDeltaTx(tx, &snapshot, -1); err != nil {
		return false, err
	}
	upload.Status = models.UploadStatusDeleted
	return true, nil
}

// InvalidateUploadMetaCache evicts cached upload metadata.
func InvalidateUploadMetaCache(ctx context.Context, id uint64) {
	cache.EvictUploadMeta(ctx, id)
}

// GetActiveUpload loads an active (non-deleted) upload by ID.
func GetActiveUpload(ctx context.Context, id uint64) (models.Upload, error) {
	if u, err := uploadrepo.GetActiveUploadByID(ctx, id); err == nil {
		return u, nil
	}
	conn := database.DB(ctx)
	if conn == nil {
		return models.Upload{}, errors.New("database not initialized")
	}
	var upload models.Upload
	err := conn.Where("id = ? AND status <> ?", id, models.UploadStatusDeleted).First(&upload).Error
	return upload, err
}

// OpenedUploadObject is a stored object stream plus the upload record.
type OpenedUploadObject struct {
	Upload        models.Upload
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// OpenStoredUpload opens the stored object for an active upload.
func OpenStoredUpload(ctx context.Context, id uint64) (*OpenedUploadObject, error) {
	upload, err := GetActiveUpload(ctx, id)
	if err != nil {
		return nil, err
	}
	obj, err := uploadstorage.OpenStoredObject(ctx, &upload)
	if err != nil {
		return nil, err
	}
	return &OpenedUploadObject{
		Upload:        upload,
		Body:          obj.Body,
		ContentType:   obj.ContentType,
		ContentLength: obj.ContentLength,
	}, nil
}

// LocalFileCandidateRequest describes filesystem locations that may host a legacy blob.
type LocalFileCandidateRequest struct {
	StoredPath    string
	RelativePaths []string
}

// ResolveLocalFile returns the first existing regular file among candidate paths.
func ResolveLocalFile(_ context.Context, req LocalFileCandidateRequest) (string, int64, error) {
	candidates := append([]string{req.StoredPath}, req.RelativePaths...)
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate) //nolint:gosec // candidate is resolved from managed Pages metadata
		if err != nil || info.IsDir() {
			continue
		}
		return candidate, info.Size(), nil
	}
	return "", 0, os.ErrNotExist
}

// RebuildUploadStats rebuilds aggregate upload stats.
func RebuildUploadStats(ctx context.Context) error {
	return uploadstats.RebuildUploadStats(ctx)
}
