package ctrls

import (
	"log"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func GetOnlineDevices(tzoff int, theDay string) string {
	chartData, err := db.GetOnlineDay(tzoff, theDay)
	if err != nil {
		log.Println(err)
		return "error getting online details"
	}
	macList := make([]string, 0, len(chartData))
	for _, item := range chartData {
		macList = append(macList, item.Mac)
	}
	if len(macList) == 0 {
		return "no devices online for date selected"
	}
	macInfo, err := db.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return "missing mac info for online devices"
	}
	return svg.BuildOnlineChart(chartData, macInfo, svg.NameOnly)
}

func GetOnlineDeviceHistory(tzoff int, mac string) string {
	chartData, err := db.GetDeviceHistory(tzoff, mac)
	if err != nil {
		log.Println(err)
		return "error getting online details"
	}
	macList := make([]string, 0, len(chartData))
	macList = append(macList, mac)
	macInfo, err := db.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return "missing mac info for online devices"
	}
	return svg.BuildOnlineChart(chartData, macInfo, svg.DateOnly)
}

func GetAvoidChart(tzoff int, mac1, mac2 string) string {
	if mac1 == "" || mac2 == "" {
		return "Please select a device pair..."
	}

	// get the macInfo for each device
	macInfo1, err := db.GetMacInfoByMac(tzoff, mac1)
	if err != nil {
		log.Println(err)
		return "error getting mac1 info"
	}
	macInfo2, err := db.GetMacInfoByMac(tzoff, mac2)
	if err != nil {
		log.Println(err)
		return "missing mac info for online devices"
	}

	chartData1, err := db.GetDeviceHistory(tzoff, macInfo1.Mac)
	if err != nil {
		log.Println(err)
		return "missing mac info for online devices"
	}
	chartData2, err := db.GetDeviceHistory(tzoff, macInfo2.Mac)
	if err != nil {
		log.Println(err)
		return "error getting online details"
	}

	macList := make([]string, 0, 2)
	macList = append(macList, macInfo1.Mac, macInfo2.Mac)
	macInfo, err := db.GetChartInfo4WhatWasOnline(macList)
	if err != nil {
		log.Println("Error in GetChartInfo4WhatWasOnline:", err)
		return "missing mac info for online devices"
	}

	// Combine the chartData from the two macs and sort
	chartData := append(chartData1, chartData2...)

	// Sort the combined chartData by date
	for i := range chartData {
		for j := i + 1; j < len(chartData); j++ {
			if chartData[i].Date > chartData[j].Date {
				chartData[i], chartData[j] = chartData[j], chartData[i]
			}
		}
	}

	title := "<p>" + macInfo1.Mac + " (" + macInfo1.Name + ") ↔ " + macInfo2.Mac + " (" + macInfo2.Name + ")</p>"
	return title + svg.BuildOnlineChart(chartData, macInfo, svg.NameWithDate)
}
