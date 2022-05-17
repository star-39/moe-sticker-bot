package main

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/forPelevin/gomoji"
	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
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
	maxL := 60 - len(botName)
	if len(s) < 1 || len(s) > maxL {
		return false
	}
	if _, err := strconv.Atoi(s[:1]); err == nil {
		return false
	}
	if strings.Contains(s, "__") {
		return false
	}

	return true
}

func secHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func secNum(n int) string {
	numbers := ""
	for i := 0; i < n; i++ {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(10))
		numbers += randInt.String()
	}
	return numbers
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

	if host == "t.me" {
		host = LINK_TG
	} else if strings.HasSuffix(host, "line.me") {
		host = LINK_IMPORT
	} else if strings.HasSuffix(host, "e.kakao.com") {
		host = LINK_IMPORT
	}

	log.Debugf("link parsed: link=%s, host=%s", link, host)
	return link, host
}

func findEmojis(s string) string {
	var eString string
	gomojis := gomoji.FindAll(s)
	for _, e := range gomojis {
		eString += e.Character
	}
	return eString
}

func sanitizeCallback(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		log.Debug("Sanitizing callback data...")
		c.Callback().Data = regexAlphanum.FindString(c.Callback().Data)

		log.Debugln("now:", c.Callback().Data)
		return next(c)
	}
}
func autoRespond(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Callback() != nil {
			defer c.Respond()
		}
		return next(c)
	}
}

func escapeTagMark(s string) string {
	s = strings.ReplaceAll(s, "<", "＜")
	s = strings.ReplaceAll(s, ">", "＞")
	return s
}

func getSIDFromMessage(m *tele.Message) string {
	if m.Sticker != nil {
		return m.Sticker.SetName
	}

	link := findLink(m.Text)
	return path.Base(link)
}

func retrieveSSDetails(c tele.Context, id string, sd *StickerData) error {
	ss, err := c.Bot().StickerSet(id)
	if err != nil {
		return err
	}
	sd.title = ss.Title
	sd.id = ss.Name
	sd.cAmount = len(ss.Stickers)
	sd.isVideo = ss.Video
	return nil
}
