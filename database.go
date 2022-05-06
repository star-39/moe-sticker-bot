package main

import (
	"database/sql"
	"os"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var DB_VER = 1
var db *sql.DB

func initDB(dbname string) error {
	addr := os.Getenv("DB_ADDR")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	params := make(map[string]string)
	params["autocommit"] = "1"
	dsn := &mysql.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  "tcp",
		Addr:                 addr,
		AllowNativePasswords: true,
		Params:               params,
	}
	db, _ = sql.Open("mysql", dsn.FormatDSN())
	err := db.Ping()
	if err != nil {
		log.Error("Error connecting to mariadb!! DSN:")
		log.Errorln(dsn.FormatDSN())
		db = nil
		return err
	}

	err = verifyDB(dsn, dbname)
	if err != nil {
		db = nil
		return err
	}

	log.Info("MariaDB OK.")
	log.Infoln("DBName: ", dbname)

	return nil
}

func verifyDB(dsn *mysql.Config, dbname string) error {
	_, err := db.Exec("SHOW DATABASES;")
	if err != nil {
		err2 := createMariaDB(dbname)
		if err2 != nil {
			log.Errorln("Error initializing mariadb!! DSN:")
			log.Errorln(dsn.FormatDSN())
			return err2
		}
	}
	db.Close()
	dsn.DBName = dbname
	db, _ = sql.Open("mysql", dsn.FormatDSN())
	log.Debugln("DB DSN:", dsn.FormatDSN())

	db.Exec("CREATE TABLE line (line_id VARCHAR(128), tg_id VARCHAR(128), tg_title VARCHAR(255), line_link VARCHAR(512), auto_emoji BOOL)")
	db.Exec("CREATE TABLE properties (name VARCHAR(128) PRIMARY KEY, value VARCHAR(128))")
	db.Exec("CREATE TABLE stickers (user_id BIGINT, tg_id VARCHAR(128), tg_title VARCHAR(255), timestamp BIGINT)")

	var dbVer string
	errQ := db.QueryRow("SELECT value FROM properties WHERE name=?", "DB_VER").Scan(&dbVer)
	if errQ != nil {
		log.Debugln("error quering db ver:", errQ)
		db.Exec("INSERT properties (name, value) VALUES (?, ?)", "DB_VER", DB_VER)
		log.Infoln("Initialized dbVer to :", DB_VER)
		return nil
	}

	log.Infoln("Queried dbVer is :", dbVer)
	// Do upgrade here.
	return nil
}

func createMariaDB(dbname string) error {
	_, err := db.Exec("CREATE DATABASE " + dbname + " CHARACTER SET utf8mb4")
	if err != nil {
		return err
	}

	return nil
}

func queryLineS(id string) ([]string, []bool) {
	if db == nil {
		return nil, nil
	}
	var tgIDs []string
	var aEs []bool
	var tgID string
	var aE bool
	qs, _ := db.Query("SELECT tg_id, auto_emoji FROM line WHERE line_id=?", id)
	defer qs.Close()
	for qs.Next() {
		err := qs.Scan(&tgID, &aE)
		if err != nil {
			return nil, nil
		}
		tgIDs = append(tgIDs, tgID)
		aEs = append(aEs, aE)
	}
	err := qs.Err()
	if err != nil {
		log.Debugln("No line id match? id: ", id)
		return nil, nil
	}
	log.Debug("Matched line record.")
	log.Debugln("tgIDs:", tgIDs)
	log.Debugln("aEs:", aEs)
	return tgIDs, aEs
}

func insertLineS(lineID string, lineLink string, tgID string, tgLink string, aE bool) {
	if db == nil {
		return
	}
	if lineID == "" || lineLink == "" || tgID == "" || tgLink == "" {
		log.Warn("Empty entry to insert line s")
		return
	}
	_, err := db.Exec("INSERT line (line_id, line_link, tg_id, tg_title, auto_emoji) VALUES (?, ?, ?, ?, ?)",
		lineID, lineLink, tgID, tgLink, aE)

	if err != nil {
		log.Warnln("Failed to insert line s:", lineID, lineLink)
	} else {
		log.Infoln("Insert LineS OK ->", lineID, lineLink, tgID, tgLink, aE)
	}
}

func insertUserS(uid int64, tgID string, tgTitle string, timestamp int64) {
	if db == nil {
		return
	}
	if tgID == "" || tgTitle == "" {
		log.Warn("Empty entry to insert user s")
		return
	}
	_, err := db.Exec("INSERT stickers (user_id, tg_id, tg_title, timestamp) VALUES (?, ?, ?, ?)",
		uid, tgID, tgTitle, timestamp)

	if err != nil {
		log.Warnln("Failed to insert user s:", tgID, tgTitle)
	} else {
		log.Infoln("Insert UserS OK ->", tgID, tgTitle, timestamp)
	}
}

// tgIDs, tgTitles, timestamps
func queryUserS(uid int64) ([]string, []string, []int64) {
	if db == nil {
		return nil, nil, nil
	}
	var tgIDs []string
	var tgTitles []string
	var timestamps []int64
	var tgID string
	var tgTitle string
	var timestamp int64
	qs, _ := db.Query("SELECT tg_id, tg_title, timestamp FROM stickers WHERE user_id=?", uid)
	defer qs.Close()
	for qs.Next() {
		err := qs.Scan(&tgID, &tgTitle, &timestamp)
		if err != nil {
			return nil, nil, nil
		}
		tgIDs = append(tgIDs, tgID)
		tgTitles = append(tgTitles, tgTitle)
		timestamps = append(timestamps, timestamp)
	}
	err := qs.Err()
	if err != nil {
		log.Debugln("No userS match?")
		return nil, nil, nil
	}
	return tgIDs, tgTitles, timestamps
}

func matchUserS(uid int64, id string) bool {
	if db == nil {
		return false
	}
	qs, _ := db.Query("SELECT * FROM stickers WHERE user_id=? AND tg_id=?", uid, id)
	defer qs.Close()
	return qs.Next()
}

func deleteUserS(tgID string) {
	if db == nil {
		return
	}
	db.Exec("DELETE FROM stickers WHERE tg_id=?", tgID)
}
