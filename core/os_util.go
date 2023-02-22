package core

import (
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func purgeOutdatedStorageData() {
	dirEntries, _ := os.ReadDir(dataDir)
	for _, f := range dirEntries {
		if !f.IsDir() {
			continue
		}
		fInfo, _ := f.Info()
		fMTime := fInfo.ModTime().Unix()
		fPath := filepath.Join(dataDir, f.Name())
		// 2 Days
		if fMTime < (time.Now().Unix() - 172800) {
			os.RemoveAll(fPath)
			users.mu.Lock()
			for uid, ud := range users.data {
				if ud.sessionID == f.Name() {
					log.Warnf("Found outdated user data. Purging from map as well. SID:%s, UID:%d", ud.sessionID, uid)
					delete(users.data, uid)
					break
				}
			}
			users.mu.Unlock()
			log.Infoln("Purged outdated user dir:", fPath)
		}
	}

	if msbconf.WebappDataDir != "" {
		webappDataDirEntries, _ := os.ReadDir(msbconf.WebappDataDir)
		for _, f := range webappDataDirEntries {
			if !f.IsDir() {
				continue
			}
			fInfo, _ := f.Info()
			fMTime := fInfo.ModTime().Unix()
			fPath := filepath.Join(msbconf.WebappDataDir, f.Name())
			// 2 Days
			if fMTime < (time.Now().Unix() - 172800) {
				os.RemoveAll(fPath)
				log.Infoln("Purged outdated webapp data dir:", fPath)
			}
		}
	}

	if msbconf.BotApiDir != "" {
		filepath.Walk(msbconf.BotApiDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			// fInfo, _ := f.Info()
			fMTime := info.ModTime().Unix()
			// 2 Days
			if fMTime < (time.Now().Unix() - 172800) {
				os.RemoveAll(path)
				log.Infoln("Purged outdated LocalBotApiDir:", path)
			}
			return nil
		})
	}
}
