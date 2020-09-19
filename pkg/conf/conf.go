package conf

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	DBInfo   string
}

// New return new configuration
func New() *Config {
	return &Config{
		Webhook: WebhookCfg{
			WebhookToken:      getEnv("FBWEBHOOK_TOKEN", ""),
			MessengerEndpoint: getEnv("FBWEBHOOK_MSG_ENDPOINT", ""),
		},
		FBSecret: FacebookSecret{
			PakeToken: getEnv("FBSECRET_PAGE_TOKEN", ""),
		},
		DBInfo: getDBSecret(),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) (value string) {
	var exist bool
	if value, exist = os.LookupEnv(key); exist {
		return value
	}

	if defaultVal == "" && exist == false {
		panic("Can't get environment var: " + key)
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

func getDBSecret() string {
	DBConfig, err := url.Parse(getEnv("DBSECRET_URL", ""))

	if err != nil {
		panic(err)
	}

	password, _ := DBConfig.User.Password()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		DBConfig.Hostname(), DBConfig.Port(), DBConfig.User.Username(), password, strings.Trim(DBConfig.Path, "/"))
	return psqlInfo
}
