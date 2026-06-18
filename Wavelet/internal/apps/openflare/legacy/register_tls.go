// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/tls"
	"github.com/gin-gonic/gin"
)

func registerTLSRoutes(apiGroup *gin.RouterGroup) {
	managedDomainRoute := apiGroup.Group("/managed-domains")
	managedDomainRoute.Use(compat.AdminAuth())
	{
		managedDomainRoute.GET("/", tls.GetManagedDomains)
		managedDomainRoute.GET("/match", tls.MatchManagedDomainCertificateHandler)
		managedDomainRoute.POST("/", tls.CreateManagedDomainHandler)
		managedDomainRoute.POST("/:id/update", tls.UpdateManagedDomainHandler)
		managedDomainRoute.POST("/:id/delete", tls.DeleteManagedDomainHandler)
	}

	tlsCertificateRoute := apiGroup.Group("/tls-certificates")
	tlsCertificateRoute.Use(compat.AdminAuth())
	{
		tlsCertificateRoute.GET("/", tls.GetCertificates)
		tlsCertificateRoute.GET("/:id", tls.GetCertificateDetail)
		tlsCertificateRoute.GET("/:id/content", tls.GetCertificateContentHandler)
		tlsCertificateRoute.POST("/", tls.CreateCertificateHandler)
		tlsCertificateRoute.POST("/:id/update", tls.UpdateCertificateHandler)
		tlsCertificateRoute.POST("/:id/update-acme", tls.UpdateACMECertificateHandler)
		tlsCertificateRoute.POST("/:id/convert-acme", tls.ConvertCertificateToACMEHandler)
		tlsCertificateRoute.POST("/import-file", tls.ImportCertificateFile)
		tlsCertificateRoute.POST("/:id/delete", tls.DeleteCertificateHandler)
		tlsCertificateRoute.POST("/apply", tls.ApplyCertificateHandler)
		tlsCertificateRoute.POST("/:id/renew", tls.RenewCertificateHandler)
	}

	acmeAccountRoute := apiGroup.Group("/acme-accounts")
	acmeAccountRoute.Use(compat.AdminAuth())
	{
		acmeAccountRoute.GET("/default", tls.GetDefaultAcmeAccountHandler)
	}

	dnsAccountRoute := apiGroup.Group("/dns-accounts")
	dnsAccountRoute.Use(compat.AdminAuth())
	{
		dnsAccountRoute.GET("/", tls.GetDNSAccounts)
		dnsAccountRoute.POST("/", tls.CreateDNSAccountHandler)
		dnsAccountRoute.POST("/:id/update", tls.UpdateDNSAccountHandler)
		dnsAccountRoute.POST("/:id/delete", tls.DeleteDNSAccountHandler)
	}
}
