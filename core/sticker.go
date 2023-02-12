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
	"github.com/star-39/moe-sticker-bot/pkg/convert"
	"github.com/star-39/moe-sticker-bot/pkg/util"
	tele "gopkg.in/telebot.v3"
)

// Final stage of automated sticker submission.
func execAutoCommit(createSet bool, c tele.Context) error {
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
	defer func() {
		// done <- true
		close(done)
	}()
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
	flCount := 0

	for index, sf := range ud.stickerData.stickers {
		// select {
		// case <-ud.ctx.Done():
		// 	log.Warn("execAutoCommit received ctxDone!")
		// 	return nil
		// default:
		// }
		var err error
		ss := tele.StickerSet{
			Name:   ud.stickerData.id,
			Title:  ud.stickerData.title,
			Emojis: ud.stickerData.emojis[0],
			Video:  ud.stickerData.isVideo,
		}
		go editProgressMsg(index, len(ud.stickerData.stickers), "", pText, teleMsg, c)
		if index == 0 && createSet {
			err = commitSticker(true, index, &flCount, false, sf, c, ss)
			if err != nil {
				log.Errorln("create failed!. ", err)
				return err
			} else {
				committedStickers += 1
			}
		} else {
			err = commitSticker(false, index, &flCount, false, sf, c, ss)
			if err != nil {
				log.Warnln("execAutoCommit: a sticker failed to add. ", err)
				sendOneStickerFailedToAdd(c, index, err)
				errorCount += 1
			} else {
				log.Debugln("one sticker commited. count: ", committedStickers)
				committedStickers += 1
			}
			// If encountered flood limit more than once, set a interval.
			if flCount == 1 {
				sleepTime := 10 + rand.Intn(10)
				time.Sleep(time.Duration(sleepTime) * time.Second)
			} else if flCount > 1 {
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
			if flCount > 1 {
				verifyFloodedStickerSet(c, flCount, errorCount, ud.lineData.Amount, ud.stickerData.id)
			}
		}
		insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
	}
	editProgressMsg(0, 0, "Success! /start", pText, teleMsg, c)
	sendSFromSS(c, ud.stickerData.id, teleMsg)
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
		Video:  ud.stickerData.isVideo,
	}

	sf := ud.stickerData.stickers[ud.stickerData.pos]
	log.Debugln("ss summary:")
	log.Debugln(ss)

	if createSet && ud.stickerData.pos == 0 {
		err = commitSticker(true, ud.stickerData.pos, new(int), false, sf, c, ss)
		if err != nil {
			log.Errorln("create failed. ", err)
			return err
		} else {
			ud.stickerData.cAmount += 1
		}
	} else {
		err = commitSticker(false, ud.stickerData.pos, new(int), false, sf, c, ss)
		if err != nil {
			if strings.Contains(err.Error(), "invalid sticker emojis") {
				return c.Send("Sorry, this emoji is invalid. Try another one.\n抱歉, 這個emoji無效, 請另試一次.")
			}
			sendOneStickerFailedToAdd(c, ud.stickerData.pos, err)
			log.Warnln("execEmojiAssign: a sticker failed to add. ", err)
		} else {
			ud.stickerData.cAmount += 1
		}
	}

	log.Debugf("execEmojiAssign: one sticker commit attempted. pos:%d, lAmount:%d, cAmount:%d", ud.stickerData.pos, ud.stickerData.lAmount, ud.stickerData.cAmount)

	ud.stickerData.pos += 1

	if ud.stickerData.pos == ud.stickerData.lAmount {
		if createSet {
			if ud.command == "import" {
				insertLineS(ud.lineData.Id, ud.lineData.Link, ud.stickerData.id, ud.stickerData.title, false)
			}
			insertUserS(c.Sender().ID, ud.stickerData.id, ud.stickerData.title, time.Now().Unix())
		}
		c.Send("Success! /start")
		sendSFromSS(c, ud.stickerData.id, nil)
		endSession(c)
	} else {
		sendAskEmojiAssign(c)
	}

	return nil
}

// Commit single sticker, retry happens inside this function.
// If all retries failed, return err.
//
// ss contains metadata for the single sticker.
// it feels weird but it's the framework's way to do so.
// therefore, Video? must be set.
//
// flCount counts the total flood limit for entire sticker set.
// pos is for logging only.
func commitSticker(createSet bool, pos int, flCount *int, safeMode bool, sf *StickerFile, c tele.Context, ss tele.StickerSet) error {
	var err error
	var floodErr tele.FloodError
	var f string

	sf.wg.Wait()
	if ss.Video {
		if !safeMode {
			f = sf.cPath
		} else {
			f, _ = convert.FFToWebmSafe(sf.oPath)
		}
		ss.WebM = &tele.File{FileLocal: f}
	} else {
		f = sf.cPath
		ss.PNG = &tele.File{FileLocal: f}
	}

	log.Debugln("sticker file path:", sf.cPath)
	log.Debugln("attempt commiting:", ss)
	// Retry loop.
	// For each sticker, retry at most 2 times, means 3 commit attempts in total.
	for i := 0; i < 3; i++ {
		if createSet {
			err = c.Bot().CreateStickerSet(c.Recipient(), ss)
		} else {
			err = c.Bot().AddSticker(c.Recipient(), ss)
		}
		if err == nil {
			break
		}
		//Deal with error below
		log.Errorf("commit sticker error:%s for set:%s. creatSet?: %v", err, ss.Name, createSet)
		if i == 2 {
			log.Warn("too many retries, end retry loop")
			return err
		}
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
				extraRA := *flCount * 30
				log.Warnf("Sleeping for %d seconds due to FL.", floodErr.RetryAfter+extraRA)
				time.Sleep(time.Duration(floodErr.RetryAfter+extraRA) * time.Second)
			}

			log.Warnf("Woken up from RA sleep. ignoring this error. user:%d, set:%s, count:%d, pos:%d", c.Sender().ID, ss.Name, *flCount, pos)

			//According to collected logs, exceeding 2 flood counts will cause api server to stop auto retrying.
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
				return commitSticker(createSet, pos, flCount, true, sf, c, ss)
			}
		} else if strings.Contains(err.Error(), "400") {
			// return remaining 400 BAD REQUEST immediately to parent without retry.
			return err
		} else {
			// Handle unknown error here.
			// We simply retry for 2 more times with 5 sec interval.
			log.Warnln("commitSticker: retrying... cause:", err)
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
	err := commitSticker(false, -1, &flCount, false, sf, c, ss)
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
		commitedFID := ssNew.Stickers[len(ssNew.Stickers)-1].FileID
		if commitedFID == fid {
			log.Warn("FID duplicated, try again?")
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

		time.Sleep(1 * time.Second)
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

	err := teleDownload(c.Message().Media().MediaFile(), savePath)
	if err != nil {
		return errors.New("error downloading media")
	}
	if c.Message().Media().MediaType() == "document" && guessIsArchive(c.Message().Document.FileName) {
		files = append(files, util.ArchiveExtract(savePath)...)
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
				cf, err = convert.FFToWebm(f)
			}
		} else {
			cf, err = convert.IMToWebp(f)
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
