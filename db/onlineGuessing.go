package db

import (
	"log"
	"strings"
)

func GuessMacInfo() error {
	isWriteNeeded := false
	items, err := getMacInfo("WHERE Intruder=1 AND cid IS NULL", 0) // only want new items from probe
	if err != nil {
		log.Println(err)
		return err
	}

	hostnameList := make([]string, 0)
	for i, item := range items {
		if item.Hostname != "" {
			hostnameList = append(hostnameList, strings.ToUpper(item.Hostname))
		}

		// Name should not be empty or the same as the MAC address if the hostname is available
		if (item.Name == "" || item.Name == item.Mac) && item.Hostname != "" {
			item.Name = item.Hostname
			isWriteNeeded = true
		} else if item.Name == "" {
			item.Name = item.Mac
			isWriteNeeded = true
		}

		// If MAC address is random, assume its a cellphone
		if isRandomMac(item.Mac) {
			item.Kind = "CELLPHONE"
			isWriteNeeded = true

			// If the hostname is "iPhone", we can be more specific about the device type and OS
			if item.Hostname == "iPhone" {
				item.Name = "iPhone"
				item.Os = "iOS"
				item.Vendor = "Apple"
			}

			// If the hostname is "watch", we can be more specific about the device type and OS
			if item.Hostname == "Watch" {
				item.Name = "Apple Watch"
				item.Os = "Watch OS"
				item.Vendor = "Apple"
				item.Kind = "WATCH"
			}
			// If the hostname is "iPad", we can be more specific about the device type and OS
			if item.Hostname == "iPad" {
				item.Name = "iPad"
				item.Os = "iOS"
				item.Vendor = "Apple"
			}
		}
		items[i] = item
	}
	if len(hostnameList) > 0 {
		devs, err := GetDevicesByNames(hostnameList)
		if err != nil {
			log.Println(err)
			return err
		}

		for i, item := range items {
			if dev, ok := devs[strings.ToUpper(item.Hostname)]; ok {
				item.Cid = dev.Cid
				item.Name = dev.Name
				item.Os = dev.Os
				item.Kind = dev.Type
				item.User = dev.Assigned
				item.Site = dev.Site
				item.Office = dev.Office
				item.Location = dev.Location
				isWriteNeeded = true
			}
			items[i] = item
		}
	}

	if isWriteNeeded {
		err = SaveMacs(items)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// The second hexadecimal digit of a MAC address changes for randomization.
// If this digit is a 2, 6, A, or E, it indicates a randomized
// If the second hexadecimal digit is one of these values, we can assume it's a mobile device.
func isRandomMac(macAddress string) bool {
	if len(macAddress) < 2 {
		return false
	}
	switch macAddress[1] {
	case '2', '6', 'A', 'E', 'a', 'e':
		return true
	}
	return false
}
