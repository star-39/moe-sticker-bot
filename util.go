package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/forPelevin/gomoji"
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"mvdan.cc/xurls/v2"
)

var regexAlphanum = regexp.MustCompile(`[a-zA-Z0-9_]+`)

func checkTitle(t string) bool {
	if len(t) > 128 || len(t) == 0 {
		return false
	} else {
		return true
	}
}

func checkID(s string) bool {
	if _, err := strconv.Atoi(s[:1]); err == nil {
		return false
	}
	if len(s) > 63 {
		return false
	}
	return true
}

func secHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func findLink(s string) string {
	rx := xurls.Strict()
	return rx.FindString(s)
}

func findLinkWithType(s string) (string, string) {
	rx := xurls.Strict()
	link := rx.FindString(s)
	if link == "" {
		return "", ""
	}

	u, _ := url.Parse(link)
	host := u.Host

	return link, host
}

func getEmojis(s string) string {
	var eString string
	gomojis := gomoji.FindAll(s)
	for _, e := range gomojis {
		eString += e.Character
	}
	return eString
}

func queryLinksByLineID(s string) []string {
	ids, aes := queryLineS(s)
	if ids == nil || aes == nil {
		return nil
	}
	var links []string
	for index, id := range ids {
		if aes[index] {
			links = append(links, "https://t.me/addstickers/"+id)
		} else {
			// append to top.
			links = append([]string{"https://t.me/addstickers/" + id}, links...)
		}
	}
	return links
}

func sanitizeCallback(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		log.Debug("Sanitizing callback data...")
		log.Debugln("was:", hex.EncodeToString([]byte(c.Callback().Data)))

		c.Callback().Data = regexAlphanum.FindString(c.Callback().Data)

		log.Debugln("now:", hex.EncodeToString([]byte(c.Callback().Data)))
		return next(c)
	}
}

func initWorkspace(b *tele.Bot) {
	botName = b.Me.Username
	dataDir = botName + "_data"

	users = Users{data: make(map[int64]*UserData)}
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.WithFields(log.Fields{"botName": botName, "dataDir": dataDir}).Info("Bot OK.")

	if os.Getenv("USE_DB") == "1" {
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = botName + "_db"
		}
		initDB(dbName)
	} else {
		log.Warn("Not using database because USE_DB is not set to 1.")
	}

	wpConvertWebm, _ = ants.NewPoolWithFunc(8, wConvertWebm)

	if runtime.GOOS == "linux" {
		BSDTAR_BIN = "bsdtar"
	} else {
		BSDTAR_BIN = "tar"
	}
}

func escapeTagMark(s string) string {
	s = strings.ReplaceAll(s, "<", "＜")
	s = strings.ReplaceAll(s, ">", "＞")
	return s
}

func initLogrus() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
	})

	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.TraceLevel)
	}
}
