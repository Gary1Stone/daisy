package svg

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildOnlineChart(items []db.Online, macInfo map[string]db.ChartMacInfo, barTitle string) string {
	if len(items) == 0 {
		return ""
	}
	var chart strings.Builder
	chartHeight := len(items)*barHeight + headerHeight
	writeSVGHeader(&chart, chartHeight)
	writeAxes(&chart)
	writeTimeLabels(&chart)
	writeRows(&chart, items, macInfo, barTitle)
	writeGridLines(&chart, chartHeight)
	writeBorder(&chart, chartHeight)
	return chart.String()
}

func writeSVGHeader(chart *strings.Builder, height int) {
	fmt.Fprintf(chart, `<svg width="100%%" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" >`, height, chartWidth, height)
	chart.WriteString(svgStyle)
}

func writeAxes(chart *strings.Builder) {
	fmt.Fprintf(chart, `<line x1="0" y1="40" x2="%d" y2="40" stroke="black"/>`, chartWidth) // horizontal axis
}

func writeTimeLabels(chart *strings.Builder) {
	times := []string{"01:00", "03:00", "06:00", "09:00", "12:00", "15:00", "18:00", "21:00", "23:00"}
	positions := []int{60, 180, 360, 540, 720, 900, 1080, 1260, 1390}
	for i, t := range times {
		fmt.Fprintf(chart, `<text class="titleText" x="%d" y="%d" >%s</text>`, positions[i], timeLabelY, t)
	}
}

func writeRows(chart *strings.Builder, items []db.Online, macInfo map[string]db.ChartMacInfo, barTitle string) {
	for i, item := range items {
		mid, name := lookupMac(item.Mac, macInfo)
		y := (i * barHeight) + headerHeight

		// Open Group for the row
		fmt.Fprintf(chart, `<g id="Svg%d" onclick="svgClicked('%d')" style="cursor:pointer;">`, mid, mid)

		// Run-length encoding for slots to reduce SVG size
		if len(item.Slots) > 0 {
			currentVal := item.Slots[0]
			currentStart := 0
			for j := 1; j < len(item.Slots); j++ {
				if item.Slots[j] != currentVal {
					drawSegment(chart, currentStart, j, currentVal, y)
					currentVal = item.Slots[j]
					currentStart = j
				}
			}
			drawSegment(chart, currentStart, len(item.Slots), currentVal, y)
		}

		// Draw text label
		middle := barHeight/2 + y + 5
		switch barTitle {
		case NameOnly:
			fmt.Fprintf(chart, `<text class="whiteText" x="10" y="%d" pointer-events="none">%s</text>`, middle, name)
		case NameWithDate:
			class := "blackText"
			if item.Mac == items[0].Mac {
				class = "whiteText"
			}
			fmt.Fprintf(chart, `<text class="%s" x="10" y="%d" pointer-events="none">%s %s</text>`, class, middle, name, formatDate(item.Date))
		case DateOnly:
			fmt.Fprintf(chart, `<text class="whiteText" x="10" y="%d" pointer-events="none">%s</text>`, middle, formatDate(item.Date))
		}
		chart.WriteString(`</g>`) // Close Group
	}
}

func writeGridLines(chart *strings.Builder, height int) {
	for t := 60; t < 1390; t += 60 {
		fmt.Fprintf(chart, `<line x1="%d" y1="40" x2="%d" y2="%d" stroke="grey"/>`, t, t, height)
	}
	// major tic mark stubs
	chart.WriteString(`<line x1="60" y1="30" x2="60" y2="40" stroke="grey"/>`)
	for t := 180; t < 1390; t += 180 {
		fmt.Fprintf(chart, `<line x1="%d" y1="30" x2="%d" y2="40" stroke="grey"/>`, t, t)
	}
}

func writeBorder(chart *strings.Builder, height int) {
	fmt.Fprintf(chart, `<rect width="%d" height="%d" x="0" y="0" fill="none" stroke-width="2" stroke="black" /></svg>`, chartWidth, height)
}

// Variable width segment
func drawSegment(chart *strings.Builder, start, end, val, y int) {
	width := (end - start) * slotPixelWidth
	x := start * slotPixelWidth
	colour := "black"
	status := "Online"
	switch val {
	case offline:
		colour = offlineColor
		status = "Offline"
	case online:
		colour = onlineColor
		status = "Online"
	case outage:
		colour = outageColor
		status = "Outage"
	}
	fmt.Fprintf(chart, `<rect width="%d" height="30" x="%d" y="%d" fill="%s" stroke-width="0" stroke="%s"><title>%s %s - %s</title></rect>`, width, x, y, colour, colour, status, slotToTime(start), slotToTime(end))
}

// Convert slot index (0–96) to "HH:MM"
func slotToTime(slot int) string {
	totalMinutes := slot * slotPixelWidth
	h := totalMinutes / 60
	m := totalMinutes % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// Convert YYYYMMDD to YYYY-MM-DD
func formatDate(date int64) string {
	year := date / 10000
	month := (date % 10000) / 100
	day := date % 100
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func lookupMac(mac string, macInfo map[string]db.ChartMacInfo) (mid int, name string) {
	if macInfo == nil {
		return 0, ""
	}
	if info, ok := macInfo[mac]; ok {
		return info.Mid, info.Name
	}
	return 0, ""
}
