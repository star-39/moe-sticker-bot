package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

func parseImportLink(link string, lineData *LineData) error {
	page, err := httpGet(link)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Errorln("Failed gq parsing line link!", err)
		return err
	}

	var jsonData map[string]interface{}
	// For LINE STORE, the first script is always sticker's metadata in JSON.
	err = json.Unmarshal([]byte(doc.Find("script").First().Text()), &jsonData)
	if err != nil {
		log.Errorln("Failed json parsing line link!", err)
		return err
	}
	t := jsonData["name"].(string)
	i := jsonData["sku"].(string)
	u := jsonData["url"].(string)
	a := false
	c := ""
	d := "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/"

	if strings.Contains(page, "MdIcoPlay_b") || strings.Contains(page, "MdIcoAni_b") {
		c = LINE_STICKER_ANIMATION
		d += "stickerpack@2x.zip"
		a = true
	} else if strings.Contains(page, "MdIcoMessageSticker_b") {
		d = u
		c = LINE_STICKER_MESSAGE
	} else if strings.Contains(page, "MdIcoNameSticker_b") {
		d += "sticker_name_base@2x.zip"
		c = LINE_STICKER_NAME
	} else if strings.Contains(page, "MdIcoFlash_b") || strings.Contains(page, "MdIcoFlashAni_b") {
		c = LINE_STICKER_POPUP
		d += "stickerpack@2x.zip"
		a = true
	} else if strings.Contains(page, "MdIcoEffectSoundSticker_b") || strings.Contains(page, "MdIcoEffectSticker_b") {
		c = LINE_STICKER_POPUP_EFFECT
		d += "stickerpack@2x.zip"
		a = true
	} else {
		c = LINE_STICKER_STATIC
		d += "stickers@2x.zip"
	}

	lineData.link = u
	lineData.category = c
	lineData.dLink = d
	lineData.id = i
	lineData.title = t
	lineData.isAnimated = a
	log.Debugln("line data parsed:", lineData)
	return nil
}

func prepLineStickers(ud *UserData) error {
	// ud.wg.Add(1)
	workDir := filepath.Join(ud.userDir, ud.lineData.id)
	savePath := filepath.Join(workDir, "line.zip")
	os.MkdirAll(workDir, 0755)

	err := fDownload(ud.lineData.dLink, savePath)
	if err != nil {
		return err
	}
	pngFiles := lineZipExtract(savePath, ud.lineData)
	if len(pngFiles) == 0 {
		return errors.New("no line image")
	}

	ud.lineData.files = pngFiles
	ud.lineData.amount = len(pngFiles)

	ud.stickerData.isVideo = ud.lineData.isAnimated
	ud.stickerData.lAmount = len(pngFiles)
	ud.stickerData.id = ud.lineData.category + ud.lineData.id + secHex(2) + "_by_" + botName
	ud.stickerData.title = ud.lineData.title
	ud.stickerData.link = "https://t.me/addstickers/" + ud.stickerData.id

	for _, f := range pngFiles {
		sf := &StickerFile{oPath: f}
		sf.wg = sync.WaitGroup{}
		sf.wg.Add(1)
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
	}
	log.Debugln("start converting...")
	doConvert(ud)

	log.Debug("Done preparing line files:")
	log.Debugln(ud.lineData, ud.stickerData)

	// ud.wg.Done()

	return nil

}

func lineZipExtract(f string, ld *LineData) []string {
	var files []string
	workDir := fExtract(f)
	if workDir == "" {
		return nil
	}
	log.Debugln("scanning workdir:", workDir)

	switch ld.category {
	case LINE_STICKER_ANIMATION:
		files, _ = filepath.Glob(filepath.Join(workDir, "animation@2x", "*.png"))
	case LINE_STICKER_POPUP:
		files = lsFilesR(workDir, []string{".png", "popup"}, []string{"tab", "key", "json"})
	case LINE_STICKER_POPUP_EFFECT:
		pfs, _ := filepath.Glob(filepath.Join(workDir, "popup"))
		for _, pf := range pfs {
			os.Rename(pf, filepath.Join(workDir, strings.TrimSuffix(path.Base(pf), ".png"), "@99.png"))
		}
		files, _ = filepath.Glob(filepath.Join(workDir, "*.png"))
	default:
		files = lsFiles(workDir, []string{".png"}, []string{"tab", "key", "json"})
	}

	return files
}

func doConvert(ud *UserData) {
	sf := ud.stickerData.stickers
	for _, s := range sf {
		var err error
		// If lineS is animated, commit to worker pool,
		// since encoding vp9 is time and resource costy.
		if ud.lineData.isAnimated {
			wpConvertWebm.Invoke(s)
		} else {
			s.cPath, err = imToWebp(s.oPath)
			if err != nil {
				s.cError = err
			}
			s.wg.Done()
		}
	}
}
