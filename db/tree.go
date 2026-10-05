package db

import (
	"database/sql"
	"log"
)

type Node struct {
	Cid         int    `json:"cid"`         // Computer ID
	Name        string `json:"name"`        // Computer Name (not hostname)
	Icon        string `json:"icon"`        // The icon for the kind (role) of the device
	Parent      int    `json:"parent"`      // The parent device's cid
	Office      string `json:"office"`      // The office the device is in
	OfficeTitle string `json:"officetitle"` // User visible description of the office
	Kind        string `json:"kind"`        // The role (switch, router, desktop...) of the device
	KindTitle   string `json:"kindtitle"`   // User visible description of the kind
	Model       string `json:"model"`       // Device model
	IsOnline    bool   `json:"isonline"`    // Is the device currently online
	Site        string `json:"site"`        // The node's site
}

type TreeNode struct {
	Node
	Children []*TreeNode `json:"children"`
}

const nodeQuery = `SELECT D.cid, D.name, D.model,
COALESCE(D.parent, 0) AS parent,
COALESCE(D.office, '') AS office, 
COALESCE(D.kind, '') AS kind,
COALESCE(I.icon,'') AS icon, 
COALESCE(O.description, '') AS officetitle, 
COALESCE(K.description, '') AS kindtitle,
COALESCE(online, 0) AS online
FROM devices D
LEFT JOIN icons I ON D.kind = I.name
LEFT JOIN choices O ON D.office = O.code AND O.field='OFFICE' AND O.parent=?
LEFT JOIN choices K ON D.type = K.code AND K.field='KIND' 
LEFT JOIN (SELECT cid, MAX(online) AS online FROM macs GROUP BY cid) M2 ON D.cid = M2.cid
WHERE D.active=1 AND D.status != 'STORAGE' `

func GetNode(cid int, site string) (node Node, err error) {
	query := nodeQuery + ` AND D.cid=?`
	err = Conn.QueryRow(query, site, cid).Scan(&node.Cid, &node.Name, &node.Model, &node.Parent, &node.Office, &node.Kind, &node.Icon, &node.OfficeTitle, &node.KindTitle, &node.IsOnline)
	if err != nil {
		log.Println(err)
		return
	}
	return node, nil
}

func GetNodes(site string) (nodes []Node, err error) {
	query := nodeQuery + `AND D.site=? ORDER BY D.name`
	rows, err := Conn.Query(query, site, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var node Node
		err = rows.Scan(&node.Cid, &node.Name, &node.Model, &node.Parent, &node.Office, &node.Kind, &node.Icon, &node.OfficeTitle, &node.KindTitle, &node.IsOnline)
		if err != nil {
			log.Println(err)
			return
		}
		nodes = append(nodes, node)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	return
}

func GetTreeNodes(site string) (roots []*TreeNode, err error) {
	nodes, err := GetNodes(site)
	if err != nil {
		return
	}

	treeNodes := make([]*TreeNode, 0, len(nodes))
	nodeMap := make(map[int]*TreeNode)
	for _, node := range nodes {
		treeNode := &TreeNode{
			Node:     node,
			Children: []*TreeNode{},
		}
		treeNodes = append(treeNodes, treeNode)
		nodeMap[treeNode.Cid] = treeNode
	}

	for _, node := range treeNodes {
		if node.Parent == 0 {
			roots = append(roots, node)
			continue
		}
		if parent, exists := nodeMap[node.Parent]; exists {
			parent.Children = append(parent.Children, node)
		}
	}
	return
}

// Update the Parent, kind, officce fields of all the Device and MAC records
func SetTreeParent(cid, parent int, kind, office string) error {

	tx, err := Conn.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE devices SET parent=?, kind=?, office=? WHERE cid=?", parent, kind, office, cid)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = tx.Exec("UPDATE macs SET parent=?, kind=?, office=? WHERE cid=?", parent, kind, office, cid)
	if err != nil {
		log.Println(err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Fetch the one site with the most elements
func GetDefaultSite() string {
	site := "WKNC"
	cnt := 0
	query := "SELECT COUNT(*) AS cnt, site FROM devices WHERE active=1 GROUP BY site ORDER BY cnt DESC LIMIT 1"
	err := Conn.QueryRow(query).Scan(&cnt, &site)
	if err != nil {
		if err != sql.ErrNoRows {
			return getChoiceSite()
		} else {
			log.Println(err)
			return site
		}
	}
	return site
}

// Return first site in the admin cache
func getChoiceSite() string {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	for _, item := range AdminCache.theSlice {
		if item.Active == 1 && item.Field == "SITE" {
			return item.Code
		}
	}
	return "WKNC"
}
