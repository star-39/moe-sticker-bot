package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func fDownload(link string, savePath string) error {
	cmd := exec.Command("curl", "-o", savePath, link)
	_, err := cmd.CombinedOutput()
	return err
}

func httpGet(link string) (string, error) {
	// cmd := exec.Command("curl", link)
	// output, err := cmd.CombinedOutput()
	// return string(output), err
	client := &http.Client{}
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "curl/7.61.1")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)

	return string(content), nil
}

func fExtract(f string) string {
	targetDir := filepath.Join(filepath.Dir(f), secHex(4))
	os.MkdirAll(targetDir, 0755)
	log.Debugln("Extracting to :", targetDir)

	err := exec.Command("tar", "-xvf", f, "-C", targetDir).Run()
	if err != nil {
		return ""
	} else {
		return targetDir
	}
}

func archiveExtract(f string) []string {
	targetDir := filepath.Join(path.Dir(f), secHex(4))
	os.MkdirAll(targetDir, 0755)
	err := exec.Command("tar", "-xvf", f, "-C", targetDir).Run()
	if err != nil {
		return []string{}
	}
	return lsFilesR(targetDir, []string{}, []string{})
}

func lsFilesR(dir string, mustHave []string, mustNotHave []string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// if info.IsDir() {
		// 	return nil
		// }
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
	dir := path.Dir(f)
	// strip data dir in zip.
	comps := strconv.Itoa(len(strings.Split(dir, string(os.PathSeparator))) + 1)
	var bin string
	if runtime.GOOS == "linux" {
		bin = "bsdtar"
	} else {
		bin = "tar"
	}

	args := []string{"--strip-components", comps, "-avcf", f}
	args = append(args, flist...)

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	return cmd.Run()
}
