# Moe-Sticker-Bot Import Component

## Description
This package is intended to fetch, parse, download and convert LINE and KakaoTalk Stickers from share link.

It is designed to be able to operate independentaly from moe-sticker-bot core so third party apps can also utilize this package.

## Usage

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

for _, lf := range ld.Files {
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
