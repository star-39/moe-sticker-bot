package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

func downloadSAndC(path string, s *tele.Sticker, needConvert bool, shrinkGif bool, c tele.Context) (string, string) {
	var f string
	var cf string
	var err error
	if s.Video {
		f = path + ".webm"
		err = c.Bot().Download(&s.File, f)
		if needConvert {
			if shrinkGif {
				cf, _ = ffToGifShrink(f)
			} else {
				cf, _ = ffToGif(f)
			}
		}
	} else if s.Animated {
		f = path + ".tgs"
		err = c.Bot().Download(&s.File, f)
	} else {
		f = path + ".webp"
		err = c.Bot().Download(&s.File, f)
		if needConvert {
			cf, _ = imToPng(f)
		}
	}
	if err != nil {
		return "", ""
	}
	return f, cf
}

func downloadStickersToZip(s *tele.Sticker, wantSet bool, c tele.Context) error {
	cache := *s
	s = &cache
	id := s.SetName
	ud := users.data[c.Sender().ID]
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	workDir := filepath.Join(ud.workDir, id)
	os.MkdirAll(workDir, 0755)
	var flist []string
	var cflist []string
	var err error

	if !wantSet {
		_, cf := downloadSAndC(filepath.Join(workDir, id+"_"+s.Emoji), s, true, false, c)
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
	pText, teleMsg, _ := sendProcessStarted(c, "")
	sendNotifySDOnBackground(c)
	cleanUserData(c.Sender().ID)
	for index, s := range ss.Stickers {
		go editProgressMsg(index, len(ss.Stickers), "", pText, teleMsg, c)
		fName := filepath.Join(workDir, fmt.Sprintf("%s_%d_%s", id, index+1, s.Emoji))
		f, cf := downloadSAndC(fName, &s, true, true, c)
		if f == "" || cf == "" {
			return errors.New("sticker download failed")
		}
		flist = append(flist, f)
		cflist = append(cflist, cf)

		log.Debugf("Download one sticker OK, path:%s cPath:%s", f, cf)
	}
	go editProgressMsg(0, 0, "Uploading...", pText, teleMsg, c)

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
		_, err := c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zipPath), File: tele.FromDisk(zipPath)})
		time.Sleep(1 * time.Second)
		if err != nil {
			return err
		}
	}

	editProgressMsg(0, 0, "success! /start", pText, teleMsg, c)
	return nil
}

func downloadGifToZip(c tele.Context) error {
	workDir := filepath.Join(users.data[c.Sender().ID].workDir, secHex(4))
	os.MkdirAll(workDir, 0755)
	f := filepath.Join(workDir, "gif.mp4")
	err := c.Bot().Download(&c.Message().Animation.File, f)
	cf, _ := ffToGif(f)
	zip := secHex(4) + ".zip"
	fCompress(zip, []string{cf})

	c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})

	return err
}

func downloadLineSToZip(c tele.Context, ud *UserData) error {
	workDir := filepath.Dir(ud.lineData.files[0])
	zipName := ud.lineData.id + ".zip"
	zipPath := filepath.Join(workDir, zipName)
	fCompress(zipPath, ud.lineData.files)
	_, err := c.Bot().Send(c.Recipient(), &tele.Document{FileName: zipName, File: tele.FromDisk(zipPath)})
	return err
}
