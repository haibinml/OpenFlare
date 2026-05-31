package controller

import (
	"github.com/gin-gonic/gin"
	"openflare/model"
)

// GetDefaultAcmeAccount godoc
// @Summary Get default ACME account
// @Tags AcmeAccounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/acme-accounts/default [get]
func GetDefaultAcmeAccount(c *gin.Context) {
	account, err := model.GetDefaultAcmeAccount()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, account)
}
