package dashboard

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/icons"
	pluralize "github.com/gertd/go-pluralize"
)

type siteInfo struct {
	site     string
	siteName string
	count    int
	devices  []kindInfo
}

type kindInfo struct {
	kind     string // i.e. DESKTOP, LAPTOP, TABLET, PHONE...
	kindName string // i.e. Desktop
	icon     string //
	count    int    // Number of devices of that kind
	isknown  bool
	isonline bool
}

func BuildKPIs() string {
	var ctrl strings.Builder
	items, err := getKPIs()
	if err != nil {
		log.Println(err)
		return ""
	}

	pluralize := pluralize.NewClient()

	for _, item := range items {
		inInventory := 0
		onlineCount := 0
		for _, item := range item.devices {
			if item.isknown {
				inInventory += item.count
			}
			if item.isonline {
				onlineCount += item.count
			}
		}

		fmt.Fprintf(&ctrl, `<div class="row"><div class="col-12"><article class="card"><header><h3>%s</h3><p>%d Devices (%d <span title='known'>✔️</span>, %d <span title='online'>🟢</span>)</p></header>`,
			item.siteName, item.count, inInventory, onlineCount)

		for _, dto := range item.devices {

			word := dto.kindName
			if dto.count > 1 {
				word = pluralize.Plural(word)
			}
			color := "warning"
			online := "<span title='offline'>🟡</span>"
			if dto.isonline {
				online = "<span title='online'>🟢</span>"
				color = "success"
			}
			known := "<span title='unknown'>❓</span>"
			if dto.isknown {
				known = "<span title='known (in inventory records)'>✔️</span>"
			}

			fmt.Fprintf(&ctrl, `<span class="badge" data-variant="%s">%d %s %s %s %s</span>`, color, dto.count, icons.GetIcon(dto.icon), word, online, known)
		}
		ctrl.WriteString(`</article></div></div>`)
	}

	return ctrl.String()
}

// Get active device counts by site and kind.
func getKPIs() ([]siteInfo, error) {
	query := `SELECT count(*) AS cnt, 1 AS isknown, D.site, COALESCE(C2.description, '') AS sitename,
		D.kind, COALESCE(C.description, I.description, 'unknown') AS kindname, 
		COALESCE(I.icon, '') AS icon, COALESCE(M.online, 0) AS isonline
		FROM devices D
		LEFT JOIN icons I ON I.name=D.kind
		LEFT JOIN choices C ON C.field='KIND' AND C.code=D.kind
		LEFT JOIN choices C2 ON C2.field='SITE' AND C2.code=D.site
		LEFT JOIN macs M ON M.cid=D.cid
		WHERE D.active=1
		GROUP BY D.site, isonline, D.kind

		UNION

		SELECT count(*) AS cnt, M.cid IS NOT NULL AS isknown, M.site, COALESCE(C2.description,'') AS sitename, 
		M.kind, COALESCE(C.description, I.description, 'unknown') AS kindname, 
		COALESCE(I.icon, '') AS icon, M.online AS isonline
		FROM macs M
		LEFT JOIN icons I ON I.name=M.kind
		LEFT JOIN choices C ON C.field='KIND' AND C.code=M.kind
		LEFT JOIN choices C2 ON C2.field='SITE' AND C2.code=M.site
		WHERE M.active=1 AND M.cid IS NULL
		GROUP BY isknown, M.site, M.kind, M.online
`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query dashboard KPIs: %w", err)
	}
	defer rows.Close()
	items := make([]siteInfo, 0)
	siteIndexes := make(map[string]int)
	for rows.Next() {
		var kpi kindInfo
		var site, siteName string
		if err := rows.Scan(&kpi.count, &kpi.isknown, &site, &siteName, &kpi.kind, &kpi.kindName, &kpi.icon, &kpi.isonline); err != nil {
			return nil, fmt.Errorf("scan dashboard KPI: %w", err)
		}

		index, ok := siteIndexes[site]
		if !ok {
			index = len(items)
			siteIndexes[site] = index
			items = append(items, siteInfo{
				site:     site,
				siteName: siteName,
				devices:  make([]kindInfo, 0),
			})
		}
		items[index].count += kpi.count
		items[index].devices = append(items[index].devices, kpi)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard KPIs: %w", err)
	}
	return items, nil
}
