package ctrls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildAvoidListCtrl() string {
	var ctrl strings.Builder
	items, err := getAvoidData()
	if err != nil {
		log.Println(err)
		return ""
	}

	for _, item := range items {
		if item.Corr < 0 {
			fmt.Fprintf(&ctrl, "<p>%s ↔ %s</p>", item.Hostname1, item.Hostname2)
		} else {
			fmt.Fprintf(&ctrl, "<p>%s = %d%%</p>", item.Hostname1, int(item.Corr/100))
		}
	}
	return ctrl.String()
}

// Build the avoid select pair control, and return the first pair of macs in the list
func BuildAvoidSelectCtrl() (string, string, string) {
	var ctrl strings.Builder
	mac1 := ""
	mac2 := ""
	items, err := getAvoidData()
	if err != nil {
		log.Println(err)
		return "", "", ""
	}
	fmt.Fprintf(&ctrl, "<label for='avoidSelect'><b>Macs With Low Overlap</b></label>")
	fmt.Fprintf(&ctrl, "<select name='avoidSelect' id='avoidSelect' data-role='select' data-filter='false' title='Select a pair>")

	for i, item := range items {
		if i == 0 {
			mac1 = item.Mac1 // returned item
			mac2 = item.Mac2 // returned item
		}
		key := item.Mac1 + "_" + item.Mac2
		if item.Mac1 > item.Mac2 {
			key = item.Mac2 + "_" + item.Mac1
		}
		if item.Corr < 0 {
			fmt.Fprintf(&ctrl, "<option value='%s'>%s ↔ %s</option>", key, item.Hostname1, item.Hostname2)
		} else {
			fmt.Fprintf(&ctrl, "<option value='%s'>%s ↔ %s = %d%%</option>", key, item.Hostname1, item.Hostname2, int(item.Corr/100))
		}
	}
	fmt.Fprintf(&ctrl, "</select>")
	return ctrl.String(), mac1, mac2
}

// Build the avoid select/list pair data
func getAvoidData() ([]db.PairInfo, error) {
	// First find any duplicate hostnames
	macList, err := db.GetDuplicateHostnames()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// If no duplicates, try all macs
	if len(macList) == 0 {
		macList, err = db.GetMacList()
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	// Get correlations only for the macs we're interested in
	items, err := db.GetMacCorrelations(false, macList)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}

func BuildMacCorrelationTable(filter db.MacCorrelationFilter) string {
	var ctrl strings.Builder
	items, err := db.GetMacCorrelation(filter)
	if err != nil {
		log.Println(err)
		return ""
	}

	ctrl.WriteString(`<table data-role="table" id="alerttable" 
    data-rows="-1" data-show-rows-steps="false" 
    data-show-search="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="false" 
    data-show-table-info="false" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead><tr>
	<th>Device1</th><th>Device2</th><th>Jaccard</th><th>Pearsons</th><th>&nbsp;</th>
	</tr></thead><tbody>`)
	for _, item := range items {
		name1 := strings.ReplaceAll(item.Name1, "'", "")
		name2 := strings.ReplaceAll(item.Name2, "'", "")
		host1 := strings.ReplaceAll(item.Hostname1, "'", "")
		host2 := strings.ReplaceAll(item.Hostname2, "'", "")
		jaccard := strconv.FormatFloat(float64(item.Jaccard)/100, 'f', 2, 64)
		if item.Jaccard < 0 {
			jaccard = "-"
		}
		pearson := strconv.FormatFloat(float64(item.Pearson)/100, 'f', 2, 64)
		if item.Pearson == -1 {
			pearson = "-"
		}
		fmt.Fprintf(&ctrl, "<tr><td>%s <br>(%s)</td><td>%s <br>(%s)</td><td>%s</td><td>%s</td>", item.Name1, item.Hostname1, item.Name2, item.Hostname2, jaccard, pearson)
		fmt.Fprintf(&ctrl, `<td><button onclick="showLink('%s', '%s', '%s', '%s', '%s', '%s');">Decide...</button></td></tr>`, item.Mac1, name1, host1, item.Mac2, name2, host2)
	}
	ctrl.WriteString(`</tbody></table>`)
	return ctrl.String()
}
