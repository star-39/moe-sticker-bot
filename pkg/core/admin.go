package core

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func cmdStatRep(c tele.Context) error {
	// Report status.
	stat := []string{}
	py_emoji_ok, _ := httpGet("http://127.0.0.1:5000/status")
	stat = append(stat, "py_emoji_ok? :"+py_emoji_ok)
	return c.Send(strings.Join(stat, "\n"))
}
