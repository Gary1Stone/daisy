package svg

import (
	"sync"
	"time"
)

type GraphCacheStruct struct {
	mu      sync.RWMutex
	Graphs  [9]string
	Max     [9]int
	Updated time.Time
}

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

	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			generateGraphs()
		}
	}()
}

// This takes 3 seconds to run (single thread) and 2 seconds to run (a thread for each individual graph using go wait routines)
func generateGraphs() {

	type graphs struct {
		graph [9]string
		max   [9]int
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
	hits.Warning = 1000 // If more than 1000 hits in a 15 minute interval, highlight red

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

	// Copy the temp variable onto the graphCache global variable
	GraphCache.mu.Lock()
	GraphCache.Graphs = temp.graph
	GraphCache.Max = temp.max
	GraphCache.Updated = time.Now()
	GraphCache.mu.Unlock()
}
