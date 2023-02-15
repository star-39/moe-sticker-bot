# Moe-Sticker-Bot Import Component (msbimport)

## Description
This package is intended to fetch, parse, download and convert LINE and KakaoTalk Stickers from share link.

It is designed to be able to operate independentaly from moe-sticker-bot core so third party apps can also utilize this package.

此套件或CLI用於解析LINE/Kakaotalk貼圖的分享連結並下載和轉換。

此套件可獨立於moe-sticker-bot使用， 第三方App可以輕鬆利用此套件或CLI處理複雜貼圖。


## CLI Usage/終端機程式使用
Source code of CLI is on: [/moe-sticker-bot/cmd/msbimport](https://github.com/star-39/moe-sticker-bot/tree/master/cmd/msbimport)

Download `msbimport`： 下載`msbimport`： https://github.com/star-39/moe-sticker-bot/releases

```bash
msbimport --help
Usage of ./msbimport:
  -convert
    	Convert to Telegram format(WEBP/WEBM) 轉換為Telegram格式。
  -dir string
    	Where to put sticker files. 指定存放貼圖檔的資料夾。
  -json
    	Output JSON serialized LineData, useful when integrating with other apps. 列印LineData為JSON。
  -link string
    	Import link(LINE/kakao) 分享連結。
  -log_level string
    	Log level (default "debug")
        
msbimport --link https://store.line.me/stickershop/product/27286

msbimport --link https://store.line.me/stickershop/product/27286 --convert --json

```



## API Usage

A typical workflow is to call `parseImportLink` then `prepareImportStickers`.

```
go get -u https://github.com/star-39/moe-sticker-bot
```


```go
import "github.com/star-39/moe-sticker-bot/pkg/msbimport"

//Create a context, which can be used to interrupt the process.
ctx, _ := context.WithCancel(context.Background())

//Create a empty LineData struct pointer.
ld := &msbimport.LineData{}

//LineData will be parsed to ld.
warn, err := msbimport.ParseImportLink("https://store.line.me/stickershop/product/27286", ld)
if err != nil {
    //handle error here.
}
if warn != "" {
    //handle warning message here.
}

err = msbimport.PrepareImportStickers(ctx, ld, "./", false)
if err != nil {
    //handle error here.
}

//If I18n title is needed(LINE only), TitleWg must be waited.
ld.TitleWg.Wait()
println(ld.I18nTitles)

for _, lf := range ld.Files {
    //Each file has its own waitgroup and musted be waited.
    lf.Wg.Wait()
    if lf.CError != nil {
        //hanlde sticker error here.
    }
    println(lf.OriginalFile)
    println(lf.ConvertedFile)
    //...
}

//Your stickers files will appear in the work dir you specified.
```


## License
GPL v3 License.

Source code of this package MUST ALWAYS be disclosed no matter what use case is, 

and source code referring to this package MUST ALSO be discolsed and share the same GPL v3 License.
