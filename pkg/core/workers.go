package core

import (
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
)

// Workers pool for converting webm
var wpConvertWebm *ants.PoolWithFunc
var wpDownloadStickerSet *ants.PoolWithFunc

func initWorkersPool() {
	wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)
	wpDownloadStickerSet, _ = ants.NewPoolWithFunc(
		16, wDownloadStickerSet)
}

func wConvertWebm(i interface{}) {
	sf := i.(*StickerFile)
	defer sf.wg.Done()
	log.Debugln("Converting in pool for:", sf)

	var err error
	sf.cPath, err = ffToWebm(sf.oPath)
	if err != nil {
		sf.cError = err
	}
	log.Debugln("convert OK: ", sf.cPath)
}

func wDownloadStickerSet(i interface{}) {
	obj := i.(*StickerDownloadObject)
	defer obj.wg.Done()
	log.Debugf("Downloading in pool: %s -> %s", obj.sticker.FileID, obj.dest)
	err := obj.bot.Download(&obj.sticker.File, obj.dest)
	if err != nil {
		obj.err = err
	}
}

func wSubmitSMove(i interface{}) {
	obj := i.(*StickerMoveObject)
	defer obj.wg.Done()
	sid := obj.sd.stickerSet.Stickers[obj.oldIndex].FileID
	log.Debugf("Moving in pool %d(%s) -> %d", obj.oldIndex, sid, obj.newIndex)
	err := b.SetStickerPosition(sid, obj.newIndex)
	if err != nil {
		log.Debugln("SMove failed!!", err)
		obj.err = err
	} else {
		log.Debugf("Sticker move OK for %s", obj.sd.stickerSet.Name)
		obj.sd.stickerSet.Stickers =
			sliceMove(obj.oldIndex, obj.newIndex, obj.sd.stickerSet.Stickers)
	}
}
