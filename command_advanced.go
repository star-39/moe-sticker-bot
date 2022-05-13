package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	tele "github.com/star-39/telebot"
)

func cmdRegister(c tele.Context) error {
	c.Send("Entering advanced command: register. You can use register your line sticker set to bot's database.")
	c.Send("Use with care, if you are not sure, send /quit .")

	initUserData(c, "register", "waitRegLineLink")
	return c.Send("Send LINE link now:")
}

func waitRegLineLink(c tele.Context) error {
	link := findLink(c.Message().Text)

	ud := users.data[c.Sender().ID]
	err := parseImportLink(link, ud.lineData)
	if err != nil {
		return c.Send("Wrong link! Try again.")
	}

	setState(c, "waitRegS")
	c.Send(fmt.Sprintf("Parsed LINE Type:%s, ID:%s, Title:%s", ud.lineData.category, ud.lineData.id, ud.lineData.title))
	return c.Send("Send TG Sticker or link now:")
}

func waitRegS(c tele.Context) error {
	id := ""
	ud := users.data[c.Sender().ID]

	if c.Message().Sticker != nil {
		id = c.Message().Sticker.SetName
	} else {
		id = findLink(c.Message().Text)
		id = path.Base(id)
	}
	if id == "" || !strings.HasSuffix(id, "_by_"+botName) {
		return c.Send("Bad TG Sticker! Try again.")
	}

	ss, err := c.Bot().StickerSet(id)
	if err != nil {
		return c.Send("Bad TG Sticker! Try again.")
	}

	ae := true
	if ss.Stickers[0].Emoji != ss.Stickers[1].Emoji || ss.Stickers[1].Emoji != ss.Stickers[2].Emoji || ss.Stickers[2].Emoji != ss.Stickers[3].Emoji {
		ae = false
	}

	lsqs := queryLineS(ud.lineData.id)
	for _, lsq := range lsqs {
		if lsq.tg_id == id {
			c.Send("Already exist in line db!! try another one.")
			goto INSERT_USER_S
		}
	}
	insertLineS(ud.lineData.id, ud.lineData.link, id, ss.Title, ae)

INSERT_USER_S:
	usq := queryUserS(c.Sender().ID)
	for _, us := range usq {
		if us.tg_id == id {
			c.Send("Already exists in user db! try another one")
			goto RETURN
		}
	}
	insertUserS(c.Sender().ID, id, ss.Title, time.Now().Unix())
	c.Send("Insert to database OK!")

RETURN:
	c.Send("Returning back to cmdRegister")
	return cmdRegister(c)
}
