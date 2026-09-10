package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SetupAuth2FAProxyRoutes proxies the 2FA verification endpoints to s-erp-auth.
// The UI calls these on the main API base URL; they are forwarded unchanged to
// the auth service. No Bearer token is needed (login/verify is a pre-auth
// endpoint), but cookies are forwarded so the auth service can set the refresh
// cookie on successful verification.
func SetupAuth2FAProxyRoutes(auth fiber.Router) {
	authBase := strings.TrimRight(getenvDefault("AUTH_SERVICE_URL", "http://auth:4020"), "/")
	target := authBase + "/v1/auth"

	auth.Post("/login/verify", proxyTo(target+"/login/verify"))
	auth.Post("/login/resend", proxyTo(target+"/login/resend"))
	// TOTP login verification (pre-auth).
	auth.Post("/mfa/totp/verify", proxyTo(target+"/mfa/totp/verify"))
	// Fallback: switch a TOTP login challenge to email OTP (pre-auth).
	auth.Post("/mfa/login/use-email", proxyTo(target+"/mfa/login/use-email"))
	// Alternate: switch an email login challenge to TOTP (pre-auth).
	auth.Post("/mfa/login/use-totp", proxyTo(target+"/mfa/login/use-totp"))
}

// SetupAuth2FAManagementProxyRoutes proxies the per-account 2FA management
// endpoints to s-erp-auth. These require a valid Bearer token (the user must
// be logged in to view/change their 2FA setting).
func SetupAuth2FAManagementProxyRoutes(auth fiber.Router) {
	authBase := strings.TrimRight(getenvDefault("AUTH_SERVICE_URL", "http://auth:4020"), "/")
	target := authBase + "/v1/auth"

	auth.Get("/2fa/status", proxyTo(target+"/2fa/status"))
	auth.Post("/2fa/toggle", proxyTo(target+"/2fa/toggle"))
	auth.Post("/2fa/toggle/verify", proxyTo(target+"/2fa/toggle/verify"))

	// MFA / TOTP management (authenticated).
	auth.Get("/mfa/status", proxyTo(target+"/mfa/status"))
	auth.Post("/mfa/totp/setup", proxyTo(target+"/mfa/totp/setup"))
	auth.Post("/mfa/totp/confirm", proxyTo(target+"/mfa/totp/confirm"))
	auth.Post("/mfa/totp/disable", proxyTo(target+"/mfa/totp/disable"))
	auth.Post("/mfa/primary", proxyTo(target+"/mfa/primary"))
	auth.Post("/mfa/recovery/regenerate", proxyTo(target+"/mfa/recovery/regenerate"))
}
