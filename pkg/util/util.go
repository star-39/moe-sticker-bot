package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const BSDTAR_BIN = "bsdtar"

func SecHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func ArchiveExtract(f string) []string {
	targetDir := filepath.Join(path.Dir(f), SecHex(4))
	os.MkdirAll(targetDir, 0755)

	err := exec.Command(BSDTAR_BIN, "-xvf", f, "-C", targetDir).Run()
	if err != nil {
		return []string{}
	}
	return LsFilesR(targetDir, []string{}, []string{})
}

func LsFilesR(dir string, mustHave []string, mustNotHave []string) []string {
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

func LsFiles(dir string, mustHave []string, mustNotHave []string) []string {
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

func FCompress(f string, flist []string) error {
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

func FCompressVol(f string, flist []string) []string {
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
		err := FCompress(zipPath, files)
		if err != nil {
			return nil
		}
		zipPaths = append(zipPaths, zipPath)
	}
	return zipPaths
}
