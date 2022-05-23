package main

import (
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

func handleNoSession(c tele.Context) error {
	log.Debugf("user %d entered no session with message: %s", c.Sender().ID, c.Message().Text)

	if c.Callback() != nil && c.Message().ReplyTo != nil {
		switch c.Callback().Data {
		case "single":
			return downloadStickersToZip(c.Message().ReplyTo.Sticker, "", c)
		case "whole":
			id := getSIDFromMessage(c.Message().ReplyTo)
			return downloadStickersToZip(nil, id, c)
		case "manage":
			ud := initUserData(c, "manage", "waitCbEditChoice")
			id := getSIDFromMessage(c.Message().ReplyTo)
			err := retrieveSSDetails(c, id, ud.stickerData)
			if err == nil {
				return sendAskEditChoice(c)
			} else {
				endSession(c)
				return c.Send("Wrong stickter set?")
			}
		case "yesimport":
			ud := initUserData(c, "import", "waitSTitle")
			parseImportLink(findLink(c.Message().ReplyTo.Text), ud.lineData)
			sendAskTitle_Import(c)
			return prepImportStickers(ud, true)
		}
	}

	// bare sticker, ask user's choice.
	if c.Message().Sticker != nil {
		if matchUserS(c.Sender().ID, c.Message().Sticker.SetName) {
			return sendAskSChoice(c)
		} else {
			return sendAskSDownloadChoice(c)
		}
	}

	if c.Message().Animation != nil {
		return downloadGifToZip(c)
	}

	// bare message, we expect a link.
	link, tp := findLinkWithType(c.Message().Text)
	switch tp {
	case LINK_TG:
		if matchUserS(c.Sender().ID, path.Base(link)) {
			return sendAskTGLinkChoice(c)
		} else {
			return sendAskWantSDown(c)
		}
	case LINK_IMPORT:
		ld := &LineData{}
		err := parseImportLink(link, ld)
		if err != nil {
			return sendBadImportLinkWarn(c)
		} else {
			sendNotifySExist(c, ld.id)
			return sendAskWantImport(c)
		}
	default:
		return sendNoSessionWarning(c)
	}
}

func stateProcessing(c tele.Context) error {
	if c.Callback() != nil {
		if c.Callback().Data == "bye" {
			return cmdQuit(c)
		}
	}
	return c.Send("processing, please wait... 作業中, 請稍後... /quit")
}

func waitSEmojiEdit(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker == nil {
		return c.Send("Send sticker! try again or /quit")
	}
	if c.Message().Sticker.SetName != ud.stickerData.id {
		return c.Send("Sticker from wrong set! try again or /quit")
	}
	ud.stickerManage.pendingS = c.Message().Sticker
	setState(c, "waitEmojiEdit")
	return sendAskEmojiEdit(c)
}

func waitEmojiEdit(c tele.Context) error {
	log.Debug("Attempting edit emoji.")
	ud := users.data[c.Sender().ID]
	setState(c, "process")
	c.Send("please wait...")
	err := editStickerEmoji(c, ud)
	if err != nil {
		c.Send("Edit Error! try again or /quit\n\n" + err.Error())
	} else {
		c.Send("Edit emoji OK.")
		sendSEditOK(c)
	}
	setState(c, "waitCbEditChoice")
	return sendAskEditChoice(c)
}

func waitSMovFrom(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker == nil {
		return c.Send("Send sticker! try again or /quit")
	}
	if c.Message().Sticker.SetName != ud.stickerData.id {
		return c.Send("Sticker from wrong set! try again or /quit")
	}

	ud.stickerManage.pendingS = c.Message().Sticker
	setState(c, "waitSMovTo")
	return sendAskMovTarget(c)
}

func waitSMovTo(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Message().Sticker == nil {
		return c.Send("Send sticker! try again or /quit")
	}
	if c.Message().Sticker.SetName != ud.stickerData.id {
		return c.Send("Sticker from wrong set! try again or /quit")
	}

	ss, err := c.Bot().StickerSet(c.Message().Sticker.SetName)
	if err != nil {
		return err
	}
	targetPos := -1
	for i, s := range ss.Stickers {
		if s.FileID == c.Message().Sticker.FileID {
			targetPos = i
		}
	}
	log.Debugf("Moving sticker to %d", targetPos)
	err = c.Bot().SetStickerPosition(ud.stickerManage.pendingS.FileID, targetPos)
	if err != nil {
		return err
	}

	c.Send("Move sticker OK.")
	sendSEditOK(c)
	setState(c, "waitCbEditChoice")
	return sendAskEditChoice(c)
}

func waitSManage(c tele.Context) error {
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
	// Allow admin to manage all sticker sets.
	if c.Sender().ID == ADMIN_UID {
		goto NEXT
	}
	if !matchUserS(c.Sender().ID, ud.stickerData.id) {
		return c.Send("Not owned by you. try again or /quit")
	}

NEXT:
	err := retrieveSSDetails(c, ud.stickerData.id, ud.stickerData)
	if err != nil {
		return c.Send("sticker set wrong! try again or /quit")
	}

	setState(c, "waitCbEditChoice")
	return sendAskEditChoice(c)
}

func waitCbEditChoice(c tele.Context) error {
	if c.Callback() == nil {
		return sendNoCbWarn(c)
	}

	switch c.Callback().Data {
	case "add":
		setState(c, "waitSFile")
		return sendAskStickerFile(c)
	case "del":
		setState(c, "waitSDel")
		return sendAskSDel(c)
	case "delset":
		setState(c, "waitCbDelset")
		return sendConfirmDelset(c)
	case "mov":
		setState(c, "waitSMovFrom")
		return sendAskSFrom(c)
	case "emoji":
		setState(c, "waitSEmojiEdit")
		return sendSEditEmoji(c)
	case "bye":
		terminateSession(c)
	}
	return nil
}

func waitSDel(c tele.Context) error {
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
	sendSEditOK(c)
	ud.stickerData.cAmount--
	if ud.stickerData.cAmount == 0 {
		deleteUserS(ud.stickerData.id)
		deleteLineS(ud.stickerData.id)
		terminateSession(c)
		return nil
	} else {
		setState(c, "waitCbEditChoice")
		return sendAskEditChoice(c)
	}
}

func waitCbDelset(c tele.Context) error {
	if c.Callback() == nil {
		return c.Send("press a button!")
	}
	if c.Callback().Data != "yes" {
		terminateSession(c)
		return nil
	}
	ud := users.data[c.Sender().ID]
	setState(c, "process")
	c.Send("please wait...")

	ss, _ := c.Bot().StickerSet(ud.stickerData.id)
	for _, s := range ss.Stickers {
		c.Bot().DeleteSticker(s.FileID)
	}
	deleteUserS(ud.stickerData.id)
	deleteLineS(ud.stickerData.id)
	c.Send("Delete set OK. bye")
	terminateSession(c)
	return nil
}

func waitCbImportChoice(c tele.Context) error {
	if c.Callback() == nil {
		return handleNoSession(c)
	}
	ud := users.data[c.Sender().ID]
	switch c.Callback().Data {
	case "yesimport":
		setCommand(c, "import")
		setState(c, "waitSTitle")
		sendAskTitle_Import(c)
		return prepImportStickers(ud, true)
	case "bye":
		terminateSession(c)
	}
	return nil
}

func waitSType(c tele.Context) error {
	if c.Callback() == nil {
		return c.Send("Please press a button.")
	}

	if strings.Contains(c.Callback().Data, "video") {
		users.data[c.Sender().ID].stickerData.isVideo = true
	}

	sendAskTitle(c)
	setState(c, "waitSTitle")
	return nil
}

func waitSFile(c tele.Context) error {
	if c.Callback() != nil {
		switch c.Callback().Data {
		case "done":
			goto NEXT
		case "bye":
			terminateSession(c)
			return nil
		default:
			return sendPromptStopAdding(c)
		}
	}
	if c.Message().Media() != nil {
		err := appendMedia(c)
		if err != nil {
			c.Reply("Failed processing this file. 處理此檔案時錯誤:\n" + err.Error())
		}
		return nil
	} else {
		return sendPromptStopAdding(c)
	}
NEXT:
	if len(users.data[c.Sender().ID].stickerData.stickers) == 0 {
		return c.Send("No image received. try again or /quit")
	}

	setState(c, "waitEmojiChoice")
	sendAskEmoji(c)

	return nil
}

func waitSDownload(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	var err error
	link, tp := findLinkWithType(c.Message().Text)

	switch {
	case c.Message().Animation != nil:
		err = downloadGifToZip(c)
	case c.Message().Sticker != nil:
		// We handle this in "handleNoSession"
		endSession(c)
		return sendAskSDownloadChoice(c)
	case tp == LINK_TG:
		ud.stickerData.id = path.Base(link)
		ss, sserr := c.Bot().StickerSet(ud.stickerData.id)
		if sserr != nil {
			return c.Send("bad link! try again or /quit")
		}
		err = downloadStickersToZip(nil, ss.Name, c)
	case tp == LINK_IMPORT:
		c.Send("Please wait...")
		err = parseImportLink(link, ud.lineData)
		if err != nil {
			return err
		}
		err = prepImportStickers(ud, false)
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

func waitImportLink(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	link, tp := findLinkWithType(c.Message().Text)
	if tp != LINK_IMPORT {
		return sendBadImportLinkWarn(c)
	}

	err := parseImportLink(link, ud.lineData)
	if err != nil {
		return err
	}

	if sendNotifySExist(c, ud.lineData.id) {
		setState(c, "waitCbImportChoice")
		return sendAskWantImport(c)
	}

	setState(c, "waitSTitle")
	sendAskTitle_Import(c)
	return prepImportStickers(ud, true)

}

func waitSTitle(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	command := ud.command

	if c.Callback() == nil {
		ud.stickerData.title = c.Message().Text
	}

	if !checkTitle(ud.stickerData.title) {
		return c.Send("bad title! try again or /quit")
	}

	switch command {
	case "import":
		setState(c, "waitEmojiChoice")
		return sendAskEmoji(c)
		// return prepImportStickers(users.data[c.Sender().ID], true)
	case "create":
		setState(c, "waitSID")
		sendAskID(c)
	}

	return nil
}

func waitSID(c tele.Context) error {
	var id string
	if c.Callback() != nil {
		if c.Callback().Data == "auto" {
			users.data[c.Sender().ID].stickerData.id = "sticker_" + secHex(4) + "_by_" + botName
			goto NEXT
		}
	}

	id = regexAlphanum.FindString(c.Message().Text)
	if !checkID(id) {
		return sendBadIDWarn(c)
	}
	id = id + "_by_" + botName
	if _, err := c.Bot().StickerSet(id); err == nil {
		return sendIDOccupiedWarn(c)
	}
	users.data[c.Sender().ID].stickerData.id = id

NEXT:
	setState(c, "waitSFile")
	return sendAskStickerFile(c)
}

func waitEmojiChoice(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	if c.Callback() != nil {
		switch c.Callback().Data {
		case "random":
			users.data[c.Sender().ID].stickerData.emojis = []string{"⭐"}
		case "manual":
			sendProcessStarted(ud, c, "preparing...")
			setState(c, "waitSEmojiAssign")
			return sendAskEmojiAssign(c)
		default:
			return nil
		}
	} else {
		emojis := findEmojis(c.Message().Text)
		if emojis == "" {
			return c.Send("Send emoji or press button!")
		}
		users.data[c.Sender().ID].stickerData.emojis = []string{emojis}
	}

	setState(c, "process")

	err := execAutoCommit(!(ud.command == "manage"), c)
	endSession(c)
	if err != nil {
		return err
	}
	return nil
}

func waitSEmojiAssign(c tele.Context) error {
	emojis := findEmojis(c.Message().Text)
	if emojis == "" {
		return c.Send("send emoji! try again or /quit")
	}
	return execEmojiAssign(!(users.data[c.Sender().ID].command == "manage"), emojis, c)
}
