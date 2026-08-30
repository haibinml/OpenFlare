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
	"sync"

	"Wavelet/core/contracts"
	waveletupload "Wavelet/plugins/domain/upload"
	"Wavelet/plugins/domain/upload/models"
	"Wavelet/plugins/infra/database"
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

var (
	storageMu  sync.RWMutex
	storageSvc contracts.StorageService
)

// SetStorage injects the platform StorageService used to open stored objects.
func SetStorage(s contracts.StorageService) {
	storageMu.Lock()
	defer storageMu.Unlock()
	storageSvc = s
}

func currentStorage() contracts.StorageService {
	storageMu.RLock()
	defer storageMu.RUnlock()
	return storageSvc
}

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

// GetActiveUpload loads an active (non-deleted) upload by ID.
func GetActiveUpload(ctx context.Context, id uint64) (models.Upload, error) {
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

// OpenStoredUpload opens the stored object for an active upload via StorageService.
func OpenStoredUpload(ctx context.Context, id uint64) (*OpenedUploadObject, error) {
	upload, err := GetActiveUpload(ctx, id)
	if err != nil {
		return nil, err
	}
	svc := currentStorage()
	if svc == nil {
		return nil, errors.New("storage service not available")
	}
	obj, err := svc.Get(ctx, upload.FilePath)
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
	return waveletupload.RebuildUploadStats(ctx)
}
