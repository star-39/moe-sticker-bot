package main

import (
	"errors"
	"fmt"
	"os"
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

	sendProcessStarted(c, "")
	ud.wg.Wait()

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!")
		return errors.New("no sticker available")
	}

	log.Debugln("stickerData summary:")
	log.Debugln(ud.stickerData)
	committedStickers := 0
	errorCount := 0

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
		go editProgressMsg(index, len(ud.stickerData.stickers), "", c)
		if index == 0 && createSet {
			err = commitSticker(true, 1, false, sf, c, ss)
			if err != nil {
				log.Errorln("create failed. ", err)
				return err
			} else {
				committedStickers += 1
			}
		} else {
			err = commitSticker(false, committedStickers+1, false, sf, c, ss)
			if err != nil {
				log.Warnln("a sticker failed to add. ", err)
				c.Send("one sticker failed to add, index is:" + strconv.Itoa(index))
				errorCount += 1
				if errorCount > 2 {
					return errors.New("too many errors when adding sticker")
				}
			} else {
				committedStickers += 1
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
	editProgressMsg(0, 0, "Success! /start", c)
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
		err = commitSticker(true, 1, false, sf, c, ss)
		if err != nil {
			log.Errorln("create failed. ", err)
			return err
		} else {
			ud.stickerData.cAmount += 1
		}
	} else {
		err = commitSticker(false, ud.stickerData.cAmount+1, false, sf, c, ss)
		if err != nil {
			if strings.Contains(err.Error(), "invalid sticker emojis") {
				return c.Send("Bad emoji. try again.\n這個emoji無效, 請再試一次.")
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
		terminateSession(c)
	} else {
		sendAskEmojiAssign(c)
	}

	return nil
}

// This function handles sticker conversion and upload.
// The "amountSupposed" is for detecting fake flood limit.
func commitSticker(createSet bool, amountSupposed int, safeMode bool, sf *StickerFile, c tele.Context, ss tele.StickerSet) error {
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
			// This Error is NASTY.
			// It only happens to specific user at specific time.
			// It is "fake" most of time, since TDLib in API Server will automatically retry.
			// However! API always return 429 without mentioning its self retry.
			// As a workaround, we need to verify whether this error is "genuine".
			// This leads to another problem, API sometimes return the sticker set before self retry being made,
			// or the result was being cached in API.
			// We need to wait long enough to verify the actual result.
			//
			// EDIT: No! Seems API side will always do retry at TDLib level, message_id was also being kept so
			// no position shifting will happen.
			// Yep, we are gonna ignore the FLOOD_LIMIT!
			c.Send("We encountered a small issue and might take some time (< 1min) to resolve, please wait...\n" +
				"BOT遇到了點小問題, 可能需要一點時間(少於1分鐘)解決, 請耐心等待...")
			log.Warnf("Flood limit encountered by set:%s", ss.Name)
			log.Warnln("commit sticker retry after: ", floodErr.RetryAfter)
			log.Warn("sleeping...zzz")
			if floodErr.RetryAfter > 60 {
				log.Error("RA too crazy! should be framework bug.")
				log.Error("Attempt to sleep for 65 seconds.")
				time.Sleep(65 * time.Second)
			} else {
				// Sleep for extra 5 seconds than RA.
				time.Sleep(time.Duration((floodErr.RetryAfter + 10) * int(time.Second)))
			}

			log.Warn("woke up from RA sleep. ignoring this error.")
			break
			// do this check AFTER sleep.
			// if verifyRetryAfterIsFake(amountSupposed, c, ss) {
			// 	log.Warn("The RA is fake, breaking retry loop...")
			// 	// Break retry loop if RA is fake.
			// 	break
			// } else {
			// 	log.Warn("Oops! The flood limit is real, retrying...")
			// 	continue
			// }
		} else if strings.Contains(strings.ToLower(err.Error()), "video_long") {
			// Redo with safe mode on.
			// This should happen only one time.
			// So if safe mode is on and this error still occurs, return err.
			if safeMode {
				log.Error("safe mode DID NOT resolve video_long problem.")
				return err
			} else {
				log.Warnln("returned video_long, attempting safe mode.")
				return commitSticker(createSet, amountSupposed, true, sf, c, ss)
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

// Completely useless!
// API server caches the SS result and always return a outdated value!
// They also silently do retry at TDLib level after getting "can_retry" from TG.
// Goodbye! Duplicated stickers!
// func verifyRetryAfterIsFake(amountSupposed int, c tele.Context, ss tele.StickerSet) bool {
// 	var isFake bool
// 	// go crazy! let's check it FIVE TIMES!
// 	// How dare you https://github.com/tdlib/telegram-bot-api
// 	for i := 0; i < 5; i++ {
// 		time.Sleep(5 * time.Second)
// 		log.Warnln("Check RA... loop:", i)
// 		cloudSS, err := c.Bot().StickerSet(ss.Name)
// 		// if RA is fake, return immediately! so we can continue operation.
// 		if amountSupposed == 1 {
// 			if err != nil {
// 				// Sticker set exists.
// 				return true
// 			} else {
// 				isFake = false
// 			}
// 		} else {
// 			log.Warnln("Checked cAmount is :", len(cloudSS.Stickers))
// 			log.Warnln("We suppose :", amountSupposed)
// 			if len(cloudSS.Stickers) == amountSupposed {
// 				return true
// 			} else {
// 				isFake = false
// 			}
// 		}
// 	}
// 	return isFake
// }

func downloadSAndC(path string, s *tele.Sticker, c tele.Context) (string, string) {
	var f string
	if s.Video {
		f = path + ".webm"
		c.Bot().Download(&s.File, f)
		cf, _ := ffToGif(f)
		return f, cf
	} else if s.Animated {
		f = path + ".tgs"
		c.Bot().Download(&s.File, f)
		return f, ""
	} else {
		f = path + ".webp"
		c.Bot().Download(&s.File, f)
		cf, _ := imToPng(f)
		return f, cf
	}
}

func downloadStickersToZip(s *tele.Sticker, wantSet bool, c tele.Context) error {
	id := s.SetName
	ud := users.data[c.Sender().ID]
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	workDir := filepath.Join(ud.userDir, id)
	os.MkdirAll(workDir, 0755)
	var flist []string
	var cflist []string
	var err error

	if !wantSet {
		_, cf := downloadSAndC(filepath.Join(workDir, id+"_"+s.Emoji), s, c)
		log.Debugln("downloading:", cf)
		if s.Video {
			zip := filepath.Join(workDir, secHex(4)+".zip")
			fCompress(zip, []string{cf})
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})
		} else {
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(cf), File: tele.FromDisk(cf)})
		}
		return err
	}

	ss, _ := c.Bot().StickerSet(id)
	ud.stickerData.id = ss.Name
	ud.stickerData.title = ss.Title
	ud.stickerData.link = "https://t.me/addstickers/" + ss.Name
	sendProcessStarted(c, "")
	for index, s := range ss.Stickers {
		select {
		case <-ud.ctx.Done():
			log.Warn("downloadStickersToZip received ctxDone!")
			return nil
		default:
		}
		go editProgressMsg(index, len(ss.Stickers), "", c)
		fName := filepath.Join(workDir, fmt.Sprintf("%s_%d_%s", id, index+1, s.Emoji))
		f, cf := downloadSAndC(fName, &s, c)
		flist = append(flist, f)
		if cf != "" {
			cflist = append(cflist, cf)
		}
		log.Debugln("Download one sticker OK, path: ", f)
	}
	go editProgressMsg(0, 0, "Uploading...", c)

	webmZipPath := filepath.Join(workDir, id+"_webm.zip")
	webpZipPath := filepath.Join(workDir, id+"_webp.zip")
	pngZipPath := filepath.Join(workDir, id+"_png.zip")
	gifZipPath := filepath.Join(workDir, id+"_gif.zip")
	tgsZipPath := filepath.Join(workDir, id+"_tgs.zip")

	var zipPaths []string

	if ss.Video {
		zipPaths = append(zipPaths, fCompressVol(webmZipPath, flist)...)
		zipPaths = append(zipPaths, fCompressVol(gifZipPath, cflist)...)
	} else if ss.Animated {
		zipPaths = append(zipPaths, fCompressVol(tgsZipPath, flist)...)
	} else {
		zipPaths = append(zipPaths, fCompressVol(webpZipPath, flist)...)
		zipPaths = append(zipPaths, fCompressVol(pngZipPath, cflist)...)
	}
	for _, zipPath := range zipPaths {
		select {
		case <-ud.ctx.Done():
			log.Warn("downloadStickersToZip received ctxDone!")
			return nil
		default:
		}
		_, err := c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zipPath), File: tele.FromDisk(zipPath)})
		time.Sleep(1 * time.Second)
		if err != nil {
			return err
		}
	}

	editProgressMsg(0, 0, "success! /start", c)
	return nil
}

func downloadGifToZip(c tele.Context) error {
	workDir := filepath.Join(users.data[c.Sender().ID].userDir, secHex(4))
	os.MkdirAll(workDir, 0755)
	f := filepath.Join(workDir, "gif.mp4")
	err := c.Bot().Download(&c.Message().Animation.File, f)
	cf, _ := ffToGif(f)
	zip := secHex(4) + ".zip"
	fCompress(zip, []string{cf})

	c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})

	return err
}

func appendMedia(c tele.Context) error {
	var files []string
	ud := users.data[c.Sender().ID]
	ud.wg.Add(1)
	workDir := users.data[c.Sender().ID].userDir
	savePath := filepath.Join(workDir, secHex(4))
	if c.Message().Document != nil {
		c.Bot().Download(&c.Message().Document.File, savePath)
		fName := c.Message().Document.FileName
		if guessIsArchive(strings.ToLower(fName)) {
			files = append(files, archiveExtract(savePath)...)
		} else {
			files = append(files, savePath)
		}
	} else if c.Message().Photo != nil {
		c.Bot().Download(&c.Message().Photo.File, savePath)
		files = append(files, savePath)
	} else if c.Message().Video != nil {
		c.Bot().Download(&c.Message().Video.File, savePath)
		files = append(files, savePath)
	} else if c.Message().Sticker != nil {
		c.Bot().Download(&c.Message().Sticker.File, savePath)
		files = append(files, savePath)
	} else {
		log.Debug("?unknown media.")
	}

	var sfs []*StickerFile
	for _, f := range files {
		var cf string
		var err error
		if ud.stickerData.isVideo {
			cf, err = ffToWebm(f)
		} else {
			cf, err = imToWebp(f)
		}
		if err != nil {
			log.Warnln("Failed converting one user sticker", err)
			continue
		}
		sfs = append(sfs, &StickerFile{
			oPath: f,
			cPath: cf,
		})
	}
	ud.wg.Done()
	if len(sfs) == 0 {
		return errors.New("download or convert error")
	}

	ud.stickerData.stickers = append(ud.stickerData.stickers, sfs...)
	return nil
}

func guessIsArchive(f string) bool {
	archiveExts := []string{".rar", ".7z", ".zip", ".tar", ".gz", ".bz2", ".zst", ".rar5"}
	for _, ext := range archiveExts {
		if strings.HasSuffix(f, ext) {
			return true
		}
	}
	return false
}
