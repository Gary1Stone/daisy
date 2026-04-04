package webserver

import (
	"sync/atomic"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

// Count the number of requests (hits) the web server recieves within 15-minute intervals
// Global atomic counter for requests in the current 15-minute interval
var hitCounter uint64 = 0

// Atomically get the current count and reset it to 0 for the next interval
func ResetHits() {
	count := atomic.SwapUint64(&hitCounter, 0)
	db.LogHits(count)
}

// Ask GoFiber to use this hit counter
// Counts requests: activate after recovery but before protection logic!
// app.Use(middleware.AddHitCounter())
// This Increments the hit counter for each incoming request.
func AddHitCounter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		atomic.AddUint64(&hitCounter, 1)
		return c.Next()
	}
}
