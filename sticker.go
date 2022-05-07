package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func execAutoCommit(createSet bool, c tele.Context) error {
	ud := users.data[c.Sender().ID]
	sendProcessStarted(c, "")
	ud.wg.Wait()

	if len(ud.stickerData.stickers) == 0 {
		log.Error("No sticker to commit!!")
		return errors.New("no sticker available")
	}

	log.Debugln("stickerData summary:")
	log.Debugln(ud.stickerData)
	committedStickers := 0
	errorCount := 0

	for index, sf := range ud.stickerData.stickers {
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
	// ss := tele.StickerSet{
	// 	Name:   ud.stickerData.id,
	// 	Title:  ud.stickerData.title,
	// 	Emojis: ud.stickerData.emojis[0],
	// }

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
	for i := 0; i < 5; i++ {
		if createSet {
			err = c.Bot().CreateStickerSet(c.Recipient(), ss)
		} else {
			err = c.Bot().AddSticker(c.Recipient(), ss)
		}
		if err == nil {
			break
		}

		if errors.As(err, &floodErr) {

			log.Warnln("upload sticker retry after: ", floodErr.RetryAfter)
			if floodErr.RetryAfter > 60 {
				log.Error("RA too crazy! die now.")
				return err
			}
			log.Warn("sleeping...zzz")
			// Sleep for extra 5 seconds.
			time.Sleep(time.Duration((floodErr.RetryAfter + 5) * int(time.Second)))
			log.Warn("woke up from RA sleep.")
			// do this check AFTER sleep.
			if verifyRetryAfterIsFake(amountSupposed, c, ss) {
				log.Warn("The RA is fake, breaking retry loop...")
				// Break retry loop if RA is fake.
				break
			} else {
				log.Warn("Oops! The flood limit is real, retrying...")
				continue
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
				return commitSticker(createSet, amountSupposed, true, sf, c, ss)
			}
		} else {
			log.Warnln("upload sticker error:", err)
			return err
		}
	}
	if safeMode {
		log.Warn("safe mode resolved video_long problem.")
	}
	return nil
}

func verifyRetryAfterIsFake(amountSupposed int, c tele.Context, ss tele.StickerSet) bool {
	var cloudSS *tele.StickerSet
	var cloudAmount int
	var err error
	var isFake bool
	for i := 0; i < 2; i++ {
		cloudSS, err = c.Bot().StickerSet(ss.Name)
		if amountSupposed == 1 {
			if err != nil {
				// Sticker set exists.
				isFake = true
			} else {
				isFake = false
			}
		} else {
			cloudAmount = len(cloudSS.Stickers)
			if cloudAmount == amountSupposed {
				isFake = true
			} else {
				isFake = false
			}
		}
		time.Sleep(3 * time.Second)
	}
	return isFake
}

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
		go editProgressMsg(index, len(ss.Stickers), "", c)
		fName := filepath.Join(workDir, fmt.Sprintf("%s_%d_%s", id, index+1, s.Emoji))
		f, cf := downloadSAndC(fName, &s, c)
		flist = append(flist, f)
		if cf != "" {
			cflist = append(cflist, cf)
		}
		log.Debugln("Download one sticker OK, path: ", f)
	}
	go editProgressMsg(0, 0, "Packaging...", c)

	webmZipPath := filepath.Join(workDir, id+"_webm.zip")
	webpZipPath := filepath.Join(workDir, id+"_webp.zip")
	pngZipPath := filepath.Join(workDir, id+"_png.zip")
	gifZipPath := filepath.Join(workDir, id+"_gif.zip")
	tgsZipPath := filepath.Join(workDir, id+"_tgs.zip")

	if ss.Video {
		fCompress(webmZipPath, flist)
		fCompress(gifZipPath, cflist)
		_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(webmZipPath), File: tele.FromDisk(webmZipPath)})
		_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(gifZipPath), File: tele.FromDisk(gifZipPath)})
	} else if ss.Animated {
		fCompress(tgsZipPath, flist)
		_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(tgsZipPath), File: tele.FromDisk(tgsZipPath)})
	} else {
		fCompress(webpZipPath, flist)
		fCompress(pngZipPath, cflist)
		_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(webpZipPath), File: tele.FromDisk(webpZipPath)})
		_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(pngZipPath), File: tele.FromDisk(pngZipPath)})
	}
	if err != nil {
		return err
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
