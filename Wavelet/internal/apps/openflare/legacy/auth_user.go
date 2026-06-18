// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	ofauth "github.com/Rain-kl/Wavelet/internal/apps/openflare/auth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Nickname    string `json:"nickname"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Code        string `json:"code"`
}

type updateSelfRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	user, err := ofauth.Login(c.Request.Context(), c, ofauth.LoginInput{
		Username: req.Username,
		Password: req.Password,
		Code:     req.Code,
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, user)
}

func Logout(c *gin.Context) {
	if err := ofauth.Logout(c.Request.Context(), c); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func GetSelf(c *gin.Context) {
	user, err := ofauth.GetSelf(c.Request.Context(), callerUserID(c))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, user)
}

func UpdateSelf(c *gin.Context) {
	var req updateSelfRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	if err := ofauth.UpdateSelf(c.Request.Context(), callerUserID(c), ofauth.UpdateSelfInput{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	}); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func DeleteSelf(c *gin.Context) {
	if err := ofauth.DeleteSelf(c.Request.Context(), callerUserID(c)); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func GenerateToken(c *gin.Context) {
	token, err := ofauth.GenerateUserToken(c.Request.Context(), callerUserID(c))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, token)
}

func Register(c *gin.Context) {
	var req registerRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	user, err := ofauth.Register(c.Request.Context(), c, ofauth.RegisterInput{
		Username:    req.Username,
		Password:    req.Password,
		Nickname:    req.Nickname,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Code:        req.Code,
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, user)
}

func SendEmailVerification(c *gin.Context) {
	email := c.Query("email")
	if err := ofauth.SendRegisterVerificationEmail(c.Request.Context(), email); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}

func ResetPassword(c *gin.Context) {
	var req passwordResetRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	password, err := ofauth.ResetPassword(c.Request.Context(), ofauth.PasswordResetInput{
		Email: req.Email,
		Token: req.Token,
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, password)
}
