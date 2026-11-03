package macs

import "github.com/gbsto/daisy/db"

// The two macs are not same device, set the macs table isSolitary field to true (1)
func SetIsSolitaryDevices(mac1, mac2 string) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE macs SET isSolitary=1 WHERE mac IN (?, ?)", mac1, mac2)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func SetIsIgnoreDevices(mac1, mac2 string) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE macs SET isIgnore=1 WHERE mac IN (?, ?)", mac1, mac2)
	if err != nil {
		return err
	}
	return tx.Commit()
}

/*
About
The second hexadecimal digit of a MAC address changes for randomization. If this digit is a 2, 6, A, or E, it indicates a randomized (locally administered) MAC address, as the operating system sets the "locally administered" bit in the first byte to '1'. The other 46 bits are then randomized.
Locally Administered (LA) bit: This bit is part of the first byte (the first octet) of the MAC address.
Universal vs. Local: A '0' in this bit means the address is universally administered (assigned by the manufacturer), while a '1' means it is locally administered (assigned by the device or network).
How it works: By setting this bit to '1' and randomizing the remaining 46 bits, the device creates a valid, yet non-unique, MAC address for its network connection.
Example: A randomized MAC address might start with 02:XX:XX:XX:XX:XX or A2:XX:XX:XX:XX:XX, whereas a standard, non-randomized address would not start with a 2, 6, A, or E in the second positio

For a standard, non-private MAC address, the last six hexadecimal characters (numbers 0-9 and letters A-F) are changed, while the first six are the manufacturer's identifier. However, iPhones use a feature called "Private Wi-Fi Address" to assign a different, randomized MAC address for each network. This randomization affects all 12 hexadecimal characters to provide enhanced privacy.
Standard MAC address
Structure: A 12-character hexadecimal number (e.g., 00:1A:2B:3C:4D:5E).
Manufacturer Identifier: The first six characters (00:1A:2B) are the Organizationally Unique Identifier (OUI) and are constant for all of Apple's devices.
Device Identifier: The last six characters (3C:4D:5E) are the device-specific part and are unique to that particular piece of hardware.
iPhone "Private Wi-Fi Address" (default)
Dynamic Assignment: The iPhone generates a unique, randomized MAC address for each Wi-Fi network you connect to.
Randomized Characters: This randomization affects all 12 hexadecimal characters in the MAC address.
Persistence: The randomized address is persistent for that network, meaning it stays the same for every future connection to that same network, but changes if you "forget" the network and reconnect.

CONCEPT
1) Pick the oldest mac to be the actual (earliest created timestamp) [ok]
2) Check if maclist table has matching entries, else add one
3) Merge all mac histories into one mac history [ok]
4) Write the new history
4) Check if either mac has USER data filled in within the macs table [ok]
5) If newer mac has data, and older is missing some, merge data into the old one [ok]
6) Delete newer mac from online and from mac table, keeping it in the maclist table [ok]
*/
