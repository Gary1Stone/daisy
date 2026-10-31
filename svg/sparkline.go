package svg

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/activeusers"
	"github.com/gbsto/daisy/db"
)

const (
	Day   = 1
	Week  = 7
	Month = 30
)

type SparklineOptions struct {
	Duration   int // 1=Day, 7=Week, 30=Month
	Width      int // Pixel Width of the chart
	Height     int // Pixel Height of the chart
	Warning    int // Warning Threshold for color change, values above this value sets color RED
	points     []int
	MaxValue   int
	color      string
	background string
}

// 1=Day, 7=Week, 30=Month
func BuildAttackChart(opts *SparklineOptions) string {
	var err error
	setChartSize(opts)
	opts.points, err = GetAttacks(opts.Duration, opts.Width)
	if err != nil {
		return "Server Error"
	}
	setOptions(opts)
	return buildChart(opts)
}

func BuildLoginsChart(opts *SparklineOptions) string {
	opts.points = nil
	setChartSize(opts)
	var err error
	var ausr activeusers.ActiveUsers
	opts.points, err = ausr.GetLoginCounts(opts.Duration, opts.Width)
	if err != nil {
		return "Server Error"
	}
	// Don't use auto-ranging the y-axis scale
	opts.MaxValue, err = ausr.GetActiveUserCount()
	if err != nil {
		return "Server Error"
	}
	setOptions(opts)
	return buildChart(opts)
}

func BuildHitsChart(opts *SparklineOptions) string {
	opts.points = nil
	setChartSize(opts)
	var err error
	var ausr activeusers.ActiveUsers
	opts.points, err = ausr.GetHitsCounts(opts.Duration, opts.Width)
	if err != nil {
		return "Server Error"
	}
	setOptions(opts)
	return buildChart(opts)
}

func BuildNetworkLoadChart(opts *SparklineOptions) string {
	opts.points = nil
	setChartSize(opts)
	var err error
	opts.points, err = getNetworkDeviceCountsPerDay(opts.Duration, opts.Width)
	if err != nil {
		return "Server Error"
	}
	setOptions(opts)
	return buildChart(opts)
}

func buildChart(opts *SparklineOptions) string {
	Width := strconv.Itoa(opts.Width)
	Height := strconv.Itoa(opts.Height)
	var chart strings.Builder
	chart.WriteString(`<svg width="`)
	chart.WriteString(Width)
	chart.WriteString(`" height="`)
	chart.WriteString(Height)
	chart.WriteString(`" xmlns="http://www.w3.org/2000/svg">`)
	chart.WriteString(`<rect width="`)
	chart.WriteString(Width)
	chart.WriteString(`" height="`)
	chart.WriteString(Height)
	chart.WriteString(`" x="0" y="0" fill="`)
	chart.WriteString(opts.background)
	chart.WriteString(`" />`)
	chart.WriteString(polyline(opts))
	chart.WriteString(`</svg>`)
	return chart.String()
}

func setChartSize(opts *SparklineOptions) {
	if opts.Width < 100 {
		opts.Width = 320
	}
	if opts.Height < 24 {
		opts.Height = 24
	}
}

func setOptions(opts *SparklineOptions) {
	// Not enough points to draw a polyline
	if len(opts.points) < 2 {
		opts.points = append(opts.points, 0)
		opts.points = append(opts.points, 0)
	}
	// Find the maximum value in the points slice
	for _, value := range opts.points {
		if value > opts.MaxValue {
			opts.MaxValue = value
		}
	}

	// Avoid division by zero if all points are zero
	if opts.MaxValue == 0 {
		opts.MaxValue = 1
	}

	if opts.MaxValue > opts.Warning {
		opts.color = "Maroon"
		opts.background = "LightPink"
	} else {
		opts.color = "Navy"
		opts.background = "LightGreen"
	}
}

func polyline(opts *SparklineOptions) string {
	// Calculate spacing between points
	spacing := float64(opts.Width) / float64(len(opts.points)-1)

	// Build the polyline points
	var line strings.Builder
	line.WriteString(`<polyline points="`)

	for idx, point := range opts.points {
		// Normalize the point's height relative to the maximum value
		scaledHeight := float64(point) / float64(opts.MaxValue) * float64(opts.Height)
		x := int(spacing * float64(idx))
		y := opts.Height - int(scaledHeight)

		// Append the coordinates to the polyline
		line.WriteString(strconv.Itoa(x))
		line.WriteString(",")
		line.WriteString(strconv.Itoa(y))
		if idx < len(opts.points)-1 {
			line.WriteString(" ")
		}
	}

	line.WriteString(`" style="fill:none;stroke:`)
	line.WriteString(opts.color)
	line.WriteString(`;stroke-width:2" />`)

	return line.String()
}

// Returns the counts of how many unique people logged in within 15 minute periods
// for the last 30 days
func getNetworkDeviceCountsPerDay(duration, width int) ([]int, error) {
	var points []int
	query := `WITH RECURSIVE dates AS (
		SELECT DATE('now', '-30 days') AS d
		UNION ALL
		SELECT DATE(d, '+1 day')
		FROM dates
		WHERE d < DATE('now')
		)
		SELECT COALESCE(COUNT(online.date), 0) AS count
		FROM dates
		LEFT JOIN online ON online.date = REPLACE(dates.d, '-', '')
		GROUP BY dates.d
		ORDER BY dates.d DESC
		`
	rows, err := db.Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return points, err
	}
	defer rows.Close()
	for rows.Next() {
		var cnt int
		err := rows.Scan(&cnt)
		if err != nil {
			log.Println(err)
			continue
		} else {
			points = append(points, cnt)
		}
	}
	if rows.Err() != nil {
		log.Println(rows.Err())
		return points, rows.Err()
	}

	return db.BucketDownsample(points, width), nil
}
