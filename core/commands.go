package core

import (
	"strings"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func cmdCreate(c tele.Context) error {
	initUserData(c, "create", "waitSType")
	return sendAskSTypeToCreate(c)
}

func cmdManage(c tele.Context) error {
	err := sendUserOwnedS(c)
	if err != nil {
		return sendNoSToManage(c)
	}

	initUserData(c, "manage", "waitSManage")
	sendAskSToManage(c)
	setState(c, "waitSManage")
	return nil
}

func cmdImport(c tele.Context) error {
	// V2.2: Do not init command on /import
	// initUserData(c, "import", "waitImportLink")
	return sendAskImportLink(c)
}

func cmdDownload(c tele.Context) error {
	// V2.2: Do not init command on /download
	// initUserData(c, "download", "waitSDownload")
	return sendAskWhatToDownload(c)
}

func cmdAbout(c tele.Context) error {
	sendAboutMessage(c)
	return nil
}

func cmdFAQ(c tele.Context) error {
	sendFAQ(c)
	return nil
}

func cmdChangelog(c tele.Context) error {
	return sendChangelog(c)
}

func cmdStart(c tele.Context) error {
	return sendStartMessage(c)
}

func cmdSearch(c tele.Context) error {
	if c.Chat().Type == tele.ChatGroup || c.Chat().Type == tele.ChatSuperGroup {
		return cmdGroupSearch(c)
	}
	initUserData(c, "search", "waitSearchKW")
	return sendAskSearchKeyword(c)
}

func cmdGroupSearch(c tele.Context) error {
	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return sendBadSearchKeyword(c)
	}
	keywords := args[1:]
	lines := searchLineS(keywords)
	if len(lines) == 0 {
		return sendSearchNoResult(c)
	}
	return sendSearchResult(10, lines, c)
}

func cmdQuit(c tele.Context) error {
	log.Debug("Received user quit request.")
	ud, exist := users.data[c.Sender().ID]
	if !exist {
		return c.Send("Please use /start", &tele.ReplyMarkup{RemoveKeyboard: true})
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
