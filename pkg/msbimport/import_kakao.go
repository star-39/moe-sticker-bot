package msbimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func parseKakaoLink(link string, ld *LineData) (string, error) {
	var kakaoID string
	var eid string
	var err error
	var warn string

	url, _ := url.Parse(link)

	switch url.Host {
	// Kakao web link.
	case "e.kakao.com":
		kakaoID = path.Base(url.Path)
	// Kakao mobile app share link.
	case "emoticon.kakao.com":
		eid, kakaoID, err = fetchKakaoDetailsFromShareLink(link)
		if err != nil {
			return warn, err
		}
	// unknown host
	default:
		return warn, errors.New("unknown kakao link type")
	}

	var kakaoJson KakaoJson
	err = fetchKakaoMetadata(&kakaoJson, kakaoID)
	if err != nil {
		log.Debugln("Failed fetchKakaoMetadata:", err)
		return warn, err
	}

	log.Debugln("Parsed kakao link:", link)
	log.Debugln(kakaoJson.Result)

	if url.Host == "emoticon.kakao.com" {
		ld.DLink = fmt.Sprintf("http://item.kakaocdn.net/dw/%s.file_pack.zip", eid)
	} else {
		ld.DLinks = kakaoJson.Result.ThumbnailUrls
		warn = WARN_KAKAO_PREFER_SHARE_LINK
	}

	ld.Title = kakaoJson.Result.Title
	ld.Id = kakaoJson.Result.TitleUrl
	ld.Link = link
	ld.Amount = len(ld.DLinks)
	ld.Category = KAKAO_EMOTICON
	return warn, nil
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

// Download and convert(if needed) stickers to work directory.
// *ld will be modified and loaded with local sticker information.
func prepareKakaoStickers(ctx context.Context, ld *LineData, workDir string, needConvert bool) error {
	// If no dLink, continue importing static ones.
	if ld.DLink != "" {
		return prepareKakaoZipStickers(ctx, ld, workDir, needConvert)
	}

	os.MkdirAll(workDir, 0755)

	//Initialize Files with wg added.
	//This is intended for async operation.
	//When user reached commitSticker state, sticker will be waited one by one.
	for range ld.DLinks {
		lf := &LineFile{}
		lf.Wg.Add(1)
		ld.Files = append(ld.Files, lf)
	}

	//Download stickers one by one.
	go func() {
		for i, l := range ld.DLinks {
			select {
			case <-ctx.Done():
				log.Warn("prepareKakaoStickers received ctxDone!")
				return
			default:
			}

			f := filepath.Join(workDir, path.Base(l)+".png")
			err := httpDownload(l, f)
			if err != nil {
				ld.Files[i].CError = err
			}
			cf, _ := IMToWebpTGStatic(f, false)
			ld.Files[i].OriginalFile = f
			ld.Files[i].ConvertedFile = cf
			ld.Files[i].Wg.Done()

			log.Debug("Done process one kakao emoticon")
			log.Debugf("f:%s, cf:%s", f, cf)
		}
		log.Debug("Done process ALL kakao emoticons")
	}()
	return nil
}

func prepareKakaoZipStickers(ctx context.Context, ld *LineData, workDir string, needConvert bool) error {
	zipPath := filepath.Join(workDir, "kakao.zip")
	os.MkdirAll(workDir, 0755)

	err := fDownload(ld.DLink, zipPath)
	if err != nil {
		return err
	}

	kakaoFiles := kakaoZipExtract(zipPath, ld)
	if len(kakaoFiles) == 0 {
		return errors.New("no kakao image in zip")
	}

	if filepath.Ext(kakaoFiles[0]) != ".png" {
		ld.IsAnimated = true
	}

	for _, wf := range kakaoFiles {
		lf := &LineFile{
			OriginalFile: wf,
		}
		if needConvert {
			lf.Wg.Add(1)
		}
		ld.Files = append(ld.Files, lf)
	}
	ld.Amount = len(kakaoFiles)

	if needConvert {
		go convertSToTGFormat(ctx, ld)
	}

	log.Debug("Done preparing kakao files:")
	log.Debugln(ld)

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
	files = LsFiles(workDir, []string{}, []string{})

	for _, f := range files {
		//PNG is not encrypted.
		if filepath.Ext(f) != ".png" {
			//This script decrypts the file in-place.
			exec.Command("msb_kakao_decrypt.py", f).Run()
		}
	}
	return files
}

// kakao eid(code), kakao id
func fetchKakaoDetailsFromShareLink(link string) (string, string, error) {
	log.Debugln("fetchKakaoDetailsFromShareLink: Link is:", link)
	res, err := httpGetAndroidUA(link)
	if err != nil {
		log.Errorln("fetchKakaoDetailsFromShareLink: failed httpGetAndroidUA!", err)
		return "", "", err
	}

	eidRegex := regexp.MustCompile(`data-i="(\d+)"`)
	rawEid := eidRegex.FindStringSubmatch(res)[1]
	rawEidInt, err := strconv.Atoi(string(rawEid))
	if err != nil {
		log.Errorln("fetchKakaoDetailsFromShareLink: failed to parse rawEid!", err)
		return "", "", errors.New("error fetchKakaoDetailsFromShareLink")
	}
	xorKeysRegex := regexp.MustCompile(`(\d+)\^(\d+)`)
	xorKeys := xorKeysRegex.FindStringSubmatch(res)
	xorKey1, _ := strconv.Atoi(string(xorKeys[1]))
	xorKey2, err := strconv.Atoi(string(xorKeys[2]))
	if err != nil {
		log.Errorln("fetchKakaoDetailsFromShareLink: failed to parse XOR Keys!", err)
		return "", "", errors.New("error fetchKakaoDetailsFromShareLink")
	}

	eid := fmt.Sprintf("%d", rawEidInt-xorKey1^xorKey2)
	log.Debugln("kakao eid is: ", eid)
	redirLink, _, err := httpGetWithRedirLink(link)
	if err != nil {
		return "", "", err
	}
	kakaoID := path.Base(redirLink)
	return eid, kakaoID, nil
}
