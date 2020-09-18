package conf

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// init is invoked before main()
func init() {
	if err := godotenv.Load(); err != nil {
		util.Danger("Can't load env file")
	}
}

// WebhookCfg for facebook webhook
type WebhookCfg struct {
	WebhookToken      string
	MessengerEndpoint string
}

// FacebookSecret include token and app secret facebook provide
type FacebookSecret struct {
	PakeToken string
}

// WorkerData for workerpool configuration
type WorkerData struct {
	WokerNum string
	Timeout  string
}

// Config main struct for get config from env
type Config struct {
	Webhook  WebhookCfg
	FBSecret FacebookSecret
}

func New() *Config {
	return &Config{
		Webhook: WebhookCfg{
			WebhookToken:      getEnv("WEBHOOK_TOKEN", ""),
			MessengerEndpoint: getEnv("MESSENGER_ENDPOINT", ""),
		},
		FBSecret: FacebookSecret{
			PakeToken: getEnv("PAGE_TOKEN", ""),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}
