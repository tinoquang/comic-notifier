package conf

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// WebhookCfg for facebook webhook
type WebhookCfg struct {
	WebhookToken  string
	GraphEndpoint string
}

// FacebookSecret include token and app secret facebook provide
type FacebookSecret struct {
	PakeToken string
	AppID     string
	AppSecret string
	AppToken  string
}

// WorkerData for workerpool configuration
type WorkerData struct {
	WorkerNum int
	Timeout   int
}

// Imgur token info
type Imgur struct {
	Endpoint     string
	AccessToken  string
	RefreshToken string
}

// Config main struct for get config from env
type Config struct {
	Webhook     WebhookCfg
	FBSecret    FacebookSecret
	DBInfo      string
	PageSupport *model.PageList
	WrkDat      WorkerData
	Imgur       Imgur
}

// New return new configuration
func New() *Config {

	if err := godotenv.Load(".env"); err != nil {
		util.Danger("Can't load env file")
	}

	return &Config{
		Webhook: WebhookCfg{
			WebhookToken:  getEnv("FBWEBHOOK_TOKEN", ""),
			GraphEndpoint: getEnv("FBWEBHOOK_GRAPH_ENDPOINT", ""),
		},
		FBSecret: FacebookSecret{
			PakeToken: getEnv("FBSECRET_PAGE_TOKEN", ""),
			AppID:     getEnv("FBSECRET_APP_ID", ""),
			AppSecret: getEnv("FBSECRET_APP_SECRET", ""),
			AppToken:  getEnv("FBSECRET_APP_TOKEN", ""),
		},
		DBInfo:      getDBSecret(),
		PageSupport: getPageSupport(),
		WrkDat: WorkerData{
			WorkerNum: getEnvAsInt("WORKER_NUM", 10),
			Timeout:   getEnvAsInt("WORKER_TIMEOUT", 30),
		},
		Imgur: Imgur{
			Endpoint:     getEnv("IMGUR_ENDPOINT", ""),
			AccessToken:  getEnv("IMGUR_ACCESS_TOKEN", ""),
			RefreshToken: getEnv("IMGUR_REFRESH_TOKEN", ""),
		},
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

func getPageSupport() *model.PageList {

	jsonFile, err := os.Open("./pkg/conf/page_support.json")
	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	pages := new(model.PageList)
	err = decoder.Decode(pages)
	if err != nil {
		panic(err)
	}

	return pages
}
