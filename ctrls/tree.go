package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

/* Build a html tree view of the network assets from the DEVICES table */
func BuildTreeView(site string) string {
	nodes, err := db.GetTreeNodes(site)
	if err != nil {
		return "Server Error: Unable to build tree view."
	}

	// Generate HTML for the tree view
	var html strings.Builder
	fmt.Fprintf(&html, `<ul><li><a href='#' data-target="siteSelect" onclick="toggleModal(event)"> %s Internet %s</a></li><ul>`, svg.GetIcon("internet"), htmlEscape(site))

	for _, rootNode := range nodes {
		html.WriteString(generateNodeHTML(rootNode))
	}
	html.WriteString("</ul></ul>")
	return html.String()
}

func generateNodeHTML(node *db.TreeNode) string {
	var html strings.Builder
	fmt.Fprintf(&html, "<li data-id='%d'> %s <a href='#' onclick='showDetail(%d)'>%s</a> %s (%s)",
		node.Cid, setIconColor(node.Icon, node.IsOnline), node.Cid, htmlEscape(node.Name), htmlEscape(node.Model), htmlEscape(node.OfficeTitle))

	if len(node.Children) > 0 {
		html.WriteString("<ul>")
		for _, child := range node.Children {
			html.WriteString(generateNodeHTML(child))
		}
		html.WriteString("</ul>")
	}
	html.WriteString("</li>")
	return html.String()
}

// Simple helper to prevent raw strings from breaking HTML tags or introducing injection vulnerabilities
func htmlEscape(s string) string {
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;", "\"", "&quot;")
	return r.Replace(s)
}

func BuildParentSelect(selected int, site string, readOnly bool) string {
	var ctrl strings.Builder
	items, err := db.GetNodes(site)
	if err != nil {
		log.Println(err)
		return ""
	}
	disabled := ""
	if readOnly {
		disabled = "disabled"
	}

	fmt.Fprintf(&ctrl, `<select id="parent" name="parent" data-tooltip="Select Parent Device" %s aria-invalid="false" aria-describedby="parentErr" ><option value=""></option>`, disabled)
	for _, item := range items {
		txt := ""
		if selected == item.Cid {
			txt = "selected"
		}
		fmt.Fprintf(&ctrl, `<option value='%d' %s>%s %s (%s)</option>`, item.Cid, txt, item.Name, item.Model, item.Office)
	}
	ctrl.WriteString("</select>")
	return ctrl.String()
}

func setIconColor(iconName string, online bool) string {
	class := "fg-red"
	tip := "offline"
	if online {
		class = "fg-green"
		tip = "online"
	}
	return fmt.Sprintf(`<span class="%s" data-tooltip="%s">%s</span>`, class, tip, svg.GetIcon(iconName))
}
