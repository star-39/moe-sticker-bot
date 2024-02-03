package core

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

// Final stage of automated sticker submission.
func submitStickerSetAuto(createSet bool, c tele.Context) error {
	uid := c.Sender().ID
	ud := users.data[uid]
	pText, teleMsg, _ := sendProcessStarted(ud, c, "Waiting...")
	ud.wg.Wait()

	// cache ud and clean, allow user to be release from session.
	// lock is ignored here.
	cud := *users.data[uid]
	ud = &cud
	cleanUserData(c.Sender().ID)
	sendNotifyWorkingOnBackground(c)

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!")
		return errors.New("no sticker available")
	}

	// wait for all channels to be done.
	// cache the list.
	list := autocommitWorkersList[uid]
	// append a new channel to original list
	done := make(chan bool)
	autocommitWorkersList[uid] = append(autocommitWorkersList[uid], done)
	defer close(done)

	for _, c := range list {
		select {
		case _, ok := <-c:
			if !ok {
				//channel already closed
				log.Debug("channel already closed")
				continue
			}
		//Should never be triggered
		case <-time.After(30 * time.Minute):
			log.Warn("Channel wait reached 10m timeout!")
		}
	}

	log.Debugln("stickerData summary:")
	log.Debugln(ud.stickerData)
	committedStickers := 0
	errorCount := 0
	flCount := &ud.stickerData.flCount

	ss := tele.StickerSet{
		Name:  ud.stickerData.id,
		Title: ud.stickerData.title,
		Video: ud.stickerData.isVideo,
		Type:  ud.stickerData.stickerSetType,
	}

	for index, sf := range ud.stickerData.stickers {
		var err error
		go editProgressMsg(index, len(ud.stickerData.stickers), "", pText, teleMsg, c)

		inputSticker := tele.InputSticker{
			Emojis:   ud.stickerData.emojis,
			Keywords: []string{"sticker"},
		}
		if index == 0 && createSet {
			err = commitSticker(true, index, flCount, false, sf, c, inputSticker, ss)
			if err != nil {
				log.Errorln("create sticker set failed!. ", err)
				return err
			} else {
				committedStickers += 1
			}
		} else {
			err = commitSticker(false, index, flCount, false, sf, c, inputSticker, ss)
			if err != nil {
				log.Warnln("execAutoCommit: a sticker failed to add. ", err)
				sendOneStickerFailedToAdd(c, index, err)
				errorCount += 1
			} else {
				log.Debugln("one sticker commited. count: ", committedStickers)
				committedStickers += 1
			}
			// If encountered flood limit more than once, set a interval.
			if *flCount == 1 {
				sleepTime := 10 + rand.Intn(10)
				time.Sleep(time.Duration(sleepTime) * time.Second)
			} else if *flCount > 1 {
				sleepTime := 60 + rand.Intn(10)
				time.Sleep(time.Duration(sleepTime) * time.Second)
			}
		}
		// Tolerate at most 3 errors when importing sticker set.
		if ud.command == "import" && errorCount > 3 {
			return errors.New("too many errors importing")
		}
	}
	if createSet {
		if ud.command == "import" {
			insertLineS(ud.lineData.Id, ud.lineData.Link, ud.stickerData.id, ud.stickerData.title, true)
			// Only verify for import.
			// User generated sticker set might intentionally contain same stickers.
			if *flCount > 1 {
				verifyFloodedStickerSet(c, *flCount, errorCount, ud.lineData.Amount, ud.stickerData.id)
			}
		}
		insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
	}
	editProgressMsg(0, 0, "Success! /start", pText, teleMsg, c)
	sendSFromSS(c, ud.stickerData.id, teleMsg)
	return nil
}

// Only fatal error should be returned.
func submitStickerManual(createSet bool, pos int, emojis []string, keywords []string, c tele.Context) error {
	ud := users.data[c.Sender().ID]

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!!")
		return errors.New("no sticker available")
	}
	var err error
	ss := tele.StickerSet{
		Name:  ud.stickerData.id,
		Title: ud.stickerData.title,
		Video: ud.stickerData.isVideo,
		Type:  ud.stickerData.stickerSetType,
	}

	log.Debugln("execEmojiAssign: sticker summary: ", ss)
	log.Debugf("execEmojiAssign: attempting to commit: pos:%d, lAmount:%d, cAmount:%d", pos, ud.stickerData.lAmount, ud.stickerData.cAmount)

	sf := ud.stickerData.stickers[pos]
	input := tele.InputSticker{
		Emojis:   emojis,
		Keywords: keywords,
	}

	//Do not submit to goroutine when creating sticker set.
	if createSet && pos == 0 {
		defer close(ud.commitChans[pos])
		err = commitSticker(true, pos, &ud.stickerData.flCount, false, sf, c, input, ss)
		if err != nil {
			log.Errorln("create failed. ", err)
			return err
		} else {
			ud.stickerData.cAmount += 1
		}
		if ud.stickerData.lAmount == 1 {
			return finalizeSubmitStickerManual(c, createSet, ud)
		}
	} else {
		go func() {
			//wait for the previous commit to be done.
			if pos > 0 {
				<-ud.commitChans[pos-1]
			}

			err = commitSticker(false, pos, &ud.stickerData.flCount, false, sf, c, input, ss)
			if err != nil {
				sendOneStickerFailedToAdd(c, pos, err)
				log.Warnln("execEmojiAssign: a sticker failed to add: ", err)
			} else {
				ud.stickerData.cAmount += 1
			}

			if pos+1 == ud.stickerData.lAmount {
				finalizeSubmitStickerManual(c, createSet, ud)
			}
			close(ud.commitChans[pos])
		}()
	}
	return nil
}

func finalizeSubmitStickerManual(c tele.Context, createSet bool, ud *UserData) error {
	if createSet {
		if ud.command == "import" {
			insertLineS(ud.lineData.Id, ud.lineData.Link, ud.stickerData.id, ud.stickerData.title, false)
		}
		insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
	}
	sendExecEmojiAssignFinished(c)
	// c.Send("Success! /start")
	sendSFromSS(c, ud.stickerData.id, nil)
	endSession(c)
	return nil
}

// Commit single sticker, retry happens inside this function.
// If all retries failed, return err.
//
// flCount counts the total flood limit for entire sticker set.
// pos is for logging only.
func commitSticker(createSet bool, pos int, flCount *int, safeMode bool, sf *StickerFile, c tele.Context, input tele.InputSticker, ss tele.StickerSet) error {
	var err error
	var floodErr tele.FloodError
	var sFormat string
	var file tele.File

	sf.wg.Wait()

	if ss.Video {
		sFormat = "video"
		if safeMode {
			f, _ := msbimport.FFToWebmSafe(sf.oPath, msbimport.FORMAT_TG_REGULAR_ANIMATED)
			file = tele.File{FileLocal: f}
		} else {
			file = tele.File{FileLocal: sf.cPath}
		}
	} else {
		sFormat = "static"
		file = tele.File{FileLocal: sf.cPath}
	}

	log.Debugln("sticker file path:", sf.cPath)
	log.Debugln("attempt commiting:", ss)
	// Retry loop.
	// For each sticker, retry at most 2 times, means 3 commit attempts in total.
	for i := 0; i < 3; i++ {
		uploadedFile, err := c.Bot().UploadSticker(c.Recipient(), sFormat, &file)
		if err != nil {
			log.Errorln("commitSticker: error on UploadSticker")
			//jump to error handling without furthur action.
			goto HANDLE_ERROR
		}
		input.Sticker = uploadedFile.FileID
		if createSet {
			err = c.Bot().CreateStickerSet(c.Recipient(), sFormat, []tele.InputSticker{input}, ss)
		} else {
			err = c.Bot().AddSticker(c.Recipient(), input, ss)
		}
		if err == nil {
			return nil
		}

	HANDLE_ERROR:
		log.Errorf("commit sticker error:%s for set:%s. creatSet?: %v", err, ss.Name, createSet)
		// Is flood limit error.
		// Telegram's flood limit is strange.
		// It only happens to a specific user at a specific time.
		// It is "fake" most of time, since TDLib in API Server will automatically retry.
		// However! API always return 429 without mentioning its self retry.
		// Since API side will always do retry at TDLib level, message_id was also being kept so
		// no position shift will happen.
		// Flood limit error could be probably ignored.
		if errors.As(err, &floodErr) {
			// This reflects the retry count for entire SS.
			*flCount += 1
			log.Warnf("commitSticker: Flood limit encountered for user:%d, set:%s, count:%d, pos:%d", c.Sender().ID, ss.Name, *flCount, pos)
			log.Warnln("commitSticker: commit sticker retry after: ", floodErr.RetryAfter)
			// If flood limit encountered when creating set, return immediately.
			if createSet {
				sendTooManyFloodLimits(c)
				return errors.New("flood limit when creating set")
			}
			if *flCount == 2 {
				sendFLWarning(c)
			}

			//Sleep
			if floodErr.RetryAfter > 60 {
				log.Error("RA too long! Telegram's bug? Attempt to sleep for 120 seconds.")
				time.Sleep(120 * time.Second)
			} else {
				extraRA := *flCount * 15
				log.Warnf("Sleeping for %d seconds due to FL.", floodErr.RetryAfter+extraRA)
				time.Sleep(time.Duration(floodErr.RetryAfter+extraRA) * time.Second)
			}

			log.Warnf("Woken up from RA sleep. ignoring this error. user:%d, set:%s, count:%d, pos:%d", c.Sender().ID, ss.Name, *flCount, pos)

			//According to collected logs, exceeding 2 flood counts will sometimes cause api server to stop auto retrying.
			//Hence, we do retry here, else, break retry loop.
			if *flCount > 2 {
				continue
			} else {
				break
			}

		} else if strings.Contains(strings.ToLower(err.Error()), "video_long") {
			// Redo with safe mode on.
			// This should happen only one time.
			// So if safe mode is on and this error still occurs, return err.
			if safeMode {
				log.Error("safe mode DID NOT resolve video_long problem.")
				return err
			} else {
				log.Warnln("returned video_long, attempting safe mode.")
				return commitSticker(createSet, pos, flCount, true, sf, c, input, ss)
			}
		} else if strings.Contains(err.Error(), "400") {
			// return remaining 400 BAD REQUEST immediately to parent without retry.
			return err
		} else if strings.Contains(err.Error(), "invalid sticker emojis") {
			log.Warn("commitSticker: invalid emoji, resetting to a star emoji and retrying...")
			input.Emojis = []string{"⭐️"}
		} else {
			// Handle unknown error here.
			// We simply retry for 2 more times with 5 sec interval.
			log.Warnln("commitSticker: retrying... cause:", err)
			time.Sleep(5 * time.Second)
		}
	}

	log.Warn("commitSticker: too many retries")
	if errors.As(err, &floodErr) {
		log.Warn("commitSticker: reached max retry for flood limit, assume success.")
		return nil
	}
	return err
}

func editStickerEmoji(newEmojis []string, fid string, ud *UserData) error {
	return b.SetStickerEmojiList(ud.lastContext.Recipient(), fid, newEmojis)
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

	err := teleDownload(c.Message().Media().MediaFile(), savePath)
	if err != nil {
		return errors.New("error downloading media")
	}
	if c.Message().Media().MediaType() == "document" && guessIsArchive(c.Message().Document.FileName) {
		files = append(files, msbimport.ArchiveExtract(savePath)...)
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
			} else if c.Message().Sticker != nil && c.Message().Sticker.Animated {
				return errors.New("appendMedia: TGS to Video sticker not supported, try another one")
			} else {
				cf, err = msbimport.FFToWebmTGVideo(f, msbimport.FORMAT_TG_REGULAR_ANIMATED)
			}
		} else {
			cf, err = msbimport.IMToWebpTGStatic(f, msbimport.FORMAT_TG_REGULAR_STATIC)
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

func verifyFloodedStickerSet(c tele.Context, fc int, ec int, desiredAmount int, ssn string) {
	time.Sleep(31 * time.Second)
	ss, err := b.StickerSet(ssn)
	if err != nil {
		return
	}
	if desiredAmount < len(ss.Stickers) {
		log.Warnf("A flooded sticker set duplicated! floodCount:%d, errorCount:%d, ssn:%s, desired:%d, got:%d", fc, ec, ssn, desiredAmount, len(ss.Stickers))
		log.Warnf("Attempting dedup!")
		workdir := filepath.Join(dataDir, secHex(8))
		os.MkdirAll(workdir, 0755)
		for si, s := range ss.Stickers {
			if si > 0 {
				fp := filepath.Join(workdir, strconv.Itoa(si-1)+".webp")
				f := filepath.Join(workdir, strconv.Itoa(si)+".webp")
				teleDownload(&s.File, f)

				if compCRC32(f, fp) {
					b.DeleteSticker(s.FileID)
				}
			}
		}
		os.RemoveAll(workdir)
	} else if desiredAmount > len(ss.Stickers) {
		log.Warnf("A flooded sticker set missing sticker! floodCount:%d, errorCount:%d, ssn:%s, desired:%d, got:%d", fc, ec, ssn, desiredAmount, len(ss.Stickers))
		c.Reply("Sorry, this sticker set seems corrupted, please check.\n抱歉, 這個貼圖包似乎有缺失貼圖, 請檢查一下.")
	} else {
		log.Infof("A flooded sticker set seems ok. floodCount:%d, errorCount:%d, ssn:%s, desired:%d, got:%d", fc, ec, ssn, desiredAmount, len(ss.Stickers))
	}

}
