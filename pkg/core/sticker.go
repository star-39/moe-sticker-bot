package core

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func execAutoCommit(createSet bool, c tele.Context) error {
	ud := users.data[c.Sender().ID]
	ud.udWg.Add(1)
	defer ud.udWg.Done()

	sendProcessStarted(ud, c, "")
	ud.wg.Wait()

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!")
		return errors.New("no sticker available")
	}

	log.Debugln("stickerData summary:")
	log.Debugln(ud.stickerData)
	committedStickers := 0
	errorCount := 0
	flCount := 0

	for index, sf := range ud.stickerData.stickers {
		select {
		case <-ud.ctx.Done():
			log.Warn("execAutoCommit received ctxDone!")
			return nil
		default:
		}
		var err error
		ss := tele.StickerSet{
			Name:   ud.stickerData.id,
			Title:  ud.stickerData.title,
			Emojis: ud.stickerData.emojis[0],
		}
		go editProgressMsg(index, len(ud.stickerData.stickers), "", "", nil, c)
		if index == 0 && createSet {
			err = commitSticker(true, &flCount, false, sf, c, ss)
			if err != nil {
				log.Errorln("create failed. ", err)
				return err
			} else {
				committedStickers += 1
			}
		} else {
			err = commitSticker(false, &flCount, false, sf, c, ss)
			if err != nil {
				log.Warnln("a sticker failed to add. ", err)
				c.Send("one sticker failed to add, index is:" + strconv.Itoa(index))
				errorCount += 1
			} else {
				committedStickers += 1
			}

			if errorCount > 2 {
				return errors.New("too many errors when adding sticker")
			}
			if flCount > 3 {
				sendTooManyFloodLimits(c)
				return errors.New("too many flood limits")
			}

		}
		log.Debugln("one sticker commited. count: ", committedStickers)
	}
	if createSet {
		if ud.command == "import" {
			insertLineS(ud.lineData.id, ud.lineData.link, ud.stickerData.id, ud.stickerData.title, true)
		}
		insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
	}
	editProgressMsg(0, 0, "Success! /start", "", nil, c)
	sendSFromSS(c)
	return nil
}

func execEmojiAssign(createSet bool, emojis string, c tele.Context) error {
	ud := users.data[c.Sender().ID]
	ud.wg.Wait()

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!!")
		return errors.New("no sticker available")
	}
	var err error
	ss := tele.StickerSet{
		Name:   ud.stickerData.id,
		Title:  ud.stickerData.title,
		Emojis: emojis,
	}

	sf := ud.stickerData.stickers[ud.stickerData.pos]
	log.Debugln("ss summary:")
	log.Debugln(ss)

	if createSet && ud.stickerData.pos == 0 {
		err = commitSticker(true, new(int), false, sf, c, ss)
		if err != nil {
			log.Errorln("create failed. ", err)
			return err
		} else {
			ud.stickerData.cAmount += 1
		}
	} else {
		err = commitSticker(false, new(int), false, sf, c, ss)
		if err != nil {
			if strings.Contains(err.Error(), "invalid sticker emojis") {
				return c.Send("Sorry, this emoji is invalid. Try another one.\n抱歉, 這個emoji無效, 請另試一次.")
			}
			c.Send("one sticker failed to add, index is:" + strconv.Itoa(ud.stickerData.pos))
			log.Warnln("a sticker failed to add. ", err)
		} else {
			ud.stickerData.cAmount += 1
		}
	}

	log.Debugf("one sticker commit attempted. pos:%d, lAmount:%d, cAmount:%d", ud.stickerData.pos, ud.stickerData.lAmount, ud.stickerData.cAmount)

	ud.stickerData.pos += 1

	if ud.stickerData.pos == ud.stickerData.lAmount {
		if createSet {
			if ud.command == "import" {
				insertLineS(ud.lineData.id, ud.lineData.link, ud.stickerData.id, ud.stickerData.title, false)
			}
			insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
		}
		c.Send("Success! /start")
		sendSFromSS(c)
		endSession(c)
	} else {
		sendAskEmojiAssign(c)
	}

	return nil
}

func commitSticker(createSet bool, flCount *int, safeMode bool, sf *StickerFile, c tele.Context, ss tele.StickerSet) error {
	var err error
	var floodErr tele.FloodError
	var f string
	ud := users.data[c.Sender().ID]

	sf.wg.Wait()
	if ud.stickerData.isVideo {
		if !safeMode {
			f = sf.cPath
		} else {
			f, _ = ffToWebmSafe(sf.oPath)
		}
		ss.WebM = &tele.File{FileLocal: f}
	} else {
		f = sf.cPath
		ss.PNG = &tele.File{FileLocal: f}
	}

	log.Debugln("sticker file path:", sf.cPath)
	log.Debugln("attempt commiting:", ss)
	log.Debugln("target emoji is:", ss.Emojis)
	// Retry loop.
	for i := 0; i < 3; i++ {
		if createSet {
			err = c.Bot().CreateStickerSet(c.Recipient(), ss)
		} else {
			err = c.Bot().AddSticker(c.Recipient(), ss)
		}
		if err == nil {
			break
		}
		log.Warnf("commit sticker error:%s for set:%s. creatSet?: %v", err, ss.Name, createSet)
		if errors.As(err, &floodErr) {
			*flCount += 1
			log.Warnln("Current flood limit count:", *flCount)
			if createSet || *flCount > 4 {
				sendTooManyFloodLimits(c)
				return errors.New("too many flood limits")
			}
			// Telegram's flood limit is strange.
			// It only happens to specific user at a specific time.
			// It is "fake" most of time, since TDLib in API Server will automatically retry.
			// However! API always return 429 without mentioning its self retry.
			// Since API side will always do retry at TDLib level, message_id was also being kept so
			// no position shift will happen.
			// Flood limit error could be probably ignored.
			sendFLWarning(c)
			log.Warnf("Flood limit encountered for user:%d for set:%s", c.Sender().ID, ss.Name)
			log.Warnln("commit sticker retry after: ", floodErr.RetryAfter)
			log.Warn("sleeping...zzz")
			if floodErr.RetryAfter > 60 {
				log.Error("RA too long! Telegram's bug?")
				log.Error("Attempt to sleep for 90 seconds.")
				time.Sleep(90 * time.Second)
			} else {
				// Sleep with some extra seconds due to bugs being reported.
				extraRA := 20 * *flCount
				time.Sleep(time.Duration((floodErr.RetryAfter + extraRA) * int(time.Second)))
			}

			log.Warn("woke up from RA sleep. ignoring this error.")
			break

		} else if strings.Contains(strings.ToLower(err.Error()), "video_long") {
			// Redo with safe mode on.
			// This should happen only one time.
			// So if safe mode is on and this error still occurs, return err.
			if safeMode {
				log.Error("safe mode DID NOT resolve video_long problem.")
				return err
			} else {
				log.Warnln("returned video_long, attempting safe mode.")
				return commitSticker(createSet, flCount, true, sf, c, ss)
			}
		} else if strings.Contains(err.Error(), "400") {
			// return remaining 400 BAD REQUEST to parent.
			return err
		} else {
			// Handle unknown error here.
			// We simply retry for 2 more times with 5 sec interval.
			if i == 2 {
				log.Warn("too many retries, end retry loop")
				return err
			}
			log.Warn("retrying...")
			time.Sleep(5 * time.Second)
		}
	}
	if safeMode {
		log.Warn("safe mode resolved video_long problem.")
	}
	return nil
}

func editStickerEmoji(newEmoji string, index int, fid string, f string, ssLen int, ud *UserData) error {
	c := ud.lastContext

	//this ss will only be used to commit sticker.
	ss := *ud.stickerData.stickerSet
	if ss.Video {
		ss.WebM = &tele.File{FileLocal: f}
	} else {
		ss.PNG = &tele.File{FileLocal: f}
	}
	ss.Emojis = newEmoji
	ss.Stickers = nil
	sf := &StickerFile{
		oPath: f,
		cPath: f,
	}
	flCount := 0
	err := commitSticker(false, &flCount, false, sf, c, ss)
	if err != nil {
		return errors.New("error commiting temp sticker " + err.Error())
	}

	for i := 0; i < 10; i++ {
		select {
		case <-ud.ctx.Done():
			log.Warn("editStickerEmoji received ctxDone!")
			return errors.New("user interrupted")
		default:
		}
		time.Sleep(2 * time.Second)
		ssNew, err := c.Bot().StickerSet(ud.stickerData.id)
		if err != nil {
			continue
		}
		log.Debugln(len(ssNew.Stickers))
		log.Debugln(ssLen)
		if len(ssNew.Stickers) != ssLen+1 {
			//Not committed to API server yet.
			continue
		}
		// commitedUID := ssNew.Stickers[len(ssNew.Stickers)-1].UniqueID
		commitedFID := ssNew.Stickers[len(ssNew.Stickers)-1].FileID
		if commitedFID == fid {
			log.Warn("FID duplacates, try again?")
			continue
		}

		log.Infoln("Setting position of:", commitedFID)
		err = c.Bot().SetStickerPosition(commitedFID, index)
		if err != nil {
			//Another API bug.
			//API returns a new file ID but refuses to use it.
			//Try really hard to make it work.
			log.Errorln("error setting position, retrying...", err)
			continue
		}
		//commit back the lastest set.
		ud.stickerData.stickerSet = ssNew
		return c.Bot().DeleteSticker(fid)
	}
	return errors.New("error setting position after editing emoji")
}

// Accept telebot Media and Sticker only
func appendMedia(c tele.Context) error {
	log.Debugf("Received file, MType:%s, FileID:%s", c.Message().Media().MediaType(), c.Message().Media().MediaFile().FileID)
	var files []string
	ud := users.data[c.Sender().ID]
	ud.wg.Add(1)
	defer ud.wg.Done()

	if (ud.stickerData.isVideo && ud.stickerData.cAmount+len(ud.stickerData.stickers) > 50) ||
		(ud.stickerData.cAmount+len(ud.stickerData.stickers) > 120) {
		return errors.New("sticker set already full 此貼圖包已滿")
	}

	workDir := users.data[c.Sender().ID].workDir
	savePath := filepath.Join(workDir, secHex(4))

	err := c.Bot().Download(c.Message().Media().MediaFile(), savePath)
	if err != nil {
		return errors.New("error downloading media")
	}
	if c.Message().Media().MediaType() == "document" && guessIsArchive(c.Message().Document.FileName) {
		files = append(files, archiveExtract(savePath)...)
	} else {
		files = append(files, savePath)
	}

	var sfs []*StickerFile
	for _, f := range files {
		var cf string
		var err error
		if ud.stickerData.isVideo {
			if c.Message().Sticker != nil && c.Message().Sticker.Video {
				cf = f
			} else {
				cf, err = ffToWebm(f)
			}
		} else {
			cf, err = imToWebp(f)
		}
		if err != nil {
			log.Warnln("Failed converting one user sticker", err)
			c.Send("Failed converting one user sticker:" + err.Error())
			continue
		}
		sfs = append(sfs, &StickerFile{
			oPath: f,
			cPath: cf,
		})
		log.Debugf("One received file OK. oPath:%s | cPath:%s", f, cf)
	}

	if len(sfs) == 0 {
		return errors.New("download or convert error")
	}

	ud.stickerData.stickers = append(ud.stickerData.stickers, sfs...)
	ud.stickerData.lAmount = len(ud.stickerData.stickers)
	replySFileOK(c, len(ud.stickerData.stickers))
	return nil
}

func guessIsArchive(f string) bool {
	f = strings.ToLower(f)
	archiveExts := []string{".rar", ".7z", ".zip", ".tar", ".gz", ".bz2", ".zst", ".rar5"}
	for _, ext := range archiveExts {
		if strings.HasSuffix(f, ext) {
			return true
		}
	}
	return false
}

func moveSticker(oldIndex int, newIndex int, ud *UserData) error {
	sid := ud.stickerData.stickerSet.Stickers[oldIndex].FileID
	return b.SetStickerPosition(sid, newIndex)
}
