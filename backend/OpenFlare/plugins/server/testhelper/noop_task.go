// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"context"
	"fmt"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	adminmodel "Wavelet/plugins/domain/admin/model"
	"Wavelet/plugins/infra/database"
)

// NoopTaskService is a contracts.TaskService that records dispatch and ignores the rest.
type NoopTaskService struct {
	LastType    string
	LastPayload []byte
}

var _ contracts.TaskService = (*NoopTaskService)(nil)

func (s *NoopTaskService) Dispatch(ctx context.Context, taskType string, payload []byte, triggeredBy string) (string, error) {
	s.LastType = taskType
	s.LastPayload = payload
	taskID := fmt.Sprintf("test-task-%d", time.Now().UnixNano())
	if conn := database.DB(ctx); conn != nil {
		_ = conn.Create(&adminmodel.TaskExecution{
			ID:          idgen.NextUint64ID(),
			TaskID:      taskID,
			TaskType:    taskType,
			Status:      adminmodel.TaskExecutionStatusPending,
			TriggeredBy: triggeredBy,
			Payload:     string(payload),
		}).Error
	}
	return taskID, nil
}
func (s *NoopTaskService) Retry(context.Context, uint64) (string, error) { return "", nil }
func (s *NoopTaskService) ListTasks() []contracts.TaskMetaDTO            { return nil }
func (s *NoopTaskService) GetTaskMeta(string) (contracts.TaskMetaDTO, bool) {
	return contracts.TaskMetaDTO{}, false
}
func (s *NoopTaskService) ValidatePayload(_ string, payload []byte) ([]byte, error) {
	return payload, nil
}
func (s *NoopTaskService) ReloadScheduler() error                    { return nil }
func (s *NoopTaskService) AppendLog(context.Context, string, ...any) {}
func (s *NoopTaskService) ListExecutions(context.Context, string, string, int, int) ([]contracts.TaskExecutionDTO, int64, error) {
	return nil, 0, nil
}
func (s *NoopTaskService) GetExecution(context.Context, uint64) (*contracts.TaskExecutionDTO, error) {
	return &contracts.TaskExecutionDTO{TaskID: "test-task"}, nil
}
