package controller

import (
	"openflare/service"

	"github.com/gin-gonic/gin"
)

func ListPagesProjects(c *gin.Context) {
	projects, err := service.ListPagesProjects()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, projects)
}

func GetPagesProject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	project, err := service.GetPagesProject(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, project)
}

func CreatePagesProject(c *gin.Context) {
	var input service.PagesProjectInput
	if !bindJSON(c, &input) {
		return
	}
	project, err := service.CreatePagesProject(input)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, project)
}

func UpdatePagesProject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input service.PagesProjectInput
	if !bindJSON(c, &input) {
		return
	}
	project, err := service.UpdatePagesProject(id, input)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, project)
}

func DeletePagesProject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := service.DeletePagesProject(id); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, nil)
}

func ListPagesDeployments(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	deployments, err := service.ListPagesProjectDeployments(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, deployments)
}

func UploadPagesDeployment(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	file, err := c.FormFile("package")
	if err != nil {
		respondBadRequest(c, "缺少 Pages 部署包")
		return
	}
	deployment, err := service.UploadPagesDeployment(
		id,
		file,
		c.PostForm("root_dir"),
		c.PostForm("entry_file"),
		c.GetString("username"),
	)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, deployment)
}

func ActivatePagesDeployment(c *gin.Context) {
	projectID, ok := parseIDParam(c)
	if !ok {
		return
	}
	deploymentID, ok := parseIDParamByName(c, "deployment_id")
	if !ok {
		return
	}
	project, err := service.ActivatePagesDeployment(projectID, deploymentID)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, project)
}

func DeletePagesDeployment(c *gin.Context) {
	projectID, ok := parseIDParam(c)
	if !ok {
		return
	}
	deploymentID, ok := parseIDParamByName(c, "deployment_id")
	if !ok {
		return
	}
	if err := service.DeletePagesDeployment(projectID, deploymentID); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, nil)
}

func ListPagesDeploymentFiles(c *gin.Context) {
	deploymentID, ok := parseIDParamByName(c, "deployment_id")
	if !ok {
		return
	}
	files, err := service.ListPagesDeploymentFiles(deploymentID)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, files)
}

func AgentDownloadPagesDeploymentPackage(c *gin.Context) {
	deploymentID, ok := parseIDParamByName(c, "deployment_id")
	if !ok {
		return
	}
	filePath, fileName, err := service.GetPagesDeploymentPackagePath(deploymentID)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.File(filePath)
}
