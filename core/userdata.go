package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

func cleanUserDataAndDir(uid int64) bool {
	log.WithField("uid", uid).Debugln("Purging userdata...")
	_, exist := users.data[uid]
	if exist {
		os.RemoveAll(users.data[uid].workDir)
		users.mu.Lock()
		delete(users.data, uid)
		users.mu.Unlock()
		log.WithField("uid", uid).Debugln("Userdata purged from map and disk.")
		return true
	} else {
		log.WithField("uid", uid).Debugln("Userdata does not exisst, do nothing.")
		return false
	}
}

func cleanUserData(uid int64) bool {
	log.WithField("uid", uid).Debugln("Purging userdata...")
	_, exist := users.data[uid]
	if exist {
		users.mu.Lock()
		delete(users.data, uid)
		users.mu.Unlock()
		log.WithField("uid", uid).Debugln("Userdata purged from map.")
		return true
	} else {
		log.WithField("uid", uid).Debugln("Userdata does not exist, do nothing.")
		return false
	}
}

func initUserData(c tele.Context, command string, state string) *UserData {
	uid := c.Sender().ID
	users.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	sID := secHex(6)
	users.data[uid] = &UserData{
		state:     state,
		sessionID: sID,
		// userDir:       filepath.Join(dataDir, strconv.FormatInt(uid, 10)),
		workDir:     filepath.Join(dataDir, sID),
		command:     command,
		lineData:    &msbimport.LineData{},
		stickerData: &StickerData{},
		// stickerManage: &StickerManage{},
		ctx:    ctx,
		cancel: cancel,
	}
	users.mu.Unlock()
	// Do not anitize user work directory.
	// os.RemoveAll(users.data[uid].userDir)
	os.MkdirAll(users.data[uid].workDir, 0755)
	return users.data[uid]
}

func getState(c tele.Context) (string, string) {
	ud, exist := users.data[c.Sender().ID]
	if exist {
		return ud.command, ud.state
	} else {
		return "", ""
	}
}

func checkState(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		//If bot is summoned from group chat, check command.
		if c.Chat().Type == tele.ChatGroup || c.Chat().Type == tele.ChatSuperGroup {
			log.Debugf("User %d attempted command from group chat.", c.Sender().ID)
			//For group chat, support /search only.
			if strings.HasPrefix(c.Text(), "/search@"+botName) {
				return next(c)
			} else if strings.Contains(c.Text(), "@"+botName) {
				//has metion
				return sendUnsupportedCommandForGroup(c)
			} else {
				//do nothing
				return nil
			}
		}

		command, _ := getState(c)
		if command == "" {
			log.Debugf("User %d entering command with message: %s", c.Sender().ID, c.Message().Text)
			return next(c)
		} else {
			log.Debugf("User %d already in command: %v", c.Sender().ID, command)
			return sendInStateWarning(c)
		}
	}
}

func setState(c tele.Context, state string) {
	if c == nil {
		return
	}
	ud, ok := users.data[c.Sender().ID]
	if !ok {
		return
	}
	ud.state = state
}

// func setCommand(c tele.Context, command string) {
// 	uid := c.Sender().ID
// 	users.data[uid].command = command
// }
