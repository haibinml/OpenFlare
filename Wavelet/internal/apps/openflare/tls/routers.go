// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"errors"
	"strings"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleLogicError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		compat.Fail(c, "记录不存在")
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

// GetCertificates 列出 TLS 证书。
func GetCertificates(c *gin.Context) {
	certificates, err := ListCertificates(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificates)
}

// GetCertificateDetail 获取 TLS 证书详情。
func GetCertificateDetail(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	certificate, err := GetCertificate(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// GetCertificateContentHandler 获取 TLS 证书 PEM 内容。
func GetCertificateContentHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	content, err := GetCertificateContent(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, content)
}

// CreateCertificateHandler 从 PEM 创建证书。
func CreateCertificateHandler(c *gin.Context) {
	var input CertificateInput
	if !compat.BindJSON(c, &input) {
		return
	}
	certificate, err := CreateCertificate(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// UpdateCertificateHandler 更新证书。
func UpdateCertificateHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input CertificateInput
	if !compat.BindJSON(c, &input) {
		return
	}
	certificate, err := UpdateCertificate(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// ImportCertificateFile 从文件导入证书。
func ImportCertificateFile(c *gin.Context) {
	name := c.PostForm("name")
	remark := c.PostForm("remark")
	certFile, err := c.FormFile("cert_file")
	if err != nil {
		compat.Fail(c, "缺少证书文件")
		return
	}
	keyFile, err := c.FormFile("key_file")
	if err != nil {
		compat.Fail(c, "缺少私钥文件")
		return
	}
	certificate, err := CreateCertificateFromFiles(c.Request.Context(), name, certFile, keyFile, remark)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// DeleteCertificateHandler 删除证书。
func DeleteCertificateHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteCertificate(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}

// ApplyCertificateHandler 申请 ACME 证书。
func ApplyCertificateHandler(c *gin.Context) {
	var input ApplyInput
	if !compat.BindJSON(c, &input) {
		return
	}
	certificate, err := ApplyCertificate(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// UpdateACMECertificateHandler 更新 ACME 证书配置。
func UpdateACMECertificateHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input ApplyInput
	if !compat.BindJSON(c, &input) {
		return
	}
	certificate, err := UpdateACMECertificate(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// ConvertCertificateToACMEHandler 将上传证书转为 ACME。
func ConvertCertificateToACMEHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input ApplyInput
	if !compat.BindJSON(c, &input) {
		return
	}
	certificate, err := ConvertCertificateToACME(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// RenewCertificateHandler 续期 ACME 证书。
func RenewCertificateHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	certificate, err := RenewCertificate(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, certificate)
}

// GetManagedDomains 列出托管域名。
func GetManagedDomains(c *gin.Context) {
	domains, err := ListManagedDomains(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, domains)
}

// CreateManagedDomainHandler 创建托管域名。
func CreateManagedDomainHandler(c *gin.Context) {
	var input ManagedDomainInput
	if !compat.BindJSON(c, &input) {
		return
	}
	domain, err := CreateManagedDomain(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, domain)
}

// UpdateManagedDomainHandler 更新托管域名。
func UpdateManagedDomainHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input ManagedDomainInput
	if !compat.BindJSON(c, &input) {
		return
	}
	domain, err := UpdateManagedDomain(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, domain)
}

// DeleteManagedDomainHandler 删除托管域名。
func DeleteManagedDomainHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteManagedDomain(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}

// MatchManagedDomainCertificateHandler 匹配域名证书。
func MatchManagedDomainCertificateHandler(c *gin.Context) {
	domain := strings.TrimSpace(c.Query("domain"))
	result, err := MatchManagedDomainCertificate(c.Request.Context(), domain)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, result)
}

// GetDNSAccounts 列出 DNS 账号。
func GetDNSAccounts(c *gin.Context) {
	accounts, err := ListDNSAccounts(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, accounts)
}

// CreateDNSAccountHandler 创建 DNS 账号。
func CreateDNSAccountHandler(c *gin.Context) {
	var input DNSAccountInput
	if !compat.BindJSON(c, &input) {
		return
	}
	account, err := CreateDNSAccount(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, account)
}

// UpdateDNSAccountHandler 更新 DNS 账号。
func UpdateDNSAccountHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input DNSAccountInput
	if !compat.BindJSON(c, &input) {
		return
	}
	account, err := UpdateDNSAccount(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, account)
}

// DeleteDNSAccountHandler 删除 DNS 账号。
func DeleteDNSAccountHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteDNSAccount(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}

// GetDefaultAcmeAccountHandler 获取默认 ACME 账号。
func GetDefaultAcmeAccountHandler(c *gin.Context) {
	account, err := GetDefaultAcmeAccount(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, account)
}
