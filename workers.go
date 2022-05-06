package main

import log "github.com/sirupsen/logrus"

func wConvertWebm(i interface{}) {
	sf := i.(*StickerFile)
	log.Debugln("Converting in pool for:", sf)

	var err error
	sf.cPath, err = ffToWebm(sf.oPath)
	if err != nil {
		sf.cError = err
	}
	log.Debugln("convert OK: ", sf.cPath)
	sf.wg.Done()
}
