package core

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*

DATABASE VERSION 2 SCHEMA
MariaDB > show tables;
+----------------------------------+
| Tables_in_BOT_NAME_db |
+----------------------------------+
| line                             |
| properties                       |
| stickers                         |
+----------------------------------+

MariaDB > desc line;
+------------+--------------+------+-----+---------+-------+
| Field      | Type         | Null | Key | Default | Extra |
+------------+--------------+------+-----+---------+-------+
| line_id    | varchar(128) | YES  |     | NULL    |       |
| tg_id      | varchar(128) | YES  |     | NULL    |       |
| tg_title   | varchar(255) | YES  |     | NULL    |       |
| line_link  | varchar(512) | YES  |     | NULL    |       |
| auto_emoji | tinyint(1)   | YES  |     | NULL    |       |
+------------+--------------+------+-----+---------+-------+

MariaDB > desc stickers;
+-----------+--------------+------+-----+---------+-------+
| Field     | Type         | Null | Key | Default | Extra |
+-----------+--------------+------+-----+---------+-------+
| user_id   | bigint(20)   | YES  |     | NULL    |       |
| tg_id     | varchar(128) | YES  |     | NULL    |       |
| tg_title  | varchar(255) | YES  |     | NULL    |       |
| timestamp | bigint(20)   | YES  |     | NULL    |       |
+-----------+--------------+------+-----+---------+-------+

MariaDB > desc properties;
+-------+--------------+------+-----+---------+-------+
| Field | Type         | Null | Key | Default | Extra |
+-------+--------------+------+-----+---------+-------+
| name  | varchar(128) | NO   | PRI | NULL    |       |
| value | varchar(128) | YES  |     | NULL    |       |
+-------+--------------+------+-----+---------+-------+

Current entries for properties:
name: DB_VER
value: 2
name: last_line_dedup_index
value: -1

*/

var db *sql.DB

const DB_VER = "2"

func initDB(dbname string) error {
	addr := Config.DbAddr
	user := Config.DbUser
	pass := Config.DbPass
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
	checkUpgradeDatabase(dbVer)

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

func checkUpgradeDatabase(queriedDbVer string) {
	if queriedDbVer == "1" {
		db.Exec("INSERT properties (name, value) VALUES (?, ?)", "last_line_dedup_index", "-1") //value is string!
		db.Exec("UPDATE properties SET value=? WHERE name=?", "2", "DB_VER")
		log.Info("Upgraded DB_VER from 1 to 2")
	}
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
	db.Exec("INSERT properties (name, value) VALUES (?, ?)", "last_line_dedup_index", "-1")
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
			Tg_id:    tgID,
			Tg_title: tgTitle,
			Ae:       aE,
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

func searchLineS(keywords []string) []LineStickerQ {
	if db == nil {
		return nil
	}
	var statements []string
	for _, s := range keywords {
		statements = append(statements, "'%"+s+"%'")
	}
	statement := strings.Join(statements, " AND tg_title LIKE ")
	log.Debugln("database: search statement:", statement)
	qs, err := db.Query("SELECT tg_title, tg_id, auto_emoji FROM line WHERE tg_title LIKE " + statement)
	if err != nil {
		log.Warnln("db q err:", err)
		return nil
	}

	var lines []LineStickerQ
	var tgTitle string
	var tgID string
	var aE bool
	defer qs.Close()
	for qs.Next() {
		err := qs.Scan(&tgTitle, &tgID, &aE)
		if err != nil {
			return nil
		}
		lines = append(lines, LineStickerQ{
			Tg_id:    tgID,
			Tg_title: tgTitle,
			Ae:       aE,
		})
		log.Debugf("Search matched line record: id:%s | title:%s | ae:%v", tgID, tgTitle, aE)
	}
	err = qs.Err()
	if err != nil {
		log.Errorln("error searching line db: ", keywords)
		return nil
	}
	return lines
}

func curateDatabase() error {
	lastIndexOfDedupOK := getLastLineDedupIndex()
	//Line stickers.
	ls := queryLineS("QUERY_ALL")
	for i, l := range ls {
		// if i < startIndex {
		// 	continue
		// }
		log.Debugf("Scanning:%s", l.Tg_id)
		ss, err := b.StickerSet(l.Tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Infof("SS:%s is invalid. purging it from db...", l.Tg_id)
				deleteLineS(l.Tg_id)
				deleteUserS(l.Tg_id)
			} else {
				log.Errorln(err)
			}
			continue
		}

		for si := range ss.Stickers {
			if si > 0 {
				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
					log.Warnln("Setting auto emoji to FALSE for ", l.Tg_id)
					updateLineSAE(false, l.Tg_id)
				}
			}
		}

		if i > lastIndexOfDedupOK {
			workdir := filepath.Join(dataDir, secHex(8))
			os.MkdirAll(workdir, 0755)
			for si, s := range ss.Stickers {
				if si > 0 {

					fp := filepath.Join(workdir, strconv.Itoa(si-1)+".webp")
					f := filepath.Join(workdir, strconv.Itoa(si)+".webp")
					b.Download(&s.File, f)

					if compCRC32(f, fp) {
						b.DeleteSticker(s.FileID)
					}
				}
			}
			os.RemoveAll(workdir)
			lastIndexOfDedupOK = i
		}
	}
	setLastLineDedupIndex(lastIndexOfDedupOK)

	//User stickers.
	us := queryUserS(-1)
	for _, u := range us {
		log.Infof("Checking:%s", u.tg_id)
		_, err := b.StickerSet(u.tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Warnf("SS:%s is invalid. purging it from db...", u.tg_id)
				deleteUserS(u.tg_id)
			} else {
				log.Errorln(err)
			}
		}
	}

	return nil
}

func setLastLineDedupIndex(index int) {
	value := strconv.Itoa(index)
	db.Exec("UPDATE properties SET value=? WHERE name=?", value, "last_line_dedup_index")
	log.Infoln("setLastLineDedupIndex to :", value)
}

func getLastLineDedupIndex() int {
	var value string
	db.QueryRow("SELECT value FROM properties WHERE name=?", "last_line_dedup_index").Scan(&value)
	index, _ := strconv.Atoi(value)
	log.Infoln("getLastLineDedupIndex", value)
	return index
}
