package online

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gbsto/daisy/macs"
)

func GetOnlineDeviceChartData(tzoff int, theDay string) (chartData []Online, macInfo map[string]macs.ChartMacInfo, err error) {
	chartData, err = getOnlineDay(tzoff, theDay)
	if err != nil {
		log.Println(err)
		return // "error getting online details"
	}
	macList := make([]string, 0, len(chartData))
	for _, item := range chartData {
		macList = append(macList, item.Mac)
	}
	if len(macList) == 0 {
		return //"no devices online for date selected"
	}
	macInfo, err = macs.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return //"missing mac info for online devices"
	}
	return //svg.BuildOnlineChart(chartData, macInfo, svg.NameOnly)
}

func GetOnlineDeviceHistoryChartData(tzoff int, mac string) (chartData []Online, macInfo map[string]macs.ChartMacInfo, err error) {
	chartData, err = getMacHistory(tzoff, mac)
	if err != nil {
		log.Println(err)
		return //"error getting online details"
	}
	macList := make([]string, 0, len(chartData))
	macList = append(macList, mac)
	macInfo, err = macs.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return //"missing mac info for online devices"
	}
	return //svg.BuildOnlineChart(chartData, macInfo, svg.DateOnly)
}

func GetAvoidChartData(tzoff int, mac1, mac2 string) (chartData []Online, macInfo map[string]macs.ChartMacInfo, title string, err error) {
	if mac1 == "" || mac2 == "" {
		return //"Please select a device pair..."
	}

	// get the macInfo for each device
	macInfo1, err := macs.GetMacInfoByMac(tzoff, mac1)
	if err != nil {
		log.Println(err)
		return //"error getting mac1 info"
	}
	macInfo2, err := macs.GetMacInfoByMac(tzoff, mac2)
	if err != nil {
		log.Println(err)
		return //"missing mac info for online devices"
	}

	chartData1, err := getMacHistory(tzoff, macInfo1.Mac)
	if err != nil {
		log.Println(err)
		return //"missing mac info for online devices"
	}
	chartData2, err := getMacHistory(tzoff, macInfo2.Mac)
	if err != nil {
		log.Println(err)
		return //"error getting online details"
	}

	macList := make([]string, 0, 2)
	macList = append(macList, macInfo1.Mac, macInfo2.Mac)
	macInfo, err = macs.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return //"missing mac info for online devices"
	}

	// Combine the chartData from the two macs and sort
	chartData = append(chartData1, chartData2...)

	// Sort the combined chartData by date
	for i := range chartData {
		for j := i + 1; j < len(chartData); j++ {
			if chartData[i].Date > chartData[j].Date {
				chartData[i], chartData[j] = chartData[j], chartData[i]
			}
		}
	}

	title = "<p>" + macInfo1.Mac + " (" + macInfo1.Name + ") ↔ " + macInfo2.Mac + " (" + macInfo2.Name + ")</p>"
	return // title + svg.BuildOnlineChart(chartData, macInfo, svg.NameWithDate)
}

func GetLastOnlineDatesTable(curUid, tzoff int) string {
	var tbl strings.Builder
	items, err := GetLastOnlineDates(curUid, tzoff)
	if err != nil {
		log.Println(err)
		return ""
	}
	tbl.WriteString(`<div style="max-height: 400px; overflow-y: auto;"><table class="striped" id="seen"><thead><tr><th>Device</th><th>Last Seen Online</th></thead><tbody>`)
	for _, item := range items {
		fmt.Fprintf(&tbl, `<tr><td><a href='device.html?cid=%d'>%s</a></td><td>%s</td></tr>`, item.Cid, item.Name, item.Date)
	}
	tbl.WriteString(`</tbody></table></div>`)
	return tbl.String()
}

func BuildComputerHistoryChartData(cid, tzoff int) (chartData []Online, macInfo map[string]macs.ChartMacInfo, err error) {
	if cid < 1 || cid == math.MaxInt {
		return
	}
	chartData, err = getComputerHistory(cid, tzoff)
	if err != nil {
		log.Println(err)
		return
	}

	uniqueMacs := make(map[string]bool)
	var maclist []string
	for _, item := range chartData {
		if !uniqueMacs[item.Mac] {
			uniqueMacs[item.Mac] = true
			maclist = append(maclist, item.Mac)
		}
	}

	macInfo, err = macs.GetChartInfo4WhatWasOnline(maclist)
	if err != nil {
		log.Println(err)
		return
	}
	return
}
