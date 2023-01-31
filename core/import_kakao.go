package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

func parseKakaoLink(link string, ld *LineData) error {
	url, _ := url.Parse(link)

	var kakaoID string
	var eid string
	var err error

	switch url.Host {
	// Kakao web link.
	case "e.kakao.com":
		kakaoID = path.Base(url.Path)
	// Kakao mobile app share link.
	case "emoticon.kakao.com":
		eid, kakaoID, err = fetchKakaoDetailsFromShareLink(link)
	// unknown host
	default:
		return errors.New("unknown kakao link type")
	}
	if err != nil {
		return err
	}

	var kakaoJson KakaoJson
	err = fetchKakaoMetadata(&kakaoJson, kakaoID)
	if err != nil {
		return err
	}

	log.Debugln("Parsed kakao link:", link)
	log.Debugln(kakaoJson.Result)

	isAnimated := checkKakaoAnimated(kakaoJson.Result.TitleDetailUrl)
	if isAnimated {
		ld.dLink = fmt.Sprintf("http://item.kakaocdn.net/dw/%s.file_pack.zip", eid)
	} else {
		ld.dLinks = kakaoJson.Result.ThumbnailUrls
	}

	ld.title = kakaoJson.Result.Title
	ld.id = strings.ReplaceAll(kakaoJson.Result.TitleUrl, "-", "_")
	ld.link = link
	ld.isAnimated = isAnimated
	ld.amount = len(ld.dLinks)
	return nil
}

func fetchKakaoMetadata(kakaoJson *KakaoJson, kakaoID string) error {
	apiUrl := "https://e.kakao.com/api/v1/items/t/" + kakaoID
	page, err := httpGet(apiUrl)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(page), &kakaoJson)
	if err != nil {
		log.Errorln("Failed json parsing kakao link!", err)
		return err
	}
	return nil
}

func prepareKakaoStickers(ud *UserData, needConvert bool) error {
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	ud.stickerData.id = "kakao_" + ud.lineData.id + secHex(2) + "_by_" + botName

	if ud.lineData.isAnimated {
		return prepareKakaoAnimatedStickers(ud, needConvert)
	}

	workDir := filepath.Join(ud.workDir, ud.lineData.id)
	os.MkdirAll(workDir, 0755)

	for range ud.lineData.dLinks {
		sf := &StickerFile{}
		sf.wg.Add(1)
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
		ud.stickerData.lAmount += 1
	}
	for i, l := range ud.lineData.dLinks {
		select {
		case <-ud.ctx.Done():
			log.Warn("prepKakaoStickers received ctxDone!")
			return nil
		default:
		}
		f := filepath.Join(workDir, path.Base(l)+".png")
		err := httpDownload(l, f)
		if err != nil {
			return err
		}
		cf, _ := imToWebp(f)
		ud.lineData.files = append(ud.lineData.files, f)
		ud.stickerData.stickers[i].oPath = f
		ud.stickerData.stickers[i].cPath = cf
		ud.stickerData.stickers[i].wg.Done()

		log.Debug("Done process one kakao emoticon")
		log.Debugf("f:%s, cf:%s", f, cf)
	}
	log.Debug("Done process ALL kakao emoticons")
	return nil
}

func prepareKakaoAnimatedStickers(ud *UserData, needConvert bool) error {
	workDir := filepath.Join(ud.workDir, ud.lineData.id)
	savePath := filepath.Join(workDir, "kakao.zip")
	os.MkdirAll(workDir, 0755)

	ud.wg.Add(1)
	err := fDownload(ud.lineData.dLink, savePath)
	if err != nil {
		return err
	}

	webpFiles := kakaoZipExtract(savePath, ud.lineData)
	if len(webpFiles) == 0 {
		return errors.New("no kakao image")
	}
	ud.wg.Done()

	ud.lineData.files = webpFiles
	ud.lineData.amount = len(webpFiles)
	ud.stickerData.lAmount = len(webpFiles)

	for _, f := range webpFiles {
		sf := &StickerFile{oPath: f}
		sf.wg = sync.WaitGroup{}
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
	}

	if needConvert {
		convertSToTGFormat(ud)
	}

	log.Debug("Done preparing kakao files:")
	log.Debugln(ud.lineData, ud.stickerData)

	return nil

}

// Extract and decrypt kakao zip.
func kakaoZipExtract(f string, ld *LineData) []string {
	var files []string
	workDir := fExtract(f)
	if workDir == "" {
		return nil
	}
	log.Debugln("scanning workdir:", workDir)
	files = lsFiles(workDir, []string{".webp"}, []string{})

	for _, f := range files {
		//This script decrypts the file in-place.
		exec.Command("msb_kakao_decrypt.py", f).Run()
	}

	return files
}

// kakao eid(code), kakao id
func fetchKakaoDetailsFromShareLink(link string) (string, string, error) {
	res, err := httpGetAndroidUA(link)
	if err != nil {
		return "", "", err
	}
	split1 := strings.Split(res, "kakaotalk://store/emoticon/")
	if len(split1) < 2 {
		return "", "", errors.New("error fetchKakaoDetailsFromShareLink")
	}
	eid := strings.Split(split1[1], "?")[0]
	log.Debugln("kakao eid is: ", eid)
	redirLink, _, err := httpGetWithRedirLink(link)
	if err != nil {
		return "", "", err
	}
	kakaoID := path.Base(redirLink)
	return eid, kakaoID, nil
}

// Receive kakaoJson.Result.TitleDetailUrl
func checkKakaoAnimated(ilink string) bool {
	res, err := http.Get(ilink)
	if err != nil {
		return false
	}
	if res.Header.Get("Content-Type") == "image/gif" {
		return true
	} else {
		return false
	}
}
