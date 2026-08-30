// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"errors"

	"Wavelet/OpenFlare/plugins/server/model"
	db "Wavelet/plugins/infra/database"
)

// GetActiveAuthSources lists enabled Wavelet auth sources.
func GetActiveAuthSources(ctx context.Context) ([]model.AuthSource, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var sources []model.AuthSource
	if err := conn.Where("is_active = ?", true).Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}

// GetTaskExecutionByTaskID loads a task execution by public task ID.
func GetTaskExecutionByTaskID(ctx context.Context, taskID string) (*model.TaskExecution, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var execution model.TaskExecution
	if err := conn.Where("task_id = ?", taskID).First(&execution).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

// GetSystemUser loads the built-in system user, or returns a synthetic fallback.
func GetSystemUser(ctx context.Context) model.User {
	var user model.User
	conn := db.DB(ctx)
	if conn != nil {
		if err := conn.Where("username = ?", configTypeSystem).First(&user).Error; err == nil {
			return user
		}
	}
	return model.User{
		ID:       999,
		Username: configTypeSystem,
		Nickname: "系统",
	}
}
