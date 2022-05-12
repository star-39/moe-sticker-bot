package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

func cmdSanitize(c tele.Context) error {
	adminUID, _ := strconv.ParseInt(os.Getenv("ADMIN_UID"), 10, 64)
	if adminUID != c.Sender().ID {
		return c.Send("Admin only command. /start")
	}

	return sanitizeDatabase(c)
}

func sanitizeDatabase(c tele.Context) error {
	status := 0
	msg, _ := c.Bot().Send(c.Recipient(), "0")
	ls := queryAllLineS()
	log.Debugln(ls)
	for i, l := range ls {
		status++
		log.Debugf("Scanning:%s", l.tg_id)
		ss, err := c.Bot().StickerSet(l.tg_id)
		if err != nil {
			continue
		}
		workdir := filepath.Join(dataDir, secHex(8))
		os.MkdirAll(workdir, 0755)
		for si, s := range ss.Stickers {
			if si > 0 {
				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
					log.Debugln("Setting auto emoji to FALSE for ", l.tg_id)
					updateLineSAE(false, l.tg_id)
				}
			}
			fp := filepath.Join(workdir, strconv.Itoa(si-1))
			f := filepath.Join(workdir, strconv.Itoa(si))
			c.Bot().Download(&s.File, f)
			if ss.Video {
				out, _ := exec.Command("compare", "-metric", "MAE", fp, f, "/dev/null").CombinedOutput()
				out2, _ := exec.Command("compare", "-metric", "MAE", fp+"[1]", f+"[1]", "/dev/null").CombinedOutput()
				out3, _ := exec.Command("compare", "-metric", "MAE", fp+"[10]", f+"[10]", "/dev/null").CombinedOutput()
				if strings.Contains(string(out), "0 (0)") && (string(out) == string(out2)) && (string(out) == string(out3)) {
					c.Bot().DeleteSticker(s.FileID)
					c.Send("Deleted on animated dup s from: https://t.me/addstickers/" + s.SetName + "  indexis: " + strconv.Itoa(si))
				}
				log.Debugf(string(out))
			} else {
				out, _ := exec.Command("compare", "-metric", "MAE", fp, f, "/dev/null").CombinedOutput()
				log.Debugf(string(out))
				if strings.Contains(string(out), "0 (0)") {
					c.Bot().DeleteSticker(s.FileID)
					c.Send("Deleted on dup s from: https://t.me/addstickers/" + s.SetName + "  indexis: " + strconv.Itoa(si))
				}
			}
		}
		os.RemoveAll(workdir)

		if status == 50 {
			status = 0
			c.Bot().Edit(msg, strconv.Itoa(i))
		}
	}
	c.Send("ALL SANITIZED!")
	return nil
}
