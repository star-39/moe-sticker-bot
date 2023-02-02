package msbimport

import (
	"sync"

	tele "gopkg.in/telebot.v3"
)

type FuncSendWarning func(tele.Context) error

// Line sticker types
const (
	LINE_STICKER_STATIC       = "line_s"  //普通貼圖
	LINE_STICKER_ANIMATION    = "line_a"  //動態貼圖
	LINE_STICKER_POPUP        = "line_p"  //全螢幕
	LINE_STICKER_POPUP_EFFECT = "line_f"  //特效
	LINE_EMOJI_STATIC         = "line_e"  //表情貼
	LINE_EMOJI_ANIMATION      = "line_i"  //動態表情貼
	LINE_STICKER_MESSAGE      = "line_m"  //訊息
	LINE_STICKER_NAME         = "line_n"  //隨你填
	KAKAO_EMOTICON            = "kakao_e" //KAKAOTALK普通貼圖

	StoreLine  = "line"
	StoreKakao = "kakao"
)

type LineFile struct {
	Wg sync.WaitGroup
	// path of original file
	OriginalFile string
	// path of converted filea
	ConvertedFile string
	// conversion error
	CError error
}

// This is called linedata due to historical reason,
// instead, it handles "import" data, which includes kakao and line so far.
type LineData struct {
	//Waitgroup for when linedata become available.
	Wg sync.WaitGroup
	//Store type
	Store string
	//Store link
	Link string
	//Store links for different langs
	I18nLinks []string
	//Sticker download link, typically ZIP.
	DLink string
	//Sticker download links.
	DLinks []string
	//Sticker file paths.
	Files []*LineFile
	//Sticker category.
	Category   string
	Id         string
	Title      string
	I18nTitles []string
	TitleWg    sync.WaitGroup
	IsAnimated bool
	Amount     int
}

type LineJson struct {
	Name string
	Sku  string
	Url  string
}

type KakaoJsonResult struct {
	//Korean title
	Title string
	//kakao ID
	TitleUrl string
	//PNG urls
	ThumbnailUrls []string
	//??
	TitleImageUrl string
	//Cover image
	TitleDetailUrl string
}

type KakaoJson struct {
	Result KakaoJsonResult
}
