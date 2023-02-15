package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/star-39/moe-sticker-bot/pkg/convert"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
)

func main() {
	var link = flag.String("link", "", "Import link(LINE/kakao)")
	var convertTG = flag.Bool("convert", false, "Convert to Telegram format(WEBP/WEBM)")
	var outputJson = flag.Bool("json", false, "Output JSON serialized LineData, useful when integrating with other apps.")
	var workDir = flag.String("dir", "", "Where to put sticker files.")
	var logLevel = flag.String("log_level", "debug", "Log level")
	flag.Parse()

	if *outputJson {
		log.SetLevel(log.FatalLevel)
	} else {
		ll, err := log.ParseLevel(*logLevel)
		if err != nil {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(ll)
		}
	}

	convert.InitConvert()

	ctx, _ := context.WithCancel(context.Background())
	ld := &msbimport.LineData{}

	// LineData will be parsed to ld.
	warn, err := msbimport.ParseImportLink(*link, ld)
	if err != nil {
		log.Error("Error parsing import link!")
		log.Fatalln(err)
	}
	if warn != "" {
		log.Warnln(warn)
	}

	err = msbimport.PrepareImportStickers(ctx, ld, *workDir, *convertTG)
	if err != nil {
		log.Fatalln(err)
	}

	for _, lf := range ld.Files {
		lf.Wg.Wait()
		if lf.CError != nil {
			log.Error(lf.CError)
		}
		log.Infoln("Original File:", lf.OriginalFile)
		if *convertTG {
			log.Infoln("Converted File:", lf.ConvertedFile)
		}
	}

	if *outputJson {
		ld.TitleWg.Wait()
		jbytes, err := json.Marshal(ld)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Print(string(jbytes))
	}
}
