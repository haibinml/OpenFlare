// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package ofupload

import (
	"context"

	"Wavelet/plugins/domain/upload/cache"
	"Wavelet/plugins/domain/upload/models"
	uploadrepo "Wavelet/plugins/domain/upload/repository"
	uploadstats "Wavelet/plugins/domain/upload/stats"

	"gorm.io/gorm"
)

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
