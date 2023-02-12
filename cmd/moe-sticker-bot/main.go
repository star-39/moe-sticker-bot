package main

import (
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/core"
)

// Common abbr. in this project:
// S : Sticker
// SS : StickerSet

func main() {
	conf := parseCmdLine()
	core.Init(conf)
}

func parseCmdLine() core.ConfigTemplate {
	var help = flag.Bool("help", false, "Show help")
	var adminUid = flag.Int64("admin_uid", -1, "Admin's UID(optional)")
	var botToken = flag.String("bot_token", "", "Telegram Bot Token")
	var dataDir = flag.String("data_dir", "", "Overwrites the working directory where msb puts data.")
	var webapp = flag.Bool("webapp", false, "Enable WebApp support")
	var webappUrl = flag.String("webapp_url", "", "Public URL to WebApp, HTTPS only")
	// var webappApiUrl = flag.String("webapp_api_url", "", "Public URL to WebApp API server, if unset, same as webapp_url HTTPS only")
	var WebappListenAddr = flag.String("webapp_listen_addr", "", "Webapp API server listen address(IP:PORT)")
	var webappDataDir = flag.String("webapp_data_dir", "", "Where to put webapp data to share with ReactApp ")
	var useDB = flag.Bool("use_db", false, "Use MariaDB")
	var dbAddr = flag.String("db_addr", "", "mariadb address")
	var dbUser = flag.String("db_user", "", "mariadb usernmae")
	var dbPass = flag.String("db_pass", "", "mariadb password")
	var logLevel = flag.String("log_level", "debug", "Log level")
	// var localBotApiAddr = flag.String("local_botapi_addr", "", "Local Bot API Server Address")
	// var localBotApiDir = flag.String("local_botapi_dir", "", "Local Bot API Working directory(Absolute)")
	var webhookPublicAddr = flag.String("webhook_public_addr", "", "Webhook public address(WebhookEndpoint).")
	var webhookListenAddr = flag.String("webhook_listen_addr", "", "Webhook listen address(IP:PORT)")
	// var webhookCert = flag.String("webhook_cert", "", "Certificate for WebHook")
	flag.Parse()
	if *help {
		flag.Usage()
		println("Only --bot_token is required to run.")
		os.Exit(0)
	}

	conf := core.ConfigTemplate{}

	conf.BotToken = *botToken
	if conf.BotToken == "" {
		log.Error("Please set --bot_token")
		log.Error("Use --help to see options.")
		// log.Error("Please note that specifing BOT_TOKEN env var is no longer supported.")
		log.Fatal("No bot token provided!")
	}
	if !strings.Contains(conf.BotToken, ":") {
		log.Fatal("Bad bot token!")
	}

	//Use the second half of bot token as secret_token for webhook.
	conf.WebhookSecretToken = strings.Split(conf.BotToken, ":")[1]

	conf.UseDB = *useDB
	conf.DbAddr = *dbAddr
	conf.DbUser = *dbUser
	conf.DbPass = *dbPass

	conf.WebApp = *webapp
	conf.WebappUrl = *webappUrl
	// Defaults apiUrl to webappUrl
	// if conf.WebappApiUrl == "" {
	conf.WebappApiUrl = *webappUrl
	// } else {
	// 	conf.WebappApiUrl = *webappApiUrl
	// }
	conf.WebappDataDir = *webappDataDir
	conf.WebappListenAddr = *WebappListenAddr

	// conf.LocalBotApiAddr = *localBotApiAddr
	// conf.LocalBotApiDir = *localBotApiDir
	conf.WebhookPublicAddr = *webhookPublicAddr
	conf.WebhookListenAddr = *webhookListenAddr
	// conf.WebhookCert = *webhookCert

	conf.LogLevel = *logLevel

	conf.AdminUid = *adminUid
	conf.DataDir = *dataDir

	return conf
	// core.Config = conf
}
