package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/star-39/moe-sticker-bot/pkg/convert"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
)

func main() {
	var link = flag.String("link", "", "Import link(LINE/kakao)")
	var convertTG = flag.Bool("convert", false, "Convert to Telegram format(WEBP/WEBM)")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
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

	err = msbimport.PrepareImportStickers(ctx, ld, "./", *convertTG)
	if err != nil {
		log.Fatalln(err)
	}

	for _, lf := range ld.Files {
		lf.Wg.Wait()
		if lf.CError != nil {
			log.Error(lf.CError)
		}
		println(lf.OriginalFile)
		println(lf.ConvertedFile)
	}
}
