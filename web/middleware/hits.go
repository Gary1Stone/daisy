package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

// Count the number of requests (hits) the web server recieves within 15-minute intervals
// Global atomic counter for requests in the current 15-minute interval
var hitCounter uint64 = 0

// Start the periodic hit counter logger in a new goroutine
func init() {
	go startHitCounter()
}

// resetHits periodically logs the hit count to the database
// and resets the counter. This is run as a goroutine.
// It adjusts itself to the 1/4 hour increments; 00:00, 00:15, 00:30, 00:45....
func startHitCounter() {
	now := time.Now()
	// Calculate the end time of the current 15-minute interval for the first log.
	firstIntervalEndTime := now.Truncate(15 * time.Minute).Add(15 * time.Minute)
	initialDelay := firstIntervalEndTime.Sub(now)

	// Wait until the end of the current 15-minute interval
	time.Sleep(initialDelay)

	// Perform the first log operation
	resetHits()

	// Start the ticker for subsequent intervals
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		resetHits()
	}
}

// Atomically get the current count and reset it to 0 for the next interval
func resetHits() {
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
