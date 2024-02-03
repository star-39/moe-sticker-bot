package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

// When s is not nil, download single sticker,
// otherwise, download whole set from setID.
func downloadStickersAndSend(s *tele.Sticker, setID string, c tele.Context) error {
	var id string
	if s != nil {
		id = s.SetName
	} else {
		id = setID
	}

	sID := secHex(8)
	ud := &UserData{
		workDir:     filepath.Join(dataDir, sID),
		stickerData: &StickerData{},
		lineData:    &msbimport.LineData{},
	}

	workDir := filepath.Join(ud.workDir, id)
	os.MkdirAll(workDir, 0755)
	var flist []string
	var cflist []string
	var err error

	if s != nil {
		obj := &StickerDownloadObject{
			wg:          sync.WaitGroup{},
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
			msbimport.FCompress(zip, []string{obj.cf})
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})
		} else {
			c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(obj.cf), File: tele.FromDisk(obj.cf)})
		}
		return err
	}

	ss, err := c.Bot().StickerSet(setID)
	if err != nil {
		return sendBadSNWarn(c)
	}
	ud.stickerData.id = ss.Name
	ud.stickerData.title = ss.Title
	pText, pMsg, _ := sendProcessStarted(ud, c, "")
	sendNotifyWorkingOnBackground(c)
	cleanUserData(c.Sender().ID)
	defer os.RemoveAll(workDir)
	var wpDownloadSticker *ants.PoolWithFunc

	if ss.Animated {
		wpDownloadSticker, _ = ants.NewPoolWithFunc(4, wDownloadStickerObject)
	} else {
		wpDownloadSticker, _ = ants.NewPoolWithFunc(8, wDownloadStickerObject)
	}

	defer wpDownloadSticker.Release()
	imageTime := time.Now()
	var objs []*StickerDownloadObject
	for index, s := range ss.Stickers {
		obj := &StickerDownloadObject{
			wg:          sync.WaitGroup{},
			sticker:     s,
			dest:        filepath.Join(workDir, fmt.Sprintf("%s_%d_%s", setID, index+1, s.Emoji)),
			needConvert: true,
			shrinkGif:   false,
			forWebApp:   false,
		}
		obj.wg.Add(1)
		objs = append(objs, obj)
		go wpDownloadSticker.Invoke(obj)
	}
	for i, obj := range objs {
		go editProgressMsg(i, len(ss.Stickers), "", pText, pMsg, c)
		obj.wg.Wait()
		imageTime = imageTime.Add(time.Duration(i+1) * time.Second)
		msbimport.SetImageTime(obj.of, imageTime)
		flist = append(flist, obj.of)
		cflist = append(cflist, obj.cf)
	}
	go editProgressMsg(0, 0, "Uploading...", pText, pMsg, c)

	webmZipPath := filepath.Join(workDir, setID+"_webm.zip")
	webpZipPath := filepath.Join(workDir, setID+"_webp.zip")
	pngZipPath := filepath.Join(workDir, setID+"_png.zip")
	gifZipPath := filepath.Join(workDir, setID+"_gif.zip")
	tgsZipPath := filepath.Join(workDir, setID+"_tgs.zip")

	var zipPaths []string

	if ss.Video {
		zipPaths = append(zipPaths, msbimport.FCompressVol(webmZipPath, flist)...)
		zipPaths = append(zipPaths, msbimport.FCompressVol(gifZipPath, cflist)...)
	} else if ss.Animated {
		zipPaths = append(zipPaths, msbimport.FCompressVol(tgsZipPath, flist)...)
		zipPaths = append(zipPaths, msbimport.FCompressVol(gifZipPath, cflist)...)
	} else {
		zipPaths = append(zipPaths, msbimport.FCompressVol(webpZipPath, flist)...)
		zipPaths = append(zipPaths, msbimport.FCompressVol(pngZipPath, cflist)...)
	}
	for _, zipPath := range zipPaths {
		_, err := c.Bot().Send(c.Recipient(), &tele.Document{FileName: filepath.Base(zipPath), File: tele.FromDisk(zipPath)})
		time.Sleep(1 * time.Second)
		if err != nil {
			return err
		}
	}

	editProgressMsg(0, 0, "success! /start", pText, pMsg, c)
	return nil
}

func downloadGifToZip(c tele.Context) error {
	c.Reply("Downloading, please wait...\n正在下載, 請稍等...")
	workDir := filepath.Join(dataDir, secHex(4))
	os.MkdirAll(workDir, 0755)
	defer os.RemoveAll(workDir)

	f := filepath.Join(workDir, "animation_MP4.mp4")
	err := teleDownload(&c.Message().Animation.File, f)
	if err != nil {
		return err
	}
	cf, _ := msbimport.FFToGif(f)
	cf2 := strings.ReplaceAll(cf, "animation_MP4.mp4", "animation_GIF.gif")
	os.Rename(cf, cf2)
	zip := filepath.Join(workDir, secHex(4)+".zip")
	msbimport.FCompress(zip, []string{cf2})

	_, err = c.Bot().Reply(c.Message(), &tele.Document{FileName: filepath.Base(zip), File: tele.FromDisk(zip)})
	return err
}

func downloadLineSToZip(c tele.Context, ud *UserData) error {
	workDir := filepath.Join(ud.workDir, ud.lineData.Id)
	err := msbimport.PrepareImportStickers(ud.ctx, ud.lineData, workDir, false)
	if err != nil {
		return err
	}
	for _, f := range ud.lineData.Files {
		f.Wg.Wait()
	}
	// workDir := filepath.Dir(ud.lineData.files[0])
	zipName := ud.lineData.Id + ".zip"
	zipPath := filepath.Join(workDir, zipName)

	var files []string
	for _, lf := range ud.lineData.Files {
		files = append(files, lf.OriginalFile)
	}
	msbimport.FCompress(zipPath, files)
	_, err = c.Bot().Send(c.Recipient(), &tele.Document{FileName: zipName, File: tele.FromDisk(zipPath)})
	endSession(c)
	return err
}
