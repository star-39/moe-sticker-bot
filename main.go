package main

import (
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

// main.go should only handle states and basic response,
// complex operations are done in other files.

func main() {
	initLogrus()

	log.Debug("Warn: Log level below DEBUG might print sensitive information, including passwords. Use with care.")
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Please set BOT_TOKEN environment variable!! Exiting...")
		return
	}
	pref := tele.Settings{
		Token:       token,
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
		// Genrally, issues are tackled inside each state, only fatal error should be returned to framework.
		// onError will terminate current session and log to terminal.
		OnError: onError,
	}
	log.WithField("token", token).Info("Attempting to initialize...")
	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	initWorkspace(b)

	b.Handle("/quit", cmdQuit)
	b.Handle("/cancel", cmdQuit)
	b.Handle("/exit", cmdQuit)
	b.Handle("/help", cmdStart, checkState)
	b.Handle("/about", cmdAbout, checkState)
	b.Handle("/import", cmdImport, checkState)
	b.Handle("/download", cmdDownload, checkState)
	b.Handle("/create", cmdCreate, checkState)
	b.Handle("/manage", cmdManage, checkState)

	b.Handle("/start", cmdStart, checkState)
	// Handle contents.
	b.Handle(tele.OnText, handleMessage)
	b.Handle(tele.OnVideo, handleMessage)
	b.Handle(tele.OnAnimation, handleMessage)
	b.Handle(tele.OnSticker, handleMessage)
	b.Handle(tele.OnDocument, handleMessage)
	b.Handle(tele.OnPhoto, handleMessage)
	b.Handle(tele.OnCallback, handleMessage, middleware.AutoRespond(), sanitizeCallback)

	b.Start()
}

func handleMessage(c tele.Context) error {
	var err error
	command, state := getState(c)
	if command == "" {
		return handleNoState(c)
	}
	switch command {
	case "nostate":
		switch state {
		case "recvCbSLinkD":
			err = stateRecvCbSLinkD(c)
		case "recvCbSChoice":
			err = stateRecvCbSDown(c)
		case "recvCbImport":
			err = stateRecvCbImport(c)
		}
	case "import":
		switch state {
		case "recvLink":
			err = stateRecvLink(c)
		case "recvCbImport":
			err = stateRecvCbImport(c)
		case "recvTitle":
			err = stateRecvTitle(c)
		case "recvEmoji":
			err = stateRecvEmojiChoice(c)
		case "process":
			err = c.Send("processing, please wait...")
		case "recvEmojiAssign":
			err = stateRecvEmojiAssign(c)
		default:
			err = c.Send("???")
		}
	case "download":
		switch state {
		case "recvSticker":
			err = stateRecvSticker(c)
		case "recvCbSChoice":
			err = stateRecvCbSDown(c)
		default:
			err = c.Send("???")
		}
	case "create":
		switch state {
		case "recvType":
			err = stateRecvType(c)
		case "recvTitle":
			err = stateRecvTitle(c)
		case "recvFile":
			err = stateRecvFile(c)
		case "recvEmoji":
			err = stateRecvEmojiChoice(c)
		case "recvEmojiAssign":
			err = stateRecvEmojiAssign(c)
		default:
			err = c.Send("???")
		}
	case "manage":
		switch state {
		case "recvSManage":
			err = stateRecvSManage(c)
		case "recvEditChoice":
			err = stateRecvEditChoice(c)
		case "recvFile":
			err = stateRecvFile(c)
		case "recvEmoji":
			err = stateRecvEmojiChoice(c)
		case "recvEmojiAssign":
			err = stateRecvEmojiAssign(c)
		case "recvSDel":
			err = stateRecvSDel(c)
		case "recvCbDelset":
			err = stateRecvCbDelset(c)
		}
	case "modify":
		switch state {
		default:
			err = c.Send("???")
		}
	default:
		err = c.Send("???")
	}
	return err
}

func cmdManage(c tele.Context) error {
	log.Debugf("user %d entered manage with message: %s", c.Sender().ID, c.Message().Text)
	err := sendUserOwnedS(c)
	if err != nil {
		return c.Send("Sorry you have not created any sticker set yet.")
	}

	initUserData(c, "manage", "recvSManage")
	sendAskSToManage(c)
	setState(c, "recvSManage")
	return nil
}

func stateRecvSManage(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker != nil {
		ud.stickerData.sticker = c.Message().Sticker
		ud.stickerData.id = c.Message().Sticker.SetName
	} else {
		link, tp := findLinkWithType(c.Message().Text)
		if tp != "t.me" {
			return c.Send("Send correct telegram sticker link!")
		}
		ud.stickerData.id = path.Base(link)
	}
	if !matchUserS(c.Sender().ID, ud.stickerData.id) {
		return c.Send("Not owned by you. try again.")
	}
	ss, err := c.Bot().StickerSet(ud.stickerData.id)
	if err != nil {
		return c.Send("set does not exist! try again.")
	}
	ud.stickerData.cAmount = len(ss.Stickers)

	setState(c, "recvEditChoice")
	return sendAskEditChoice(c)
}

func stateRecvEditChoice(c tele.Context) error {
	if c.Callback() == nil {
		return nil
	}

	switch c.Callback().Data {
	case "add":
		setState(c, "recvFile")
		return sendAskStickerFile(c)
	case "del":
		setState(c, "recvSDel")
		return sendAskSDel(c)
	case "delset":
		setState(c, "recvCbDelset")
		return sendConfirmDelset(c)
	case "bye":
		c.Send("bye")
		terminateSession(c)
	}
	return nil
}

func stateRecvSDel(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker == nil {
		return c.Send("send sticker! try again")
	}
	if c.Message().Sticker.SetName != ud.stickerData.id {
		return c.Send("wrong sticker! tra again")
	}

	err := c.Bot().DeleteSticker(c.Message().Sticker.FileID)
	if err != nil {
		c.Send("error deleting sticker! try another one")
		return err
	}
	c.Send("Delete OK.")

	if ud.stickerData.cAmount == 1 {
		c.Send("bye")
		deleteUserS(ud.stickerData.id)
		terminateSession(c)
		return nil
	} else {
		setState(c, "recvEditChoice")
		return sendAskEditChoice(c)
	}
}

func stateRecvCbDelset(c tele.Context) error {
	if c.Callback() == nil {
		return c.Send("press a button!")
	}
	if c.Callback().Data != "yes" {
		c.Send("bye")
		terminateSession(c)
		return nil
	}
	ud := users.data[c.Sender().ID]
	c.Send("please wait...")

	ss, _ := c.Bot().StickerSet(ud.stickerData.id)
	for _, s := range ss.Stickers {
		c.Bot().DeleteSticker(s.FileID)
	}
	deleteUserS(ud.stickerData.id)
	c.Send("Delete set OK. bye")
	terminateSession(c)
	return nil
}

func handleNoState(c tele.Context) error {
	log.Debugf("user %d entered nostate with message: %s", c.Sender().ID, c.Message().Text)

	if c.Message().Sticker != nil {
		initUserData(c, "nostate", "recvCbSChoice")
		users.data[c.Sender().ID].stickerData.sticker = c.Message().Sticker
		sendAskSDownloadChoice(c)
		return nil
	}

	// bare message, we expect a link.
	link, tp := findLinkWithType(c.Message().Text)
	if link == "" {
		return sendNoStateWarning(c)
	}
	initUserData(c, "nostate", "recvCbImport")
	ud := users.data[c.Sender().ID]
	if tp == "t.me" {
		ss, err := c.Bot().StickerSet(path.Base(link))
		if err != nil {
			return nil
		}
		ud.stickerData.sticker = &ss.Stickers[0]
		setState(c, "recvCbSLinkD")
		sendAskWantSDown(c)

	} else if strings.Contains(tp, "line.me") {
		if err := parseImportLink(link, ud.lineData); err != nil {
			return nil
		}
		sendNotifySExist(c)
		setState(c, "recvCbImport")
		sendAskWantImport(c)
	} else {
		log.Debug("bad link sent barely, purging ud.")
		cleanUserData(c.Sender().ID)
		return sendNoStateWarning(c)
	}
	return nil
}

func stateRecvCbSLinkD(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}

	if c.Callback().Data == "yes" {
		setCommand(c, "download")
		downloadStickersToZip(users.data[c.Sender().ID].stickerData.sticker, true, c)
	}
	return nil
}

func stateRecvCbImport(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}

	if c.Callback().Data == "yes" {
		setCommand(c, "import")
		setState(c, "recvTitle")
		sendAskTitle_Import(c)

		go prepLineStickers(users.data[c.Sender().ID])
	}
	return nil
}

func stateRecvCbSDown(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}
	ud := users.data[c.Sender().ID]

	switch c.Callback().Data {
	case "single":
		downloadStickersToZip(ud.stickerData.sticker, false, c)
		terminateSession(c)
	case "whole":
		downloadStickersToZip(ud.stickerData.sticker, true, c)
		terminateSession(c)
	default:
		return c.Send("bad callback, try again or /quit")
	}
	return nil
}

func cmdCreate(c tele.Context) error {
	initUserData(c, "create", "recvType")
	return sendAskSTypeToCreate(c)
}

func stateRecvType(c tele.Context) error {
	if c.Callback() == nil {
		return c.Send("Please press a button.")
	}

	if strings.Contains(c.Callback().Data, "video") {
		users.data[c.Sender().ID].stickerData.isVideo = true
	}

	users.data[c.Sender().ID].stickerData.id = "sticker_" + secHex(4) + "_by_" + botName

	sendAskTitle(c)
	setState(c, "recvTitle")
	return nil
}

func stateRecvFile(c tele.Context) error {
	if c.Callback() != nil {
		return nil
	}
	log.Debug("Received file, text: ", c.Message().Text)
	if c.Message().Text == "" {
		err := appendMedia(c)
		if err != nil {
			c.Reply("Failed processing this file.")
		}

	} else if strings.Contains(c.Message().Text, "#") {
		if len(users.data[c.Sender().ID].stickerData.stickers) == 0 {
			return c.Send("No image received. try again.")
		} else {
			users.data[c.Sender().ID].stickerData.lAmount = len(users.data[c.Sender().ID].stickerData.stickers)
			setState(c, "recvEmoji")
			sendAskEmoji(c)
		}
	} else {
		c.Send("please send # mark.")
	}
	return nil
}

func cmdDownload(c tele.Context) error {
	initUserData(c, "download", "recvSticker")
	return c.Send("send sticker or share link:")
}

func stateRecvSticker(c tele.Context) error {
	log.Debugf("User %d reacted to state: recvSticker", c.Sender().ID)
	ud := users.data[c.Sender().ID]

	if c.Message().Animation != nil {
		return downloadGifToZip(c)
	}

	if c.Message().Sticker != nil {
		ud.stickerData.sticker = c.Message().Sticker
		setState(c, "recvCbSChoice")
		return sendAskSDownloadChoice(c)
	}

	if link, tp := findLinkWithType(c.Message().Text); tp == "t.me" {
		ud.stickerData.id = path.Base(link)
		ss, err := c.Bot().StickerSet(ud.stickerData.id)
		if err != nil {
			return c.Send("bad link! try again")
		}
		downloadStickersToZip(&ss.Stickers[0], true, c)
	}

	c.Send("done")
	terminateSession(c)
	return nil
}

func cmdImport(c tele.Context) error {
	initUserData(c, "import", "recvLink")
	sendAskImportLink(c)
	return nil
}

func stateRecvLink(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	link, tp := findLinkWithType(c.Message().Text)
	if tp != "line.me" {
		return c.Send("invalid link! try again or /exit")
	}

	err := parseImportLink(link, ud.lineData)
	if err != nil {
		return err
	}

	if sendNotifySExist(c) {
		setState(c, "recvCbImport")
		return sendAskWantImport(c)
	}

	setState(c, "recvTitle")
	sendAskTitle_Import(c)

	// keep preparing on background.
	return prepLineStickers(users.data[c.Sender().ID])
	return nil
}

func stateRecvTitle(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	command := ud.command

	if c.Callback() == nil {
		ud.stickerData.title = c.Message().Text
	} else {
		ud.stickerData.title = ud.lineData.title
	}

	if !checkTitle(ud.stickerData.title) {
		return c.Send("bad title! try again.")
	}

	if command != "import" {

		sendAskStickerFile(c)
		setState(c, "recvFile")
	} else {
		sendAskEmoji(c)
		setState(c, "recvEmoji")
	}

	return nil
}

func stateRecvEmojiChoice(c tele.Context) error {
	command := users.data[c.Sender().ID].command
	if c.Callback() != nil {
		switch c.Callback().Data {
		case "random":
			users.data[c.Sender().ID].stickerData.emojis = []string{"🌟"}
		case "manual":
			setState(c, "recvEmojiAssign")
			return sendAskEmojiAssign(c)
		default:
			return nil
		}
	} else {
		emojis := getEmojis(c.Message().Text)
		if emojis == "" {
			return c.Send("Send emoji or press button!")
		}
		users.data[c.Sender().ID].stickerData.emojis = []string{emojis}
	}

	setState(c, "process")

	err := execAutoCommit(!(command == "manage"), c)
	terminateSession(c)
	if err != nil {
		return err
	}
	return nil
}

func stateRecvEmojiAssign(c tele.Context) error {
	emojis := getEmojis(c.Message().Text)
	if emojis == "" {
		return c.Send("send emoji! try again.")
	}
	return execEmojiAssign(!(users.data[c.Sender().ID].command == "manage"), emojis, c)
}

func cmdStart(c tele.Context) error {
	return sendStartMessage(c)
}

func cmdQuit(c tele.Context) error {
	if terminateSession(c) {
		return c.Send("Exited")
	}
	return nil
}

func cmdAbout(c tele.Context) error {
	sendAboutMessage(c)
	return nil
}

func terminateSession(c tele.Context) bool {
	return cleanUserData(c.Sender().ID)
	// c.Send("bye")
}

func onError(err error, c tele.Context) {
	log.Error("User encountered fatal error!")
	log.Error(err)
	sendFatalError(err, c)
	terminateSession(c)
}
