package main

import (
	"context"
	"sync"

	"github.com/panjf2000/ants/v2"
	tele "github.com/star-39/telebot"
)

// Workers pool for converting webm
var wpConvertWebm *ants.PoolWithFunc
var dataDir string
var botName string
var botVersion = "1.0-RC5-GO"
var users Users

// Line sticker types
var LINE_STICKER_STATIC = "line_s"       //普通貼圖
var LINE_STICKER_ANIMATION = "line_a"    //動態貼圖
var LINE_STICKER_POPUP = "line_p"        //全螢幕
var LINE_STICKER_POPUP_EFFECT = "line_f" //特效
var LINE_EMOJI_STATIC = "line_e"         //表情貼
var LINE_EMOJI_ANIMATION = "line_i"      //動態表情貼
var LINE_STICKER_MESSAGE = "line_m"      //訊息
var LINE_STICKER_NAME = "line_n"         //隨你填
var KAKAO_EMOTICON = "kakao_e"           //KAKAOTALK普通貼圖

var LINK_TG = "t.me"
var LINK_LINE = "line.me"

var BSDTAR_BIN string
var CONVERT_BIN string
var CONVERT_ARGS []string

type LineStickerQ struct {
	tg_id    string
	tg_title string
	ae       bool
}

type StickerFile struct {
	wg sync.WaitGroup
	// path of original file
	oPath string
	// path of converted file
	cPath   string
	cError  error
	onCloud bool
}

type StickerData struct {
	id       string
	link     string
	title    string
	emojis   []string
	sticker  *tele.Sticker
	stickers []*StickerFile
	isVideo  bool
	isTGS    bool
	pos      int
	upCount  int
	// amount of local files
	lAmount int
	// amount on cloud
	cAmount int
}

type LineData struct {
	link       string
	dLink      string
	files      []string
	category   string
	id         string
	title      string
	isAnimated bool
	amount     int
}

type UserData struct {
	// udWg should be used for time consuming works.
	// When user signals a termination of goroutine,
	// we MUST wait for this wg to Done.
	udWg sync.WaitGroup
	// wg is a generic waitgroup, for internal use.
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	state       string
	userDir     string
	command     string
	progress    string
	progressMsg *tele.Message
	lineData    *LineData
	stickerData *StickerData
}

type Users struct {
	mu   sync.Mutex
	data map[int64]*UserData
}
