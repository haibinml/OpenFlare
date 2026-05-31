package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const invalidParamsMessage = "参数错误"

func respondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

func respondSuccessWithExtras(c *gin.Context, data any, extras gin.H) {
	payload := gin.H{
		"success": true,
		"message": "",
		"data":    data,
	}
	for key, value := range extras {
		payload[key] = value
	}
	c.JSON(http.StatusOK, payload)
}

func respondSuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

func respondFailure(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": message,
	})
}

func respondBadRequest(c *gin.Context, message string) {
	if message == "" {
		message = invalidParamsMessage
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": message,
	})
}

func respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"message": message,
	})
}

func decodeJSONBody(body io.Reader, target any) error {
	return json.NewDecoder(body).Decode(target)
}

func decodeOptionalJSONBody(body io.Reader, target any) error {
	if err := json.NewDecoder(body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		respondBadRequest(c, "")
		return 0, false
	}
	return uint(id), true
}

func bindJSON(c *gin.Context, target any) bool {
	if err := decodeJSONBody(c.Request.Body, target); err != nil {
		respondBadRequest(c, "")
		return false
	}
	return true
}
