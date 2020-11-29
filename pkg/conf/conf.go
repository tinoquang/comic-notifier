package conf

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tinoquang/comic-notifier/pkg/logging"
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
	ClientID     string
}

// JWT config info
type JWT struct {
	Issuer    string
	Audience  string
	SecretKey string
}

// Config main struct for get config from env
type Config struct {
	Port       string
	Host       string
	Webhook    WebhookCfg
	FBSecret   FacebookSecret
	DBInfo     string
	WrkDat     WorkerData
	Imgur      Imgur
	JWT        JWT
	CtxTimeout int
}

// New return new configuration
func New(path string) *Config {

	if err := godotenv.Load(path + ".env"); err != nil {
		logging.Danger("Can't load env file, err:", err)
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
		DBInfo: getDBSecret(),
		WrkDat: WorkerData{
			WorkerNum: getEnvAsInt("WORKER_NUM", 10),
			Timeout:   getEnvAsInt("WORKER_TIMEOUT", 30),
		},
		Imgur: Imgur{
			Endpoint:     getEnv("IMGUR_ENDPOINT", ""),
			AccessToken:  getEnv("IMGUR_ACCESS_TOKEN", ""),
			RefreshToken: getEnv("IMGUR_REFRESH_TOKEN", ""),
			ClientID:     getEnv("IMGUR_CLIENT_ID", ""),
		},
		JWT: JWT{
			SecretKey: getEnv("JWT_SECRET", ""),
			Issuer:    getEnv("JWT_ISSUER", ""),
			Audience:  getEnv("JWT_AUDIENCE", ""),
		},
		Port:       getEnv("PORT", ""),
		Host:       getEnv("HOST", ""),
		CtxTimeout: getEnvAsInt("CTX_TIMEOUT", 15),
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
	DBConfig, err := url.Parse(getEnv("DATABASE_URL", ""))

	sslMode := getEnv("SSLMODE", "require")

	if err != nil {
		panic(err)
	}

	password, _ := DBConfig.User.Password()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		DBConfig.Hostname(), DBConfig.Port(), DBConfig.User.Username(), password, strings.Trim(DBConfig.Path, "/"), sslMode)
	return psqlInfo
}
