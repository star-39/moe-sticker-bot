package core

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/config"
)

var db *sql.DB

func initDB(dbname string) error {
	addr := config.Config.DbAddr
	user := config.Config.DbUser
	pass := config.Config.DbPass
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

	err := verifyDB(dsn, dbname)
	if err != nil {
		return err
	}

	db.Close()
	dsn.DBName = dbname
	db, _ = sql.Open("mysql", dsn.FormatDSN())
	log.Debugln("DB DSN:", dsn.FormatDSN())

	var dbVer string
	err = db.QueryRow("SELECT value FROM properties WHERE name=?", "DB_VER").Scan(&dbVer)
	if err != nil {
		log.Errorln("Error quering dbVer, database corrupt? :", err)
		return err
	}

	log.Infoln("Queried dbVer is :", dbVer)
	log.WithFields(log.Fields{"Addr": addr, "DBName": dbname}).Info("MariaDB OK.")

	return nil
}

func verifyDB(dsn *mysql.Config, dbname string) error {
	err := db.Ping()
	if err != nil {
		log.Errorln("Error connecting to mariadb!! DSN: ", dsn.FormatDSN())
		return err
	}

	_, err = db.Exec("USE " + dbname)
	if err != nil {
		log.Infoln("Can't USE database!", err)
		log.Infof("Database name:%s does not seem to exist, attempting to create.", dbname)
		err2 := createMariadb(dsn, dbname)
		if err2 != nil {
			log.Errorln("Error creating mariadb database!! DSN:", dsn.FormatDSN())
			return err2
		}
	}
	return nil

}

func createMariadb(dsn *mysql.Config, dbname string) error {
	_, err := db.Exec("CREATE DATABASE " + dbname + " CHARACTER SET utf8mb4")
	if err != nil {
		log.Errorln("Error CREATE DATABASE!", err)
		return err
	}
	db.Close()
	dsn.DBName = dbname
	db, _ = sql.Open("mysql", dsn.FormatDSN())
	db.Exec("CREATE TABLE line (line_id VARCHAR(128), tg_id VARCHAR(128), tg_title VARCHAR(255), line_link VARCHAR(512), auto_emoji BOOL)")
	db.Exec("CREATE TABLE properties (name VARCHAR(128) PRIMARY KEY, value VARCHAR(128))")
	db.Exec("CREATE TABLE stickers (user_id BIGINT, tg_id VARCHAR(128), tg_title VARCHAR(255), timestamp BIGINT)")
	db.Exec("INSERT properties (name, value) VALUES (?, ?)", "DB_VER", DB_VER)
	log.Infoln("Mariadb initialized with DB_VER :", DB_VER)
	return nil
}

func insertLineS(lineID string, lineLink string, tgID string, tgTitle string, aE bool) {
	if db == nil {
		return
	}
	if lineID == "" || lineLink == "" || tgID == "" || tgTitle == "" {
		log.Warn("Empty entry to insert line s")
		return
	}
	_, err := db.Exec("INSERT line (line_id, line_link, tg_id, tg_title, auto_emoji) VALUES (?, ?, ?, ?, ?)",
		lineID, lineLink, tgID, tgTitle, aE)

	if err != nil {
		log.Errorln("Failed to insert line s:", lineID, lineLink)
	} else {
		log.Infoln("Insert LineS OK ->", lineID, lineLink, tgID, tgTitle, aE)
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
		log.Errorln("Failed to insert user s:", tgID, tgTitle)
	} else {
		log.Infoln("Insert UserS OK ->", tgID, tgTitle, timestamp)
	}
}

// Pass QUERY_ALL to select all rows.
func queryLineS(id string) []LineStickerQ {
	if db == nil {
		return nil
	}
	var qs *sql.Rows
	var lines []LineStickerQ
	var tgTitle string
	var tgID string
	var aE bool
	if id == "QUERY_ALL" {
		qs, _ = db.Query("SELECT tg_title, tg_id, auto_emoji FROM line")
	} else {
		qs, _ = db.Query("SELECT tg_title, tg_id, auto_emoji FROM line WHERE line_id=?", id)
	}
	defer qs.Close()
	for qs.Next() {
		err := qs.Scan(&tgTitle, &tgID, &aE)
		if err != nil {
			return nil
		}
		lines = append(lines, LineStickerQ{
			tg_id:    tgID,
			tg_title: tgTitle,
			ae:       aE,
		})
		log.Debugf("Matched line record: id:%s | title:%s | ae:%v", tgID, tgTitle, aE)
	}
	err := qs.Err()
	if err != nil {
		log.Errorln("error quering line db: ", id)
		return nil
	}
	return lines
}

// Pass -1 to query all rows.
func queryUserS(uid int64) []UserStickerQ {
	var usq []UserStickerQ
	var q *sql.Rows
	var tgTitle string
	var tgID string
	var timestamp int64

	if uid == -1 {
		q, _ = db.Query("SELECT tg_title, tg_id, timestamp FROM stickers")
	} else {
		q, _ = db.Query("SELECT tg_title, tg_id, timestamp FROM stickers WHERE user_id=?", uid)
	}
	defer q.Close()
	for q.Next() {
		err := q.Scan(&tgTitle, &tgID, &timestamp)
		if err != nil {
			log.Errorln("error scanning user db all", err)
			return nil
		}
		usq = append(usq, UserStickerQ{
			tg_id:     tgID,
			tg_title:  tgTitle,
			timestamp: timestamp,
		})
	}
	err := q.Err()
	if err != nil {
		log.Errorln("error quering all user S")
		return nil
	}
	return usq
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
	_, err := db.Exec("DELETE FROM stickers WHERE tg_id=?", tgID)
	if err != nil {
		log.Errorln("Delete user s err:", err)
	} else {
		log.Infoln("Deleted from database for user sticker:", tgID)
	}
}

func deleteLineS(tgID string) {
	if db == nil {
		return
	}
	_, err := db.Exec("DELETE FROM line WHERE tg_id=?", tgID)
	if err != nil {
		log.Errorln("Delete line s err:", err)
	} else {
		log.Infoln("Deleted from database for line sticker:", tgID)
	}
}

func updateLineSAE(ae bool, tgID string) error {
	if db == nil {
		return nil
	}
	_, err := db.Exec("UPDATE line SET auto_emoji=? WHERE tg_id=?", ae, tgID)
	return err
}
