package webserver

import "github.com/gofiber/fiber/v2"

// This function redirects all http traffic to https
// Check if the request is over HTTPS
// Redirect HTTP requests to HTTPS
func SecureOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Protocol() == "https" {
			return c.Next()
		}
		targetURL := "https://" + c.Hostname() + c.OriginalURL()
		return c.Redirect(targetURL, fiber.StatusMovedPermanently)
	}
}
