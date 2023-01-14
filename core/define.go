package core

import (
	"context"
	"sync"

	"github.com/go-co-op/gocron"
	"github.com/panjf2000/ants/v2"
	tele "gopkg.in/telebot.v3"
)

var BOT_VERSION = "2.0.1-GO"
var DB_VER string = "1"

// The telegram bot.
var b *tele.Bot
var cronScheduler *gocron.Scheduler

var dataDir string
var botName string

var downloadQueue DownloadQueue
var webAppSSAuthList WebAppQIDAuthList
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
var LINK_KAKAO = "kakao.com"
var LINK_IMPORT = "IMPORT"

var BSDTAR_BIN string
var CONVERT_BIN string
var CONVERT_ARGS []string

var CB_DN_WHOLE = "dall"
var CB_DN_SINGLE = "dsingle"
var CB_OK_IMPORT = "yesimport"
var CB_OK_DN = "yesd"
var CB_BYE = "bye"
var CB_MANAGE = "manage"
var CB_DONE_ADDING = "done"
var CB_YES = "yes"
var CB_NO = "no"
var CB_DEFAULT_TITLE = "titledefault"
var CB_EXPORT_WA = "exportwa"

var ST_WAIT_WEBAPP = "waitWebApp"
var ST_PROCESSING = "process"

type LineStickerQ struct {
	Line_id   string
	Line_link string
	Tg_id     string
	Tg_title  string
	Ae        bool
}

type UserStickerQ struct {
	tg_id     string
	tg_title  string
	timestamp int64
}

type StickerFile struct {
	wg sync.WaitGroup
	// path of original file
	oPath string
	// path of converted file
	cPath  string
	cError error
}

type StickerData struct {
	id string
	// link     string
	title      string
	emojis     []string
	sticker    *tele.Sticker
	stickers   []*StickerFile
	stickerSet *tele.StickerSet
	// Currently only for WebApp
	sDnObjects []*StickerDownloadObject
	isVideo    bool
	pos        int
	// amount of local files
	lAmount int
	// amount on cloud
	cAmount int
}

type LineData struct {
	store      string
	link       string
	i18nLinks  []string
	dLink      string
	dLinks     []string
	files      []string
	category   string
	id         string
	title      string
	i18nTitles []string
	titleWg    sync.WaitGroup
	isAnimated bool
	amount     int
}

type LineJson struct {
	Name string
	Sku  string
	Url  string
}

type KakaoJsonResult struct {
	Title         string
	TitleUrl      string
	ThumbnailUrls []string
}

type KakaoJson struct {
	Result KakaoJsonResult
}

type WebAppUser struct {
	Id            int
	Is_bot        bool
	First_name    string
	Last_name     string
	Username      string
	Language_code string
	Is_premium    bool
	Photo_url     string
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

	state            string
	sessionID        string
	workDir          string
	command          string
	progress         string
	progressMsg      *tele.Message
	lineData         *LineData
	stickerData      *StickerData
	webAppUser       *WebAppUser
	webAppQID        string
	webAppWorkerPool *ants.PoolWithFunc
	lastContext      tele.Context
}

type DownloadQueue struct {
	mu sync.Mutex
	ss map[string]bool
}

type WebAppQIDAuthList struct {
	mu sync.Mutex
	sa map[string]*WebAppQIDAuthObject
}

type WebAppQIDAuthObject struct {
	sn string
	dt int64
}

type Users struct {
	mu   sync.Mutex
	data map[int64]*UserData
}

type StickerDownloadObject struct {
	wg      sync.WaitGroup
	bot     *tele.Bot
	sticker tele.Sticker
	dest    string
	//Convert to conventional format?
	needConvert bool
	//Shrink oversized GIF?
	shrinkGif bool
	//Sticker is for WebApp?
	forWebApp bool
	//need HQ animated sticker for WhatsApp
	webAppHQ bool
	//need 96px PNG thumb for WhatsApp
	webAppThumb bool
	/*
		Following fields are yielded by worker after wg is done.
	*/
	//Original sticker file downloaded.
	of string
	//Converted sticker file.
	cf string
	//Returned error.
	err error
}

type StickerMoveObject struct {
	wg       sync.WaitGroup
	err      error
	sd       *StickerData
	oldIndex int
	newIndex int
}

func (ud *UserData) udSetState(state string) {
	ud.state = state
}
