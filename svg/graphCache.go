package svg

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type GraphCacheStruct struct {
	mu      sync.RWMutex
	Graphs  [10]string
	Max     [10]int
	Updated time.Time
}

// If more networks/sites are being monitored, it would be added here
const (
	AttacksPerDay   = 0 // Number of invalid requests (404) each day on Daisy4wknc
	AttacksPerWeek  = 1 // week
	AttacksPerMonth = 2 // month
	LoginsPerDay    = 3 // Number of user logins each day on Daisy4wknc, a single user can have multiple logins
	LoginsPerWeek   = 4 // week
	LoginsPerMonth  = 5 // month
	HitsPerDay      = 6 // Number of web page requests each day on Daisy4wknc
	HitsPerWeek     = 7 // week
	HitsPerMonth    = 8 // month
	OnlineMonth     = 9 // Sparkline of number of devices connected to the monitored network (WKNC) each day for a month
)

var GraphCache GraphCacheStruct

func (g *GraphCacheStruct) GetGraph(i int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if i < 0 || i >= len(g.Graphs) {
		return "invalid graph index"
	}
	return g.Graphs[i]
}

func (g *GraphCacheStruct) GetMax(i int) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if i < 0 || i >= len(g.Max) {
		return 0
	}
	return g.Max[i]
}

// Note: the db connection has to be initalized before this is run
func StartGraphCache() {
	generateGraphs() // Generate immediately so the first page request has graphs.
	printMemoryUsage()

	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			generateGraphs()
			printMemoryUsage()
		}
	}()
}

// This takes 3 seconds to run (single thread) and 2 seconds to run (a thread for each individual graph using go wait routines)
// Since this is a background operation, we can wait the three seconds, and keep the simpler (less cpu taxing) code
func generateGraphs() {

	type graphs struct {
		graph [10]string
		max   [10]int
	}

	var opts SparklineOptions
	var temp graphs

	// 0 = Attacks per day
	opts.Warning = 10
	opts.Duration = Day
	temp.graph[0] = BuildAttackChart(&opts)
	temp.max[0] = opts.MaxValue

	// 1 = Attacks per week
	opts.Duration = Week
	temp.graph[1] = BuildAttackChart(&opts)
	temp.max[1] = opts.MaxValue

	// 2 = Attacks per month
	opts.Duration = Month
	temp.graph[2] = BuildAttackChart(&opts)
	temp.max[2] = opts.MaxValue

	var login SparklineOptions
	login.Warning = 250 // If more than 250 logins in a 15 minute interval, highlight red

	// 3 = Logins per day
	login.Duration = Day
	temp.graph[3] = BuildLoginsChart(&login)
	temp.max[3] = login.MaxValue

	// 4 = Logins per week
	login.Duration = Week
	temp.graph[4] = BuildLoginsChart(&login)
	temp.max[4] = login.MaxValue

	// 5 = Logins per month
	login.Duration = Month
	temp.graph[5] = BuildLoginsChart(&login)
	temp.max[5] = login.MaxValue

	var hits SparklineOptions
	hits.Warning = 1000 // If more than 1000 hits (web requests) in a 15 minute interval, highlight red

	// 6 = Hits per day
	opts.Duration = Day
	temp.graph[6] = BuildHitsChart(&hits)
	temp.max[6] = hits.MaxValue

	// 7 = Hits per week
	opts.Duration = Week
	temp.graph[7] = BuildHitsChart(&hits)
	temp.max[7] = hits.MaxValue

	// 8 = Hits per month
	hits.Duration = Month
	temp.graph[8] = BuildHitsChart(&hits)
	temp.max[8] = hits.MaxValue

	// 9 = Network Load
	var online SparklineOptions
	online.Warning = 40 // If more than 40 devices in a day, highlight red
	temp.graph[9] = BuildNetworkLoadChart(&online)
	temp.max[9] = online.MaxValue

	// Copy the temp variable onto the graphCache global variable
	GraphCache.mu.Lock()
	GraphCache.Graphs = temp.graph
	GraphCache.Max = temp.max
	GraphCache.Updated = time.Now()
	GraphCache.mu.Unlock()
}

// Alloc: This is the most important one. It’s the actual amount of heap memory your Go objects are currently using.
// Sys: This is the total amount of RAM the operating system has given to your Go program. (Go tends to hold onto memory rather than immediately giving it back to the OS, so Sys is usually higher than Alloc).
// For general understanding, we convert bytes to Megabytes (MB)
func printMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("RAM Used = %v MB\n", m.Alloc/1024/1024)         // RAM currently allocated
	log.Printf("Max  Used = %v MB\n", m.TotalAlloc/1024/1024)   // Total RAM allocated ever (even if freed)
	log.Printf("System RAM = %v MB\n", m.Sys/1024/1024)         // RAM obtained from the system
	log.Printf("Number of Garbage Collections = %v\n", m.NumGC) // Number of garbage collection runs
}
