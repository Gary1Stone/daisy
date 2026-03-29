package db

import (
	"errors"
)

type PingInfo struct {
	Hostname string
	ApiKey   string
	UTC      int64
}

// Add a check-in record
func (p *PingInfo) AddRecord() error {
	// look up cid given hostname
	cid, err := GetCidByName(p.Hostname)
	if err != nil {
		return err
	}

	// if no record, return error
	if cid == 0 {
		return errors.New("no record")
	}

	// add the check-in
	query := "INSERT INTO pings (cid) VALUES (?)"
	_, err = Conn.Exec(query, foreignKey(cid))
	return err
}
