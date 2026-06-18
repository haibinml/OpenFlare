// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package pages

import (
	"errors"
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleLogicError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		compat.Fail(c, errPagesProjectNotFound)
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

func deploymentIDParam(c *gin.Context) (uint, bool) {
	raw := c.Param("deployment_id")
	if raw == "" {
		compat.Fail(c, "无效的 ID")
		return 0, false
	}
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		compat.Fail(c, "无效的 ID")
		return 0, false
	}
	return uint(id64), true
}

// ListProjectsHandler 列出全部 Pages 项目。
func ListProjectsHandler(c *gin.Context) {
	projects, err := ListProjects(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, projects)
}

// GetProjectHandler 获取 Pages 项目详情。
func GetProjectHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	project, err := GetProject(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, project)
}

// CreateProjectHandler 创建 Pages 项目。
func CreateProjectHandler(c *gin.Context) {
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	project, err := CreateProject(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, project)
}

// UpdateProjectHandler 更新 Pages 项目。
func UpdateProjectHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	project, err := UpdateProject(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, project)
}

// DeleteProjectHandler 删除 Pages 项目。
func DeleteProjectHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteProject(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}

// ListDeploymentsHandler 列出项目的全部部署。
func ListDeploymentsHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	deployments, err := ListProjectDeployments(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, deployments)
}

// UploadDeploymentHandler 上传 Pages 部署包。
func UploadDeploymentHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	file, err := c.FormFile("package")
	if err != nil {
		compat.Fail(c, errPagesPackageMissing)
		return
	}
	deployment, err := UploadDeployment(c.Request.Context(), id, file, "")
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, deployment)
}

// ActivateDeploymentHandler 激活 Pages 部署。
func ActivateDeploymentHandler(c *gin.Context) {
	projectID, ok := compat.IDParam(c)
	if !ok {
		return
	}
	deploymentID, ok := deploymentIDParam(c)
	if !ok {
		return
	}
	project, err := ActivateDeployment(c.Request.Context(), projectID, deploymentID)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, project)
}

// DeleteDeploymentHandler 删除 Pages 部署。
func DeleteDeploymentHandler(c *gin.Context) {
	projectID, ok := compat.IDParam(c)
	if !ok {
		return
	}
	deploymentID, ok := deploymentIDParam(c)
	if !ok {
		return
	}
	if err := DeleteDeployment(c.Request.Context(), projectID, deploymentID); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}

// ListDeploymentFilesHandler 列出部署文件清单。
func ListDeploymentFilesHandler(c *gin.Context) {
	deploymentID, ok := deploymentIDParam(c)
	if !ok {
		return
	}
	files, err := ListDeploymentFiles(c.Request.Context(), deploymentID)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, files)
}
