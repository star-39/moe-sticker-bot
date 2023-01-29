package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func fDownload(link string, savePath string) error {
	cmd := exec.Command("curl", "-o", savePath, link)
	_, err := cmd.CombinedOutput()
	return err
}

func fExtract(f string) string {
	targetDir := filepath.Join(filepath.Dir(f), secHex(4))
	os.MkdirAll(targetDir, 0755)
	log.Debugln("Extracting to :", targetDir)

	out, err := exec.Command(BSDTAR_BIN, "-xvf", f, "-C", targetDir).CombinedOutput()
	if err != nil {
		log.Errorln("Error extracting:", string(out))
		return ""
	} else {
		return targetDir
	}
}

func archiveExtract(f string) []string {
	targetDir := filepath.Join(path.Dir(f), secHex(4))
	os.MkdirAll(targetDir, 0755)

	err := exec.Command(BSDTAR_BIN, "-xvf", f, "-C", targetDir).Run()
	if err != nil {
		return []string{}
	}
	return lsFilesR(targetDir, []string{}, []string{})
}

func lsFilesR(dir string, mustHave []string, mustNotHave []string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		accept := true
		confidence := 0
		for _, kw := range mustHave {
			if !strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				confidence += 1
			}
		}
		if confidence < len(mustHave) {
			accept = false
		}

		for _, kw := range mustNotHave {
			if strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		if info.IsDir() {
			accept = false
		}
		log.Debugf("accept?: %t path: %s", accept, path)
		if accept {
			files = append(files, path)
		}
		return err
	})
	log.Debugln("listed following:")
	log.Debugln(files)
	if err != nil {
		return []string{}
	} else {
		return files
	}
}

func lsFiles(dir string, mustHave []string, mustNotHave []string) []string {
	var files []string
	glob, _ := filepath.Glob(path.Join(dir, "*"))

	for _, path := range glob {
		f, _ := os.Stat(path)
		if f.IsDir() {
			continue
		}

		accept := true
		for _, kw := range mustHave {
			if !strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		for _, kw := range mustNotHave {
			if strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		log.Debugf("accept?: %t path: %s", accept, path)
		if accept {
			files = append(files, path)
		}
	}
	return files
}

func fCompress(f string, flist []string) error {
	// strip data dir in zip.
	// comps are 2
	comps := "2"

	args := []string{"--strip-components", comps, "-avcf", f}
	// args := []string{"-avcf", f}
	args = append(args, flist...)

	log.Debugf("Compressing strip-comps:%s to file:%s for these files:%v", comps, f, flist)
	out, err := exec.Command(BSDTAR_BIN, args...).CombinedOutput()
	log.Debugln(string(out))
	if err != nil {
		log.Error("Compress error!")
		log.Errorln(string(out))
	}
	return err
}

func fCompressVol(f string, flist []string) []string {
	basename := filepath.Base(f)
	dir := filepath.Dir(f)
	zipIndex := 0
	var zips [][]string
	var zipPaths []string
	var curSize int64 = 0

	for _, f := range flist {
		st, err := os.Stat(f)
		if err != nil {
			continue
		}
		fSize := st.Size()
		if curSize == 0 {
			zips = append(zips, []string{})
		}
		if curSize+fSize < 50000000 {
			zips[zipIndex] = append(zips[zipIndex], f)
		} else {
			curSize = 0
			zips = append(zips, []string{})
			zipIndex += 1
			zips[zipIndex] = append(zips[zipIndex], f)
		}
		curSize += fSize
	}

	for i, files := range zips {
		var zipBN string
		if len(zips) == 1 {
			zipBN = basename
		} else {
			zipBN = strings.TrimSuffix(basename, ".zip")
			zipBN += fmt.Sprintf("_00%d.zip", i+1)
		}

		zipPath := filepath.Join(dir, zipBN)
		err := fCompress(zipPath, files)
		if err != nil {
			return nil
		}
		zipPaths = append(zipPaths, zipPath)
	}
	return zipPaths
}

func purgeOutdatedStorageData() {
	dirEntries, _ := os.ReadDir(dataDir)
	for _, f := range dirEntries {
		if !f.IsDir() {
			continue
		}
		fInfo, _ := f.Info()
		fMTime := fInfo.ModTime().Unix()
		fPath := filepath.Join(dataDir, f.Name())
		// 2 Days
		if fMTime < (time.Now().Unix() - 172800) {
			os.RemoveAll(fPath)
			users.mu.Lock()
			for uid, ud := range users.data {
				if ud.sessionID == f.Name() {
					log.Warnf("Found outdated user data. Purging from map as well. SID:%s, UID:%d", ud.sessionID, uid)
					delete(users.data, uid)
					break
				}
			}
			users.mu.Unlock()
			log.Infoln("Purged outdated user dir:", fPath)
		}
	}

	webappDataDirEntries, _ := os.ReadDir(Config.WebappDataDir)
	for _, f := range webappDataDirEntries {
		if !f.IsDir() {
			continue
		}
		fInfo, _ := f.Info()
		fMTime := fInfo.ModTime().Unix()
		fPath := filepath.Join(Config.WebappDataDir, f.Name())
		// 2 Days
		if fMTime < (time.Now().Unix() - 172800) {
			os.RemoveAll(fPath)
			log.Infoln("Purged outdated webapp data dir:", fPath)
		}
	}
}
