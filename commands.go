package main

import (
	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

func cmdCreate(c tele.Context) error {
	initUserData(c, "create", "waitSType")
	return sendAskSTypeToCreate(c)
}

func cmdManage(c tele.Context) error {
	err := sendUserOwnedS(c)
	if err != nil {
		return c.Send("Sorry you have not created any sticker set yet.")
	}

	initUserData(c, "manage", "waitSManage")
	sendAskSToManage(c)
	setState(c, "waitSManage")
	return nil
}

func cmdImport(c tele.Context) error {
	initUserData(c, "import", "waitImportLink")
	sendAskImportLink(c)
	return nil
}
func cmdAbout(c tele.Context) error {
	sendAboutMessage(c)
	return nil
}

func cmdFAQ(c tele.Context) error {
	sendFAQ(c)
	return nil
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
	// for _, s := range ud.stickerData.stickers {
	// 	s.wg.Wait()
	// }
	terminateSession(c)
	return nil
}
