package tree

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/droplists"
	"github.com/gbsto/daisy/icons"
)

/* Build a html tree view of the network assets from the DEVICES table */
func BuildTreeView(site string) string {
	nodes, err := GetTreeNodes(site)
	if err != nil {
		return "Server Error: Unable to build tree view."
	}

	// Generate HTML for the tree view
	var html strings.Builder
	fmt.Fprintf(&html, `<ul><li><a href='#' data-target="siteSelect" onclick="toggleModal(event);"> %s Internet %s</a></li><ul>`, icons.GetIcon("internet"), htmlEscape(site))

	for _, rootNode := range nodes {
		html.WriteString(generateNodeHTML(rootNode))
	}
	html.WriteString("</ul></ul>")
	return html.String()
}

func generateNodeHTML(node *TreeNode) string {
	var html strings.Builder
	color := colors.Alert
	tip := "offline"
	if node.IsOnline {
		color = colors.Success
		tip = "online"
	} else if node.NoMacs {
		color = colors.Info
		tip = "never online or offline"
	}
	fmt.Fprintf(&html, "<li data-id='%d'> %s <a href='#' onclick='showDetail(%d)'>%s</a> %s (%s)",
		node.Cid, droplists.SetIconColor(node.Icon, color, tip), node.Cid, htmlEscape(node.Name), htmlEscape(node.Model), htmlEscape(node.OfficeTitle))

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
	items, err := GetNodes(site)
	if err != nil {
		log.Println(err)
		return ""
	}

	var droplist droplists.Droplist
	droplist.Label = "Parent"
	droplist.Id = "parent"
	droplist.Name = "parent"
	droplist.Title = "Select Parent Device"
	droplist.ReadOnly = false
	droplist.ErrMsg = "A parent can be selected"
	droplist.Action = ""

	var options []droplists.DroplistOption
	for _, item := range items {
		var opt droplists.DroplistOption
		opt.Value = fmt.Sprintf(`%d`, item.Cid)
		opt.Description = fmt.Sprintf(`%s %s (%s)`, item.Name, item.Model, item.OfficeTitle)
		opt.Icon = item.Icon
		opt.Selected = false
		opt.Colour = "fg-red"
		opt.Tip = "offline"
		if item.IsOnline {
			opt.Colour = "fg-green"
			opt.Tip = "online"
		}
		options = append(options, opt)
	}

	return droplists.BuildDropdown(droplist, options, readOnly, "", "")
}
