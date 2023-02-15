package msbimport

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/convert"
	"github.com/star-39/moe-sticker-bot/pkg/util"
)

func parseLineLink(link string, ld *LineData) (string, error) {
	var warn string
	page, err := httpGet(link)
	if err != nil {
		return warn, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Errorln("Failed gq parsing line link!", err)
		return warn, err
	}

	var lineJson LineJson
	err = parseLineDetails(doc, &lineJson)
	if err != nil {
		log.Errorln("parseLineLink: ", err)
		return warn, err
	}

	t := lineJson.Name
	i := lineJson.Sku
	u := lineJson.Url
	ls := fetchLineI18nLinks(doc)
	a := false
	c := ""
	d := "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/"

	if strings.Contains(u, "stickershop") || strings.Contains(u, "officialaccount/event/sticker") {
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
			// According to collected logs, LINE ID befre exact 775 have special PNG encodings,
			// which are not parsable with libpng.
			// You will get -> CgBI: unhandled critical chunk <- from IM.
			// Workaround is to take the lower resolution "android" ones.
			if id, _ := strconv.Atoi(i); id < 775 && id != 0 {
				d = "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/android/" +
					"stickers.zip"
			} else {
				d += "stickers@2x.zip"
			}
		}
	} else if strings.Contains(u, "emojishop") {
		if strings.Contains(page, "MdIcoPlay_b") {
			c = LINE_EMOJI_ANIMATION
			d = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + i + "/iphone/package_animation.zip"
			a = true
		} else {
			c = LINE_EMOJI_STATIC
			d = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + i + "/iphone/package.zip"
		}
	} else {
		return warn, errors.New("unknown line store category")
	}
	if ld == nil {
		return warn, nil
	}

	ld.Link = u
	ld.I18nLinks = ls
	ld.Category = c
	ld.DLink = d
	ld.Id = i
	ld.Title = t
	ld.IsAnimated = a

	log.Debugln("line data parsed:", ld)

	ld.TitleWg.Add(1)
	go fetchLineI18nTitles(ld)
	return warn, nil
}

func fetchLineI18nLinks(doc *goquery.Document) []string {
	var i18nLinks []string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		hreflang, exist := s.Attr("hreflang")
		if !exist {
			return
		}
		href, exist2 := s.Attr("href")
		if !exist2 {
			return
		}
		switch hreflang {
		case "zh-Hant":
			fallthrough
		case "ja":
			fallthrough
		case "en":
			i18nLinks = append(i18nLinks, href)
		}
	})
	log.Debugln("Fetched LINE I18n Links: ", i18nLinks)
	return i18nLinks
}

func fetchLineI18nTitles(ld *LineData) {
	defer ld.TitleWg.Done()
	log.Debugln("Fetching LINE i18n titles...")
	var i18nTitles []string

	for _, l := range ld.I18nLinks {
		page, err := httpGet(l)
		if err != nil {
			continue
		}
		lineJson := &LineJson{}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
		if err != nil {
			continue
		}
		parseLineDetails(doc, lineJson)

		for _, t := range i18nTitles {
			// if title duplicates, skip loop
			if t == lineJson.Name {
				goto CONTINUE
			}
		}
		i18nTitles = append(i18nTitles, lineJson.Name)

	CONTINUE:
		continue
	}

	ld.I18nTitles = i18nTitles
	log.Debugln("Fetched I18N titles are:")
	log.Debugln(ld.I18nTitles)
}

// This function goes after parseLineLink
// Receives a gq document of LINE Store page and parse the details to *LineJson.
func parseLineDetails(doc *goquery.Document, lj *LineJson) error {
	// For typical line store page, the first script is sticker's metadata in JSON.
	// Parse to json, if OK, return nil here.
	err := json.Unmarshal([]byte(doc.Find("script").First().Text()), lj)
	if err == nil {
		return nil
	}

	// Some new line store page does not have a json metadata <script>
	log.Warnln("Failed json parsing line link!", err)
	log.Warnln("Special LINE type? Trying to guess info.")

	// Find URL.
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("property")
		if value == "og:url" {
			lj.Url, _ = s.Attr("content")
		}
	})

	// Find title.
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		oa, _ := s.Attr("data-test")
		if oa == "oa-sticker-title" {
			lj.Name = s.Text()
		}
	})
	// Find title again.
	if lj.Name == "" {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			oa, _ := s.Attr("data-test")
			if oa == "sticker-name-title" {
				lj.Name = s.Text()
			}
		})
	}
	if lj.Name == "" {
		return errors.New("guess line: no title")
	}

	// Find ID.
	var defaultLink string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		hreflang, _ := s.Attr("hreflang")
		if hreflang == "x-default" {
			defaultLink, _ = s.Attr("href")
		}
	})
	lj.Sku = path.Base(defaultLink)
	if lj.Sku == "" {
		return errors.New("guess line: no id(no link base)")
	}

	log.Debugln("parsed line detail:", lj)
	return nil
}

// Download and convert sticker files after parseLineLink.
func prepareLineStickers(ctx context.Context, ld *LineData, workDir string, needConvert bool) error {
	if ld.Category == LINE_STICKER_MESSAGE {
		return prepareLineMessageS(ctx, ld, workDir, needConvert)
	}

	savePath := filepath.Join(workDir, "line.zip")
	os.MkdirAll(workDir, 0755)

	err := fDownload(ld.DLink, savePath)
	if err != nil {
		return err
	}

	pngFiles := lineZipExtract(savePath, ld)
	if len(pngFiles) == 0 {
		return errors.New("no line image")
	}

	for _, pf := range pngFiles {
		lf := &LineFile{
			OriginalFile: pf,
		}
		if needConvert {
			lf.Wg.Add(1)
		}
		ld.Files = append(ld.Files, lf)
	}
	ld.Amount = len(pngFiles)

	if needConvert {
		log.Debugln("start converting...")
		go convertSToTGFormat(ctx, ld)
	}

	log.Debug("Done preparing line files:")
	log.Debugln(ld)

	return nil
}

func lineZipExtract(f string, ld *LineData) []string {
	var files []string
	workDir := fExtract(f)
	if workDir == "" {
		return nil
	}
	log.Debugln("scanning workdir:", workDir)

	switch ld.Category {
	case LINE_STICKER_ANIMATION:
		files, _ = filepath.Glob(filepath.Join(workDir, "animation@2x", "*.png"))
	case LINE_STICKER_POPUP:
		files, _ = filepath.Glob(filepath.Join(workDir, "popup", "*.png"))
	case LINE_STICKER_POPUP_EFFECT:
		pfs, _ := filepath.Glob(filepath.Join(workDir, "popup", "*.png"))
		for _, pf := range pfs {
			os.Rename(pf, filepath.Join(workDir, strings.TrimSuffix(filepath.Base(pf), ".png")+"@99.png"))
		}
		files = util.LsFiles(workDir, []string{".png"}, []string{"tab", "key", "json"})
	default:
		files = util.LsFiles(workDir, []string{".png"}, []string{"tab", "key", "json"})
	}
	if ld.IsAnimated {
		sanitizeLinePNGs(files)
	}
	return files
}

// Some LINE png contains a 'tEXt' textual information after `fcTL`
// FFMpeg's APNG demuxer could not parse it properly.
// I have patched ffmpeg, however, compiling it for AArch64 is not easy.
// Therefore, we are going to tackle the source itself by removing tEXt chunk.
func sanitizeLinePNGs(files []string) bool {
	for _, f := range files {
		ret := removeAPNGtEXtChunk(f)
		if !ret {
			log.Debugln("one file sanitization ignored:", f)
			//do nothing.
		}
	}
	return true
}

func removeAPNGtEXtChunk(f string) bool {
	bytes, err := os.ReadFile(f)
	if err != nil {
		return false
	}
	l := len(bytes)
	if l < 42 {
		return false
	}
	// byte index 37-40 must be acTL Animation Control Chunk.
	if string(bytes[37:41]) != "acTL" {
		return false
	}
	textStart := 0
	textEnd := 0
	for i := range bytes {
		if i > l-10 {
			break
		}
		tag := bytes[i : i+4]
		// only probe the first appearence of 'tEXt' tag.
		if string(tag) == "tEXt" && textStart == 0 {
			// 4 bytes before tag represents chunk length.
			textStart = i - 4
			// first IDAT after tEXt should be what we want.
		} else if string(tag) == "IDAT" && textStart != 0 {
			textEnd = i - 4
			break
		}
	}
	if textStart == 0 || textEnd == 0 {
		return false
	}
	newBytes := bytes[:textStart]
	newBytes = append(newBytes, bytes[textEnd:]...)

	os.Remove(f)
	fo, err := os.Create(f)
	if err != nil {
		return false
	}
	defer fo.Close()
	fo.Write(newBytes)
	log.Infoln("Sanitized one APNG file, path:", f)
	log.Infof("Length from %d to %d.", l, len(newBytes))
	return true
}

// Line message sticker is a composition of two stickers.
// One represents the backgroud and one represents the foreground text.
// We need to composite them together.
func prepareLineMessageS(ctx context.Context, ld *LineData, workDir string, needConvert bool) error {
	os.MkdirAll(workDir, 0755)

	//Only curl UA will work.
	page, err := httpGetCurlUA(ld.Link)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Errorln("Failed gq parsing line link!", err)
		return err
	}

	var baseImages []string
	var overlayImages []string
	var jsonData map[string]interface{}
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		jsonDP, exist := s.Attr("data-preview")
		if !exist {
			return
		}
		log.Debugln("Got one json data-preview:", jsonDP)

		err := json.Unmarshal([]byte(jsonDP), &jsonData)
		if err != nil {
			log.Warnln("Json parse failed!", jsonDP)
			return
		}
		baseImages = append(baseImages, jsonData["staticUrl"].(string))
		overlayImages = append(overlayImages, jsonData["customOverlayUrl"].(string))
	})
	log.Debugln("base images:", baseImages)
	log.Debugln("overlay images:", overlayImages)

	for range baseImages {
		lf := &LineFile{}
		lf.Wg.Add(1)
		ld.Files = append(ld.Files, lf)
	}
	ld.Amount = len(baseImages)

	go func() {
		for i, b := range baseImages {
			select {
			case <-ctx.Done():
				log.Warn("prepLineMessageS received ctxDone!")
				return
			default:
			}
			log.Debugln("Preparing one message sticker... index:", i)
			bPath := filepath.Join(workDir, strconv.Itoa(i)+".base.png")
			oPath := filepath.Join(workDir, strconv.Itoa(i)+".overlay.png")
			httpDownloadCurlUA(b, bPath)
			httpDownloadCurlUA(overlayImages[i], oPath)
			f, err := convert.IMStackToWebp(bPath, oPath)
			if err != nil {
				ld.Files[i].CError = err
			}
			ld.Files[i].ConvertedFile = f
			ld.Files[i].OriginalFile = f
			ld.Files[i].Wg.Done()
			log.Debugln("one line message sticker OK:", f)
		}
	}()

	return nil
}
