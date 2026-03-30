CREATE INDEX "action_log_idx" ON "action_log" (
	"opened"	ASC,
	"cid"	ASC,
	"action"	ASC,
	"sid"	ASC
) WHERE "cid" IS NOT null;

CREATE INDEX "attacks_idx" ON "attacks" (
	"id"	ASC,
	"timestamp"	ASC,
	"ip"	ASC,
	"city_id"	ASC
);

CREATE INDEX "cities_lat_lon_idx" ON "cities" (
	"latitude"	ASC,
	"longitude"	ASC
);

CREATE UNIQUE INDEX "devices_idx" ON "devices" (
	"cid"	ASC,
	"active"	ASC,
	"name"	ASC
);

CREATE INDEX logins_idx ON logins ("id" ASC, "timestamp" ASC, "uid" ASC, "success" ASC);

CREATE INDEX "profiles_idx" ON "profiles" (
	"user"	ASC
);

CREATE UNIQUE INDEX "software_idx" ON "software" (
	"name"	ASC
);

CREATE INDEX "wlog_idx" ON "wlog" (
	"aid"	ASC,
	"cmd"	ASC
);

CREATE TABLE "action_log" (
	"aid"	INTEGER NOT NULL UNIQUE,
	"action"	TEXT NOT NULL,
	"originator"	INTEGER NOT NULL,
	"opened"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"cid"	INTEGER,
	"cid_ack"	INTEGER NOT NULL DEFAULT 0,
	"sid"	INTEGER,
	"sid_ack"	INTEGER NOT NULL DEFAULT 0,
	"gid"	INTEGER NOT NULL DEFAULT 0,
	"uid"	INTEGER,
	"uid_ack"	INTEGER NOT NULL DEFAULT 0,
	"inform_gid"	INTEGER,
	"inform"	INTEGER,
	"inform_ack"	INTEGER NOT NULL DEFAULT 0,
	"impact"	INTEGER NOT NULL DEFAULT 0,
	"report"	TEXT COLLATE NOCASE,
	"notes"	TEXT NOT NULL,
	"active"	INTEGER NOT NULL DEFAULT 1,
	"closed"	INTEGER NOT NULL DEFAULT 0,
	"closed_by"	INTEGER,
	"wlog"	INTEGER NOT NULL DEFAULT 0,
	"trouble"	INTEGER DEFAULT 0,
	PRIMARY KEY("aid" AUTOINCREMENT),
	FOREIGN KEY("cid") REFERENCES "devices"("cid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("closed_by") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("inform") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("originator") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("sid") REFERENCES "software"("sid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE "alerts" (
	"id"	INTEGER NOT NULL UNIQUE,
	"aid"	INTEGER,
	"uid"	INTEGER,
	"gid"	INTEGER NOT NULL DEFAULT 0,
	"ack"	INTEGER,
	"wait"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("ack") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("aid") REFERENCES "action_log"("aid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE aliases (mac TEXT NOT NULL DEFAULT "" PRIMARY KEY UNIQUE, alias TEXT NOT NULL DEFAULT "", updated INTEGER NOT NULL DEFAULT (- 1), UNIQUE (mac)) STRICT;

CREATE TABLE "api_codes" (
	"id"	INTEGER NOT NULL UNIQUE,
	"timestamp"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"api_code"	TEXT NOT NULL,
	"ip"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE attacks (id INTEGER NOT NULL UNIQUE, timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), ip TEXT NOT NULL, uid INTEGER REFERENCES profiles (uid) ON DELETE SET NULL ON UPDATE CASCADE, method TEXT NOT NULL, path TEXT NOT NULL, browser TEXT NOT NULL DEFAULT "", longitude REAL DEFAULT (0.0) NOT NULL, latitude REAL DEFAULT (0.0) NOT NULL, city_id INTEGER REFERENCES cities (city_id) ON DELETE SET NULL ON UPDATE SET DEFAULT, community_id INTEGER REFERENCES communities (community_id) ON DELETE SET NULL ON UPDATE CASCADE, business_name TEXT DEFAULT "" NOT NULL, business_website TEXT DEFAULT "" NOT NULL, ip_name TEXT DEFAULT "" NOT NULL, ip_type TEXT DEFAULT "" NOT NULL, isp TEXT DEFAULT "" NOT NULL, org TEXT DEFAULT "" NOT NULL, PRIMARY KEY (id AUTOINCREMENT));

CREATE TABLE backups (id INTEGER PRIMARY KEY ASC AUTOINCREMENT, source TEXT NOT NULL DEFAULT "", timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), cid INTEGER REFERENCES devices (cid) ON DELETE SET NULL ON UPDATE CASCADE, volume TEXT NOT NULL DEFAULT "", computer TEXT NOT NULL DEFAULT "", date INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), size INTEGER DEFAULT (0) NOT NULL, method ANY NOT NULL DEFAULT "Full", what TEXT NOT NULL DEFAULT Files) STRICT;

CREATE TABLE cache_dirty (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Ensures only one row exists
    is_dirty BOOLEAN DEFAULT 0
);

CREATE TABLE chains (id INTEGER PRIMARY KEY NOT NULL, mac1 TEXT NOT NULL DEFAULT "", mac2 TEXT NOT NULL DEFAULT "", gap INTEGER NOT NULL DEFAULT (0)) STRICT;

CREATE TABLE "choices" (
	"id"	INTEGER NOT NULL UNIQUE,
	"field"	TEXT NOT NULL,
	"code"	TEXT NOT NULL,
	"description"	TEXT NOT NULL,
	"seq"	INTEGER NOT NULL,
	"active"	INTEGER NOT NULL DEFAULT 1,
	"parent"	TEXT NOT NULL,
	"cnt"	INTEGER NOT NULL DEFAULT 0,
	"asset_id"	TEXT NOT NULL,
	"permissions"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE "cids" (
	"cid"	TEXT NOT NULL UNIQUE,
	PRIMARY KEY("cid")
);

CREATE TABLE cities (city_id INTEGER NOT NULL UNIQUE, city TEXT NOT NULL, city_ascii TEXT NOT NULL, latitude REAL NOT NULL, longitude REAL NOT NULL, country TEXT NOT NULL, country_code TEXT NOT NULL, state TEXT NOT NULL, region_id INTEGER REFERENCES regions (region_id) ON DELETE SET NULL ON UPDATE CASCADE, cos_lat REAL NOT NULL, sin_lat REAL NOT NULL, cos_lon REAL NOT NULL, sin_lon REAL NOT NULL, continent TEXT, continent_code TEXT, PRIMARY KEY (city_id));

CREATE TABLE communities (community_id INTEGER NOT NULL UNIQUE, community TEXT NOT NULL, community_ascii TEXT NOT NULL, country_code TEXT NOT NULL, country TEXT NOT NULL, timezone TEXT NOT NULL, longitude REAL NOT NULL, latitude REAL NOT NULL, cos_lon REAL NOT NULL, sin_lon REAL NOT NULL, cos_lat REAL NOT NULL, sin_lat REAL NOT NULL, city_id INTEGER, region_id INTEGER REFERENCES regions (region_id) ON DELETE SET NULL ON UPDATE CASCADE, PRIMARY KEY (community_id AUTOINCREMENT), FOREIGN KEY (city_id) REFERENCES cities (city_id) ON DELETE SET NULL ON UPDATE CASCADE);

CREATE TABLE "continents" (
	"continent_code"	TEXT NOT NULL COLLATE NOCASE,
	"Continent"	TEXT NOT NULL COLLATE NOCASE,
	PRIMARY KEY("continent_code")
);

CREATE TABLE "countries" (
	"country_code"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"country"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"title"	TEXT COLLATE NOCASE,
	"continent_code"	TEXT NOT NULL COLLATE NOCASE,
	PRIMARY KEY("country_code")
);

CREATE TABLE credentials (credentials_id TEXT NOT NULL UNIQUE, created INTEGER DEFAULT (strftime('%s', 'now')), auth_id TEXT NOT NULL, webAuthnName TEXT, webAuthnDisplayName TEXT, PublicKey TEXT, Transport TEXT, AuthenticatorData TEXT, AttestationType TEXT NOT NULL, ClientDataHash TEXT, ClientDataJSON TEXT, Challenge TEXT, PublicKeyAlgorithm INTEGER, Attachment TEXT, AAGUID TEXT DEFAULT '00000000-0000-0000-0000-000000000000', SignCount INTEGER, CloneWarning INTEGER, BackupEligible INTEGER, BackupState INTEGER, UserPresent INTEGER, UserVerified INTEGER, PRIMARY KEY (credentials_id));

CREATE TABLE "device_filter" (
	"id"	INTEGER NOT NULL UNIQUE,
	"owner"	INTEGER UNIQUE,
	"task"	TEXT NOT NULL,
	"page"	INTEGER NOT NULL,
	"cid"	INTEGER,
	"devtype"	TEXT NOT NULL,
	"site"	TEXT NOT NULL,
	"office"	TEXT NOT NULL,
	"gid"	INTEGER NOT NULL,
	"uid"	INTEGER,
	"searchtxt"	TEXT NOT NULL,
	"islate"	INTEGER NOT NULL DEFAULT 0,
	"ismissing"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("cid") REFERENCES "devices"("cid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("owner") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE "device_mac" (
	"id"	INTEGER NOT NULL UNIQUE,
	"cid"	INTEGER,
	"mac"	TEXT NOT NULL DEFAULT "" COLLATE NOCASE,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("cid") REFERENCES "devices"("cid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE "devices" (
	"cid"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"type"	TEXT NOT NULL DEFAULT 'DESKTOP' COLLATE NOCASE,
	"site"	TEXT NOT NULL,
	"office"	TEXT NOT NULL,
	"location"	TEXT NOT NULL,
	"year"	INTEGER NOT NULL,
	"make"	TEXT NOT NULL,
	"model"	TEXT NOT NULL,
	"cpu"	TEXT NOT NULL,
	"ram"	INTEGER NOT NULL,
	"drivetype"	TEXT NOT NULL,
	"drivesize"	INTEGER NOT NULL,
	"cd"	INTEGER NOT NULL,
	"notes"	TEXT NOT NULL,
	"cores"	INTEGER NOT NULL,
	"gpu"	TEXT NOT NULL,
	"wifi"	INTEGER NOT NULL DEFAULT 0,
	"ethernet"	INTEGER NOT NULL DEFAULT 0,
	"usb"	INTEGER NOT NULL DEFAULT 0,
	"uid"	INTEGER,
	"active"	INTEGER NOT NULL DEFAULT 1,
	"last_updated_by"	INTEGER,
	"last_updated_time"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"image"	TEXT NOT NULL,
	"color"	TEXT NOT NULL,
	"speed"	NUMERIC NOT NULL,
	"status"	NUMERIC NOT NULL DEFAULT 'WORKING',
	"os"	TEXT NOT NULL,
	"serial_number"	TEXT NOT NULL,
	"gid"	INTEGER NOT NULL DEFAULT 0,
	"old_name"	BLOB NOT NULL,
	"os_name"	TEXT,
	"os_version"	TEXT,
	"os_manufacturer"	TEXT,
	"os_configuration"	TEXT,
	"os_build_type"	TEXT,
	"registered_owner"	TEXT,
	"registered_organization"	TEXT,
	"product_id"	TEXT,
	"original_install_date"	TEXT,
	"system_boot_time"	TEXT,
	"system_manufacturer"	TEXT,
	"system_model"	TEXT,
	"system_type"	TEXT,
	"processors"	TEXT,
	"bios_version"	TEXT,
	"windows_directory"	TEXT,
	"system_directory"	TEXT,
	"boot_device"	TEXT,
	"system_locale"	TEXT,
	"input_locale"	TEXT,
	"time_zone"	TEXT,
	"total_physical_memory"	TEXT,
	"available_physical_memory"	TEXT,
	"virtual_memory_max_size"	TEXT,
	"virtual_memory_available"	TEXT,
	"virtual_memory_in_use"	TEXT,
	"page_file_locations"	TEXT,
	"domain"	TEXT,
	"logon_server"	TEXT,
	"hotfixs"	TEXT,
	"network_cards"	TEXT,
	"hyperv_requirements"	TEXT,
	"battery"	INTEGER,
	"last_audit"	INTEGER,
	PRIMARY KEY("cid" AUTOINCREMENT),
	FOREIGN KEY("last_updated_by") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE disks (id INTEGER PRIMARY KEY NOT NULL UNIQUE, cid INTEGER REFERENCES devices (cid) ON DELETE SET NULL ON UPDATE CASCADE MATCH SIMPLE, drive TEXT NOT NULL DEFAULT "", total INTEGER NOT NULL DEFAULT (0), free INTEGER NOT NULL DEFAULT (0), used INTEGER NOT NULL DEFAULT (0), fill REAL NOT NULL DEFAULT (0), timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))) STRICT;

CREATE TABLE "droplists" (
	"fid"	INTEGER NOT NULL UNIQUE,
	"field"	TEXT NOT NULL UNIQUE,
	"label"	TEXT NOT NULL,
	"id"	TEXT NOT NULL,
	"name"	TEXT NOT NULL,
	"title"	TEXT NOT NULL,
	"errmsg"	TEXT NOT NULL,
	"action"	TEXT NOT NULL,
	PRIMARY KEY("fid" AUTOINCREMENT)
);

CREATE TABLE "emails" (
	"id"	INTEGER NOT NULL UNIQUE,
	"timestamp"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"function"	TEXT NOT NULL,
	"uid"	NUMERIC NOT NULL,
	"user"	TEXT NOT NULL,
	"subject"	TEXT NOT NULL,
	"body"	TEXT NOT NULL,
	"template"	TEXT NOT NULL,
	"param1"	TEXT NOT NULL,
	"param2"	TEXT NOT NULL,
	"status"	TEXT NOT NULL,
	"sent"	INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE "hits" (
	"id"	INTEGER NOT NULL UNIQUE,
	"timestamp"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')) UNIQUE,
	"hits"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE "icons" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"description"	TEXT NOT NULL,
	"color"	TEXT NOT NULL,
	"priority"	INTEGER NOT NULL,
	"icon"	TEXT NOT NULL,
	"is_device"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE "locks" (
	"id"	INTEGER NOT NULL UNIQUE,
	"uid"	INTEGER NOT NULL,
	"record_id"	INTEGER NOT NULL,
	"table_name"	INTEGER NOT NULL,
	"display"	INTEGER NOT NULL,
	"timestamp"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("uid") REFERENCES "profiles"("uid")
);

CREATE TABLE logins (id INTEGER NOT NULL UNIQUE, timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), uid INTEGER, tzoff INTEGER NOT NULL DEFAULT 0, longitude REAL NOT NULL, latitude REAL NOT NULL, community_id INTEGER, city_id INTEGER, ip TEXT NOT NULL, success INTEGER NOT NULL DEFAULT (0), session TEXT NOT NULL DEFAULT "", distance INTEGER NOT NULL DEFAULT 0, home TEXT NOT NULL DEFAULT "", country TEXT NOT NULL, state TEXT NOT NULL, city TEXT NOT NULL, community TEXT NOT NULL, timezone TEXT NOT NULL DEFAULT "", PRIMARY KEY (id AUTOINCREMENT), FOREIGN KEY (city_id) REFERENCES cities (city_id) ON DELETE SET NULL ON UPDATE CASCADE, FOREIGN KEY (community_id) REFERENCES communities (community_id) ON DELETE SET NULL ON UPDATE CASCADE, FOREIGN KEY (uid) REFERENCES profiles (uid) ON DELETE SET NULL ON UPDATE CASCADE);

CREATE TABLE mac_correlation (mac1 TEXT NOT NULL DEFAULT "", mac2 TEXT NOT NULL DEFAULT "", corr INTEGER NOT NULL DEFAULT (- 1), site TEXT NOT NULL DEFAULT "", pearson INTEGER NOT NULL DEFAULT (- 1), jaccard INTEGER NOT NULL DEFAULT (- 1), slots INTEGER NOT NULL DEFAULT (0), overlap INTEGER NOT NULL DEFAULT (0), PRIMARY KEY (mac1 ASC, mac2 ASC)) STRICT;

CREATE TABLE macs (mid INTEGER PRIMARY KEY ASC, mac TEXT UNIQUE NOT NULL, created INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), name TEXT NOT NULL DEFAULT "", hostname TEXT NOT NULL DEFAULT "", ip TEXT NOT NULL DEFAULT "", kind TEXT NOT NULL DEFAULT "", os TEXT NOT NULL DEFAULT "", user TEXT NOT NULL DEFAULT "", site TEXT NOT NULL DEFAULT "", office TEXT DEFAULT "" NOT NULL, location TEXT NOT NULL DEFAULT "", note TEXT NOT NULL DEFAULT "", scanned INTEGER NOT NULL DEFAULT (0), vendor TEXT NOT NULL DEFAULT "", online INTEGER NOT NULL DEFAULT (0), source TEXT NOT NULL DEFAULT "", intruder INTEGER NOT NULL DEFAULT (1), updated INTEGER NOT NULL DEFAULT (0), cid INTEGER REFERENCES devices (cid) ON DELETE NO ACTION ON UPDATE CASCADE, active INTEGER NOT NULL DEFAULT (1), isSolitary INTEGER NOT NULL DEFAULT (0), isRandomMac INTEGER NOT NULL DEFAULT (0), isIgnore INTEGER NOT NULL DEFAULT (0)) STRICT;

CREATE TABLE online (mac TEXT NOT NULL, date INTEGER NOT NULL, am INTEGER DEFAULT 0 NOT NULL, pm INTEGER DEFAULT 0 NOT NULL, host INTEGER NOT NULL DEFAULT (0), updated INTEGER NOT NULL DEFAULT (0), PRIMARY KEY (mac, date)) STRICT;

CREATE TABLE "pings" (
	"id"	INTEGER NOT NULL UNIQUE,
	"cid"	INTEGER,
	"utc"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("cid") REFERENCES "devices"("cid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE "profiles" (
	"uid"	INTEGER NOT NULL UNIQUE,
	"user"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"pwd"	TEXT NOT NULL,
	"gid"	INTEGER NOT NULL DEFAULT 3,
	"fullname"	TEXT NOT NULL COLLATE NOCASE,
	"first"	TEXT NOT NULL COLLATE NOCASE,
	"last"	TEXT NOT NULL COLLATE NOCASE,
	"active"	INTEGER NOT NULL DEFAULT 1,
	"last_updated_by"	INTEGER NOT NULL,
	"last_updated_time"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"color"	TEXT NOT NULL,
	"picture"	TEXT NOT NULL,
	"geo_fence"	TEXT NOT NULL,
	"geo_radius"	INTEGER NOT NULL DEFAULT 0,
	"pwd_reset"	INTEGER NOT NULL DEFAULT 0,
	"otp"	TEXT NOT NULL,
	"old_user"	TEXT NOT NULL COLLATE NOCASE,
	"notify"	INTEGER DEFAULT 0,
	"auth_id"	TEXT,
	"cnt"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("uid" AUTOINCREMENT)
);

CREATE TABLE "regions" (
	"region_id"	INTEGER NOT NULL,
	"country_code"	TEXT NOT NULL COLLATE NOCASE,
	"region"	TEXT NOT NULL COLLATE NOCASE,
	PRIMARY KEY("region_id" AUTOINCREMENT)
);

CREATE TABLE "sess" (
	"id"	INTEGER NOT NULL,
	"session"	INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE sessions (session_id TEXT NOT NULL UNIQUE, auth_id TEXT NOT NULL, challenge TEXT NOT NULL, relying_party_id TEXT NOT NULL, expires INTEGER NOT NULL, user_verification TEXT, cred_params TEXT NOT NULL DEFAULT "", PRIMARY KEY (session_id));

CREATE TABLE "sids" (
	"sid"	TEXT NOT NULL UNIQUE,
	PRIMARY KEY("sid")
);

CREATE TABLE "software" (
	"sid"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"licenses"	INTEGER NOT NULL DEFAULT 0,
	"source"	TEXT NOT NULL,
	"license_key"	TEXT NOT NULL,
	"product"	TEXT NOT NULL,
	"link"	TEXT NOT NULL,
	"notes"	TEXT NOT NULL,
	"active"	INTEGER NOT NULL DEFAULT 1,
	"last_updated_by"	INTEGER NOT NULL,
	"last_updated_time"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"color"	TEXT NOT NULL,
	"purchased"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"reuseable"	INTEGER NOT NULL DEFAULT 0,
	"old_name"	TEXT NOT NULL,
	"inv_name"	TEXT,
	"pre_installed"	INTEGER NOT NULL DEFAULT 0,
	"free"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("sid" AUTOINCREMENT),
	FOREIGN KEY("last_updated_by") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE sqlite_sequence(name,seq);

CREATE TABLE sqlite_stat1(tbl,idx,stat);

CREATE TABLE sqlite_stat4(tbl,idx,neq,nlt,ndlt,sample);

CREATE TABLE sw_inv (id INTEGER NOT NULL UNIQUE, cid INTEGER NOT NULL, name TEXT NOT NULL COLLATE NOCASE, scandate INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), sid INTEGER, preinstalled INTEGER DEFAULT (0), PRIMARY KEY (id AUTOINCREMENT), FOREIGN KEY (cid) REFERENCES devices (cid) ON DELETE SET NULL ON UPDATE CASCADE, FOREIGN KEY (sid) REFERENCES software (sid) ON DELETE SET NULL ON UPDATE CASCADE);

CREATE TABLE "tracks" (
	"id"	INTEGER NOT NULL UNIQUE,
	"cid"	INTEGER NOT NULL,
	"timestamp"	BLOB NOT NULL DEFAULT (strftime('%s', 'now')),
	"longitude"	REAL NOT NULL,
	"latitude"	REAL NOT NULL,
	"community_id"	INTEGER,
	"city_id"	INTEGER,
	"ip"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("city_id") REFERENCES "cities"("city_id") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("community_id") REFERENCES "communities"("community_id") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE tracks_wifi (id INTEGER PRIMARY KEY AUTOINCREMENT, timestamp INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL, cid INTEGER REFERENCES devices (cid) ON DELETE SET NULL ON UPDATE CASCADE, ssid TEXT NOT NULL DEFAULT "", bssid TEXT NOT NULL DEFAULT "", rssi INTEGER NOT NULL DEFAULT (0)) STRICT;

CREATE TABLE vendors (MacPrefix TEXT PRIMARY KEY NOT NULL DEFAULT "", Vendor TEXT NOT NULL DEFAULT "", Private INTEGER NOT NULL DEFAULT (0), BlockType TEXT NOT NULL DEFAULT "", LastUpdate TEXT NOT NULL DEFAULT "") STRICT;

CREATE TABLE wifi_locations (id INTEGER PRIMARY KEY AUTOINCREMENT, ssid TEXT NOT NULL DEFAULT "", bssid TEXT NOT NULL DEFAULT "", rssi INTEGER NOT NULL DEFAULT (0), longitude REAL NOT NULL DEFAULT (0.0), latitude REAL NOT NULL DEFAULT (0.0), source TEXT NOT NULL DEFAULT "") STRICT;

CREATE TABLE "wlog" (
	"wid"	INTEGER NOT NULL UNIQUE,
	"timestamp"	INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
	"aid"	INTEGER NOT NULL,
	"uid"	INTEGER,
	"cid"	INTEGER,
	"cmd"	TEXT NOT NULL,
	"notes"	TEXT NOT NULL,
	"olduid"	INTEGER,
	"oldgid"	INTEGER,
	PRIMARY KEY("wid" AUTOINCREMENT),
	FOREIGN KEY("uid") REFERENCES "profiles"("uid") ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY("cid") REFERENCES "devices"("cid") ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TRIGGER mark_cache_dirty_on_count_change
AFTER UPDATE ON choices
WHEN 
(OLD.cnt = 0 AND NEW.cnt > 0) OR  -- 0 → Positive
(OLD.cnt > 0 AND NEW.cnt = 0) OR  -- Positive → 0
(OLD.cnt > 0 AND NEW.cnt < 0)     -- Positive → Negative (unlikely)
BEGIN
UPDATE cache_dirty SET is_dirty = 1 WHERE id = 1;
END;

CREATE TRIGGER update_counts_on_delete
AFTER DELETE ON devices
WHEN OLD.active = 1 -- Only update counts if the deleted row was active
BEGIN
    UPDATE choices SET cnt = cnt - 1 WHERE field = 'TYPE' AND active = 1 AND code = OLD.type;
    UPDATE choices SET cnt = cnt - 1 WHERE field = 'SITE' AND active = 1 AND code = OLD.site;
    UPDATE choices SET cnt = cnt - 1 WHERE field = 'OFFICE' AND active = 1 AND code = OLD.office;
    UPDATE choices SET cnt = cnt - 1 WHERE field = 'GROUP' AND active = 1 AND code = OLD.gid;
    UPDATE profiles SET cnt = cnt - 1 WHERE uid = OLD.uid AND active = 1;
END;

CREATE TRIGGER update_counts_on_insert
AFTER INSERT ON devices
WHEN NEW.active = 1 -- Only update counts if the new row is active
BEGIN
    UPDATE choices SET cnt = cnt + 1 WHERE field = 'TYPE' AND active = 1 AND code = NEW.type;
    UPDATE choices SET cnt = cnt + 1 WHERE field = 'SITE' AND active = 1 AND code = NEW.site;
    UPDATE choices SET cnt = cnt + 1 WHERE field = 'OFFICE' AND active = 1 AND code = NEW.office;
    UPDATE choices SET cnt = cnt + 1 WHERE field = 'GROUP' AND active = 1 AND code = NEW.gid;
    UPDATE profiles SET cnt = cnt + 1 WHERE uid = NEW.uid AND active = 1;
END;

CREATE TRIGGER update_counts_on_update
AFTER UPDATE ON devices
WHEN OLD.active <> NEW.active 
    OR COALESCE(OLD.type, '') <> COALESCE(NEW.type, '') 
    OR COALESCE(OLD.site, '') <> COALESCE(NEW.site, '') 
    OR COALESCE(OLD.office, '') <> COALESCE(NEW.office, '') 
    OR COALESCE(OLD.gid, '') <> COALESCE(NEW.gid, '') 
    OR COALESCE(OLD.uid, '') <> COALESCE(NEW.uid, '')
BEGIN
    -- Decrement counts if active switches from 1 to 0
    UPDATE choices SET cnt = cnt - 1 WHERE field='TYPE' AND active=1 AND code=OLD.type AND OLD.active = 1 AND NEW.active = 0;
    UPDATE choices SET cnt = cnt - 1 WHERE field='SITE' AND active=1 AND code=OLD.site AND OLD.active = 1 AND NEW.active = 0;
    UPDATE choices SET cnt = cnt - 1 WHERE field='OFFICE' AND active=1 AND code=OLD.office AND OLD.active = 1 AND NEW.active = 0;
    UPDATE choices SET cnt = cnt - 1 WHERE field='GROUP' AND active=1 AND code=OLD.gid AND OLD.active = 1 AND NEW.active = 0;
    UPDATE profiles SET cnt = cnt - 1 WHERE uid=OLD.uid AND active=1 AND OLD.active = 1 AND NEW.active = 0;

    -- Increment counts if active switches from 0 to 1
    UPDATE choices SET cnt = cnt + 1 WHERE field='TYPE' AND active=1 AND code=NEW.type AND OLD.active = 0 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='SITE' AND active=1 AND code=NEW.site AND OLD.active = 0 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='OFFICE' AND active=1 AND code=NEW.office AND OLD.active = 0 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='GROUP' AND active=1 AND code=NEW.gid AND OLD.active = 0 AND NEW.active = 1;
    UPDATE profiles SET cnt = cnt + 1 WHERE uid=NEW.uid AND active=1 AND OLD.active = 0 AND NEW.active = 1;

    -- Adjust counts only if active remains 1 and fields change
    UPDATE choices SET cnt = cnt - 1 WHERE field='TYPE' AND active=1 AND code=OLD.type AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='TYPE' AND active=1 AND code=NEW.type AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt - 1 WHERE field='SITE' AND active=1 AND code=OLD.site AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='SITE' AND active=1 AND code=NEW.site AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt - 1 WHERE field='OFFICE' AND active=1 AND code=OLD.office AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='OFFICE' AND active=1 AND code=NEW.office AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt - 1 WHERE field='GROUP' AND active=1 AND code=OLD.gid AND OLD.active = 1 AND NEW.active = 1;
    UPDATE choices SET cnt = cnt + 1 WHERE field='GROUP' AND active=1 AND code=NEW.gid AND OLD.active = 1 AND NEW.active = 1;
    UPDATE profiles SET cnt = cnt - 1 WHERE uid=OLD.uid AND active=1 AND OLD.active = 1 AND NEW.active = 1;
    UPDATE profiles SET cnt = cnt + 1 WHERE uid=NEW.uid AND active=1 AND OLD.active = 1 AND NEW.active = 1;
END;

CREATE VIEW onlinehistory AS SELECT COALESCE(A.alias, O.mac) AS mac, O.date, O.am, O.pm, o.host FROM online O LEFT JOIN aliases A ON O.mac=A.mac;

