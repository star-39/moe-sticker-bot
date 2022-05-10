package main

import (
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
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
	log.WithFields(log.Fields{"botName": botName, "dataDir": dataDir}).Info("Bot OK.")

	b.Handle("/quit", cmdQuit)
	b.Handle("/cancel", cmdQuit)
	b.Handle("/exit", cmdQuit)
	b.Handle("/help", cmdStart, checkState)
	b.Handle("/about", cmdAbout, checkState)
	b.Handle("/faq", cmdAbout, checkState)
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
	b.Handle(tele.OnCallback, handleMessage, autoRespond, sanitizeCallback)

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
		case "process":
			err = c.Send("processing, please wait...")
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
		if tp != LINK_TG {
			return c.Send("Send correct telegram sticker link!")
		}
		ud.stickerData.id = path.Base(link)
	}
	if !matchUserS(c.Sender().ID, ud.stickerData.id) {
		return c.Send("Not owned by you. try again or /quit")
	}
	ss, err := c.Bot().StickerSet(ud.stickerData.id)
	if err != nil {
		return c.Send("set does not exist! try again or /quit")
	}
	ud.stickerData.cAmount = len(ss.Stickers)
	ud.stickerData.isVideo = ss.Video

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
		terminateSession(c)
	}
	return nil
}

func stateRecvSDel(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker == nil {
		return c.Send("send sticker! try again or /quit")
	}
	if c.Message().Sticker.SetName != ud.stickerData.id {
		return c.Send("wrong sticker! try again or /quit")
	}

	err := c.Bot().DeleteSticker(c.Message().Sticker.FileID)
	if err != nil {
		c.Send("error deleting sticker! try another one or /quit")
		return err
	}
	c.Send("Delete OK.")

	if ud.stickerData.cAmount == 1 {
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
		ud := initUserData(c, "nostate", "recvCbSChoice")
		ud.stickerData.sticker = c.Message().Sticker
		sendAskSDownloadChoice(c)
		return nil
	}

	// bare message, we expect a link.
	link, tp := findLinkWithType(c.Message().Text)
	switch tp {
	case LINK_TG:
		ss, err := c.Bot().StickerSet(path.Base(link))
		if err != nil {
			return nil
		}
		ud := initUserData(c, "nostate", "recvCbSLinkD")
		ud.stickerData.sticker = &ss.Stickers[0]
		sendAskWantSDown(c)
	case LINK_LINE:
		ud := initUserData(c, "nostate", "recvCbImport")
		if err := parseImportLink(link, ud.lineData); err != nil {
			endSession(c)
			return nil
		}
		sendNotifySExist(c)
		sendAskWantImport(c)
	default:
		log.Debug("bad link sent barely, purging ud.")
		return sendNoStateWarning(c)
	}
	return nil
}

func stateRecvCbSLinkD(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}

	switch c.Callback().Data {
	case "yes":
		setCommand(c, "download")
		downloadStickersToZip(users.data[c.Sender().ID].stickerData.sticker, true, c)
	case "bye":
		terminateSession(c)
	}
	return nil
}

func stateRecvCbImport(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}

	switch c.Callback().Data {
	case "yes":
		setCommand(c, "import")
		setState(c, "recvTitle")
		sendAskTitle_Import(c)
		return prepLineStickers(users.data[c.Sender().ID], true)
	case "bye":
		terminateSession(c)
	}
	return nil
}

func stateRecvCbSDown(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoState(c)
	}
	ud := users.data[c.Sender().ID]
	var err error

	switch c.Callback().Data {
	case "single":
		setCommand(c, "download")
		err = downloadStickersToZip(ud.stickerData.sticker, false, c)
	case "whole":
		setCommand(c, "download")
		err = downloadStickersToZip(ud.stickerData.sticker, true, c)
	case "bye":
	default:
		return c.Send("bad callback, try again or /quit")
	}
	terminateSession(c)
	return err
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

	if c.Message().Media() != nil {
		err := appendMedia(c)
		if err != nil {
			c.Reply("Failed processing this file. ERR:" + err.Error())
		}
		return nil
	}
	if !strings.Contains(c.Message().Text, "#") {
		return c.Send("please send # mark.")
	}
	if len(users.data[c.Sender().ID].stickerData.stickers) == 0 {
		return c.Send("No image received. try again or /quit")
	}

	users.data[c.Sender().ID].stickerData.lAmount = len(users.data[c.Sender().ID].stickerData.stickers)
	setState(c, "recvEmoji")
	sendAskEmoji(c)

	return nil
}

func cmdDownload(c tele.Context) error {
	initUserData(c, "download", "recvSticker")
	return sendAskWhatToDownload(c)
}

func stateRecvSticker(c tele.Context) error {
	log.Debugf("User %d reacted to state: recvSticker", c.Sender().ID)
	ud := users.data[c.Sender().ID]
	var err error
	link, tp := findLinkWithType(c.Message().Text)

	switch {
	case c.Message().Animation != nil:
		err = downloadGifToZip(c)
	case c.Message().Sticker != nil:
		ud.stickerData.sticker = c.Message().Sticker
		setState(c, "recvCbSChoice")
		return sendAskSDownloadChoice(c)
	case tp == LINK_TG:
		ud.stickerData.id = path.Base(link)
		ss, sserr := c.Bot().StickerSet(ud.stickerData.id)
		if sserr != nil {
			return c.Send("bad link! try again or /quit")
		}
		err = downloadStickersToZip(&ss.Stickers[0], true, c)
	case tp == LINK_LINE:
		c.Send("Please wait...")
		err = parseImportLink(link, ud.lineData)
		if err != nil {
			return err
		}
		err = prepLineStickers(ud, false)
		if err != nil {
			return err
		}
		err = downloadLineSToZip(c, ud)
	default:
		return c.Send("send link or sticker or GIF or /exit ")
	}

	if err != nil {
		return err
	}
	terminateSession(c)
	return err
}

func cmdImport(c tele.Context) error {
	initUserData(c, "import", "recvLink")
	sendAskImportLink(c)
	return nil
}

func stateRecvLink(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	link, tp := findLinkWithType(c.Message().Text)
	if tp != LINK_LINE {
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

	return prepLineStickers(users.data[c.Sender().ID], true)
}

func stateRecvTitle(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	command := ud.command

	if c.Callback() == nil {
		ud.stickerData.title = c.Message().Text
	}

	if !checkTitle(ud.stickerData.title) {
		return c.Send("bad title! try again or /quit")
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
			sendProcessStarted(c, "preparing...")
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
	endSession(c)
	if err != nil {
		return err
	}
	return nil
}

func stateRecvEmojiAssign(c tele.Context) error {
	emojis := getEmojis(c.Message().Text)
	if emojis == "" {
		return c.Send("send emoji! try again or /quit")
	}
	return execEmojiAssign(!(users.data[c.Sender().ID].command == "manage"), emojis, c)
}

func cmdStart(c tele.Context) error {
	return sendStartMessage(c)
}

func cmdQuit(c tele.Context) error {
	log.Debug("Received user quit request.")
	ud, exist := users.data[c.Sender().ID]
	if !exist {
		return c.Send("Please use /start")
	}
	c.Send("Please wait...")
	ud.cancel()
	ud.udWg.Wait()
	for _, s := range ud.stickerData.stickers {
		s.wg.Wait()
	}
	terminateSession(c)
	return nil
}

func cmdAbout(c tele.Context) error {
	sendAboutMessage(c)
	return nil
}

// This one never say goodbye.
func endSession(c tele.Context) {
	forceCleanUserData(c.Sender().ID)
}

// This one will say goodbye.
func terminateSession(c tele.Context) {
	forceCleanUserData(c.Sender().ID)
	c.Send("Bye. /start")
}

func onError(err error, c tele.Context) {
	sendFatalError(err, c)
	forceCleanUserData(c.Sender().ID)
}
