package devices

import (
	"errors"

	"github.com/gbsto/daisy/db"
)

type PingInfo struct {
	Hostname string
	ApiKey   string
	UTC      int64
}

// Add a check-in record
func (p *PingInfo) AddRecord() error {
	// look up cid given hostname
	var d Device
	cid, err := d.GetCidByName(p.Hostname)
	if err != nil {
		return err
	}

	// if no record, return error
	if cid == 0 {
		return errors.New("no record")
	}

	// add the check-in
	query := "INSERT INTO pings (cid) VALUES (?)"
	_, err = db.Conn.Exec(query, db.ForeignKey(cid))
	return err
}
