package config

var Config ConfigTemplate

type ConfigTemplate struct {
	LogLevel         string
	UseDB            bool
	BotToken         string
	WebApp           bool
	WebappUrl        string
	WebappApiUrl     string
	WebappListenAddr string
	WebappDataDir    string
	WebappCert       string
	WebappPrivkey    string
	DbAddr           string
	DbUser           string
	DbPass           string
}
