package core

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func cmdSitRep(c tele.Context) error {
	// Report status.
	stat := []string{}
	py_emoji_ok, _ := httpGet("http://127.0.0.1:5000/status")
	stat = append(stat, "py_emoji_ok? :"+py_emoji_ok)
	c.Send(strings.Join(stat, "\n"))

	return nil
}

func cmdGetFID(c tele.Context) error {
	initUserData(c, "getfid", "waitMFile")
	if c.Message().Media() != nil {
		return c.Reply(c.Message().Media().MediaFile().FileID)
	} else {
		return nil
	}
}
