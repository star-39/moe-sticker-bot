package core

import (
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

// DANGER ZONE!
// DO NOT USE UNLESS SCRUTINIZED THE CODES.

// This command is to sanitize duplicated sticker in a set, or update its auto_emoji status.
// You should not use this command unless you were using the python version before.
// It takes forever to run for HUGE databases.
func cmdSanitize(c tele.Context) error {
	if ADMIN_UID != c.Sender().ID {
		return c.Send("Admin only command. /start")
	}

	msgText := c.Message().Text
	args := strings.Split(msgText, " ")
	if len(args) <= 1 {
		return c.Send("Missing subcommand! invalid / dup / all / ae")
	}
	startIndex, _ := strconv.Atoi(args[2])
	switch args[1] {
	case "invalid":
		sanitizeInvalidSSinDB(startIndex, c)
	case "ae":
		sanitizeAE(startIndex, c)
	default:
		sanitizeDatabase(startIndex, c)
	}
	return nil
}

func sanitizeAE(startIndex int, c tele.Context) error {
	c.Send("Started.")
	ls := queryLineS("QUERY_ALL")
	for i, l := range ls {
		if i < startIndex {
			continue
		}
		log.Infof("Checking:%s", l.tg_id)
		ss, err := c.Bot().StickerSet(l.tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Infof("SS:%s is invalid. purging it from db...", l.tg_id)
				go c.Send("purging invalid: https://t.me/addstickers/" + l.tg_id)
				deleteLineS(l.tg_id)
				deleteUserS(l.tg_id)
			} else {
				c.Send("Unknow error? " + err.Error())
				log.Errorln(err)
			}
			continue
		}
		for si := range ss.Stickers {
			if si > 0 {
				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
					log.Warnln("Setting auto emoji to FALSE for ", l.tg_id)
					updateLineSAE(false, l.tg_id)
				}
			}
		}
	}
	c.Send("Sanitize AE done!")
	return nil
}

func sanitizeInvalidSSinDB(startIndex int, c tele.Context) error {
	msg, _ := c.Bot().Send(c.Recipient(), "0")
	ls := queryLineS("QUERY_ALL")
	log.Infoln(ls)
	for i, l := range ls {
		if i < startIndex {
			continue
		}
		log.Infof("Checking:%s", l.tg_id)
		_, err := c.Bot().StickerSet(l.tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Warnf("SS:%s is invalid. purging it from db...", l.tg_id)
				go c.Send("purging: https://t.me/addstickers/" + l.tg_id)
				deleteLineS(l.tg_id)
				deleteUserS(l.tg_id)
			} else {
				go c.Send("Unknow error? " + err.Error())
				log.Errorln(err)
			}
		}
		go c.Bot().Edit(msg, "line sanitize invalid: "+strconv.Itoa(i))
	}
	us := queryUserS(-1)
	log.Infoln(us)
	for i, u := range us {
		log.Infof("Checking:%s", u.tg_id)
		_, err := c.Bot().StickerSet(u.tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Warnf("SS:%s is invalid. purging it from db...", u.tg_id)
				go c.Send("purging: https://t.me/addstickers/" + u.tg_id)
				deleteUserS(u.tg_id)
			} else {
				go c.Send("Unknow error? " + err.Error())
				log.Errorln(err)
			}
		}
		go c.Bot().Edit(msg, "user S sanitize invalid: "+strconv.Itoa(i))
	}
	c.Send("Sanitize invalid done!")
	return nil
}

func sanitizeDatabase(startIndex int, c tele.Context) error {
	msg, _ := c.Bot().Send(c.Recipient(), "0")
	ls := queryLineS("QUERY_ALL")
	log.Infoln(ls)
	for i, l := range ls {
		if i < startIndex {
			continue
		}
		log.Debugf("Scanning:%s", l.tg_id)
		ss, err := c.Bot().StickerSet(l.tg_id)
		if err != nil {
			if strings.Contains(err.Error(), "is invalid") {
				log.Infof("SS:%s is invalid. purging it from db...", l.tg_id)
				go c.Send("purging invalid: https://t.me/addstickers/" + l.tg_id)
				deleteLineS(l.tg_id)
				deleteUserS(l.tg_id)
			} else {
				c.Send("Unknow error? " + err.Error())
				log.Errorln(err)
			}
			continue
		}
		workdir := filepath.Join(dataDir, secHex(8))
		os.MkdirAll(workdir, 0755)
		for si, s := range ss.Stickers {
			if si > 0 {
				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
					log.Warnln("Setting auto emoji to FALSE for ", l.tg_id)
					updateLineSAE(false, l.tg_id)
				}
			}

			fp := filepath.Join(workdir, strconv.Itoa(si-1)+".webp")
			f := filepath.Join(workdir, strconv.Itoa(si)+".webp")
			c.Bot().Download(&s.File, f)

			if compCRC32(f, fp) {
				c.Bot().DeleteSticker(s.FileID)
				log.Warnf("Deleted on animated dup s!")
				c.Send("Deleted dup S from: https://t.me/addstickers/" + s.SetName + "  indexis: " + strconv.Itoa(si))
			}

		}
		os.RemoveAll(workdir)

		go c.Bot().Edit(msg, "line s sanitize all: "+strconv.Itoa(i))
	}
	c.Send("ALL SANITIZED!")
	return nil
}

func compCRC32(f1 string, f2 string) bool {
	fb1, err := os.ReadFile(f1)
	if err != nil {
		return false
	}
	fb2, err := os.ReadFile(f2)
	if err != nil {
		return false
	}

	c1 := crc32.ChecksumIEEE(fb1)
	c2 := crc32.ChecksumIEEE(fb2)
	log.Debugf("File:%s, C:%v", f1, c1)
	log.Debugf("File:%s, C:%v", f2, c2)

	if c1 == c2 {
		return true
	} else {
		return false
	}
}
