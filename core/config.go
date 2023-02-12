package core

var msbconf ConfigTemplate

type ConfigTemplate struct {
	AdminUid         int64
	DataDir          string
	LogLevel         string
	UseDB            bool
	BotToken         string
	WebApp           bool
	WebappUrl        string
	WebappApiUrl     string
	WebappListenAddr string
	WebappDataDir    string
	DbAddr           string
	DbUser           string
	DbPass           string
	// LocalBotApiAddr   string
	// LocalBotApiDir    string
	WebhookPublicAddr string
	WebhookListenAddr string
	// WebhookCert        string
	WebhookSecretToken string
}
