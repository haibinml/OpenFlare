package controller

import (
	"github.com/gin-gonic/gin"
	"openflare/service"
)

func GetOrigins(c *gin.Context) {
	origins, err := service.ListOrigins()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, origins)
}

func GetOrigin(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	origin, err := service.GetOriginDetail(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, origin)
}

func CreateOrigin(c *gin.Context) {
	var input service.OriginInput
	if !bindJSON(c, &input) {
		return
	}
	origin, err := service.CreateOrigin(input)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, origin)
}

func UpdateOrigin(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input service.OriginInput
	if !bindJSON(c, &input) {
		return
	}
	origin, err := service.UpdateOrigin(id, input)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, origin)
}

func DeleteOrigin(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := service.DeleteOrigin(id); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, nil)
}
