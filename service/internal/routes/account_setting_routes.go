package routes

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// SetupAccountSettingRoutes proxies the account-setting endpoints to s-erp-auth,
// the single source of truth for user profiles. The route NAMES are identical
// to the ones s-erp-hr-api-go exposes, so every UI calls only its own backend's
// base URL and never s-erp-auth directly. The caller's Bearer token is forwarded
// unchanged (auth derives the user from it); profile images ride the same
// request and auth pushes them to the shared s-erp-storage service.
func SetupAccountSettingRoutes(accountSettings fiber.Router) {
	// In-network base for s-erp-auth (joined docker network). Overridable via env.
	authBase := strings.TrimRight(getenvDefault("AUTH_SERVICE_URL", "http://auth:4020"), "/")
	target := authBase + "/v1/auth/account-setting"

	accountSettings.Post("/show-account-setting", proxyTo(target+"/show-account-setting"))
	accountSettings.Post("/update-account-setting", proxyTo(target+"/update-account-setting"))
}

// proxyTo transparently forwards the current request (method, headers incl.
// Authorization, and body) to the target URL and returns auth's response as-is.
func proxyTo(url string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// c.Request().Header.Set("service_from", "s-erp-api")
		if err := proxy.Do(c, url); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"data":    nil,
				"message": "Auth service unavailable",
				"status":  fiber.StatusBadGateway,
				"errors":  err.Error(),
			})
		}
		// Strip the upstream Server header so responses look like they came from
		// this service.
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	}
}

func getenvDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
