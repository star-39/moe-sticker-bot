package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	tele "gopkg.in/telebot.v3"
)

// func downloadSAndC(path string, s *tele.Sticker, needConvert bool, shrinkGif bool, c tele.Context) (string, string) {
// 	var f string
// 	var cf string
// 	var err error
// 	if s.Video {
// 		f = path + ".webm"
// 		err = c.Bot().Download(&s.File, f)
// 		if needConvert {
// 			if shrinkGif {
// 				cf, _ = ffToGifShrink(f)
// 			} else {
// 				cf, _ = ffToGif(f)
// 			}
// 		}
// 	} else if s.Animated {
// 		f = path + ".tgs"
// 		err = c.Bot().Download(&s.File, f)
// 		if needConvert {
// 			cf, _ = lottieToGIF(f)
// 		}
// 	} else {
// 		f = path + ".webp"
// 		err = c.Bot().Download(&s.File, f)
// 		if needConvert {
// 			cf, _ = imToPng(f)
// 		}
// 	}
// 	if err != nil {
// 		return "", ""
// 	}
// 	return f, cf
// }

// When s is not nil, download single sticker,
// otherwise, download whole set from setID.
func downloadStickersAndSend(s *tele.Sticker, setID string, c tele.Context) error {
	var id string
	if s != nil {
		id = s.SetName
	} else {
		id = setID
	}
	//Make sure only one download for one sticker set
	//is in progress.
	if _, exist := downloadQueue.ss[id]; exist {
		return sendDownloadInProgressWarning(c)

	}
	defer func() {
		if _, ok := downloadQueue.ss[id]; ok {
			downloadQueue.mu.Lock()
			delete(downloadQueue.ss, id)
			downloadQueue.mu.Unlock()
		}
	}()
	downloadQueue.mu.Lock()
	downloadQueue.ss[id] = true
	downloadQueue.mu.Unlock()

	sID := secHex(8)
	ud := &UserData{
		workDir:     filepath.Join(dataDir, sID),
		stickerData: &StickerData{},
		lineData:    &LineData{},
	}
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	workDir := filepath.Join(ud.workDir, id)
	os.MkdirAll(workDir, 0755)
	var flist []string
	var cflist []string
	var err error

	if s != nil {
		obj := &StickerDownloadObject{
			wg:          sync.WaitGroup{},
			bot:         b,
			sticker:     *s,
			dest:        filepath.Join(workDir, s.SetName+"_"+s.Emoji),
			needConvert: true,
			shrinkGif:   false,
			forWebApp:   false,
		}
		obj.wg.Add(1)
		wDownloadStickerObject(obj)
		if obj.err != nil {
			return err
		}
		if s.Video || s.Animated {
			zip := filepath.Join(workDir, secHex(4)+".zip")
			fCompress(zip, []string{obj.cf})
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})
		} else {
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(obj.cf), File: tele.FromDisk(obj.cf)})
		}
		return err
	}

	ss, _ := c.Bot().StickerSet(setID)
	ud.stickerData.id = ss.Name
	ud.stickerData.title = ss.Title
	pText, teleMsg, _ := sendProcessStarted(ud, c, "")
	sendNotifySDOnBackground(c)
	cleanUserData(c.Sender().ID)
	defer os.RemoveAll(workDir)
	var wpDownloadSticker *ants.PoolWithFunc
	defer wpDownloadSticker.Release()
	if ss.Animated {
		wpDownloadSticker, _ = ants.NewPoolWithFunc(1, wDownloadStickerObject)
	} else {
		wpDownloadSticker, _ = ants.NewPoolWithFunc(8, wDownloadStickerObject)
	}
	var objs []*StickerDownloadObject
	for index, s := range ss.Stickers {
		obj := &StickerDownloadObject{
			wg:          sync.WaitGroup{},
			bot:         b,
			sticker:     s,
			dest:        filepath.Join(workDir, fmt.Sprintf("%s_%d_%s", setID, index+1, s.Emoji)),
			needConvert: true,
			shrinkGif:   false,
			forWebApp:   false,
		}
		obj.wg.Add(1)
		objs = append(objs, obj)
		wpDownloadSticker.Invoke(obj)
		go editProgressMsg(index, len(ss.Stickers), "", pText, teleMsg, c)
	}
	for _, obj := range objs {
		obj.wg.Wait()
		flist = append(flist, obj.of)
		cflist = append(cflist, obj.cf)
	}
	go editProgressMsg(0, 0, "Uploading...", pText, teleMsg, c)

	webmZipPath := filepath.Join(workDir, setID+"_webm.zip")
	webpZipPath := filepath.Join(workDir, setID+"_webp.zip")
	pngZipPath := filepath.Join(workDir, setID+"_png.zip")
	gifZipPath := filepath.Join(workDir, setID+"_gif.zip")
	tgsZipPath := filepath.Join(workDir, setID+"_tgs.zip")

	var zipPaths []string

	if ss.Video {
		zipPaths = append(zipPaths, fCompressVol(webmZipPath, flist)...)
		zipPaths = append(zipPaths, fCompressVol(gifZipPath, cflist)...)
	} else if ss.Animated {
		zipPaths = append(zipPaths, fCompressVol(tgsZipPath, flist)...)
		zipPaths = append(zipPaths, fCompressVol(gifZipPath, cflist)...)
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
	workDir := filepath.Join(dataDir, secHex(4))
	os.MkdirAll(workDir, 0755)
	f := filepath.Join(workDir, "gif.mp4")
	err := c.Bot().Download(&c.Message().Animation.File, f)
	cf, _ := ffToGif(f)
	zip := filepath.Join(workDir, secHex(4)+".zip")
	fCompress(zip, []string{cf})

	c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})
	os.RemoveAll(workDir)
	return err
}

func downloadLineSToZip(c tele.Context, ud *UserData) error {
	err := prepImportStickers(ud, false)
	if err != nil {
		return err
	}
	workDir := filepath.Dir(ud.lineData.files[0])
	zipName := ud.lineData.id + ".zip"
	zipPath := filepath.Join(workDir, zipName)
	fCompress(zipPath, ud.lineData.files)
	_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: zipName, File: tele.FromDisk(zipPath)})
	return err
}
