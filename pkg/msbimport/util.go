package msbimport

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/util"
)

var BSDTAR_BIN = "bsdtar"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func httpDownload(link string, f string) error {
	res, err := http.Get(link)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fp, _ := os.Create(f)
	defer fp.Close()
	_, err = io.Copy(fp, res.Body)
	return err
}

func httpDownloadCurlUA(link string, f string) error {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "curl/7.61.1")
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fp, _ := os.Create(f)
	defer fp.Close()
	_, err = io.Copy(fp, res.Body)
	return err
}

func httpGet(link string) (string, error) {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-Hant;q=0.9, ja;q=0.8, en;q=0.7")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return string(content), nil
}

// redirected link, body, error
func httpGetWithRedirLink(link string) (string, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-Hant;q=0.9, ja;q=0.8, en;q=0.7")
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return res.Request.URL.String(), string(content), nil
}

func httpGetCurlUA(link string) (string, error) {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "curl/7.61.1")
	req.Header.Set("Accept-Language", "zh-Hant;q=0.9, ja;q=0.8, en;q=0.7")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)

	return string(content), nil
}

func httpGetAndroidUA(link string) (string, error) {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Android")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return string(content), nil
}

func httpPost(link string, data string) (string, error) {
	bdata := []byte(data)
	req, err := http.Post(link, "Content-Type: text/plain",
		bytes.NewBuffer(bdata))
	if err != nil {
		return "", err
	}

	resbody := req.Body
	res, err := io.ReadAll(resbody)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func fDownload(link string, savePath string) error {
	cmd := exec.Command("curl", "-o", savePath, link)
	_, err := cmd.CombinedOutput()
	return err
}

func fExtract(f string) string {
	targetDir := filepath.Join(filepath.Dir(f), util.SecHex(4))
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
