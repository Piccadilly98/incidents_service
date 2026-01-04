package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	GoModName = "go.mod"
	GoSumName = "go.sum"
	NameEnv   = ".env"

	EnvNameWebHookURL    = "WEBHOOK_URL"
	EnvNameWebhookMethod = "WEBHOOK_METHOD"
	EnvNameServerAddr    = "SERVER_ADDR"
	EnvNameServerPort    = "SERVER_PORT"

	EnvNameDbName     = "DB_NAME"
	EnvNameDbSSlMode  = "DB_SSLMODE"
	EnvNameDbPort     = "DB_PORT"
	EnvNameDbHost     = "DB_HOST"
	EnvNameDbPassword = "DB_PASSWORD"
	EnvNameDbUser     = "DB_USER"
)

const (
	DefaultDbHost        = "localhost"
	DefaultDbPort        = "5432"
	DefaultDbSSLMode     = "disable"
	DefaultServerAddress = "localhost"
	DefaultServerPort    = "8080"
	DefaultWebhookURL    = "hhtp://localhost:9090"
	DefaultWebhookMethod = "POST"
)

type Config struct {
	ConnectionStr string
	WebhookURL    string
	WebhookMethod string
	ServerAddr    string
	ServerPort    string
	DefaultRadius int
	MaxRadius     int
}

func NewConfig(envCfg bool) (*Config, error) {
	if envCfg {
		if err := godotenv.Load(); err != nil {
			path, err := findEnv()
			if err != nil {
				return nil, err
			}

			err = godotenv.Load(path)
			if err != nil {
				return nil, err
			}
		}
	}
	serverAddr := os.Getenv(EnvNameServerAddr)
	if serverAddr == "" {
		serverAddr = DefaultServerAddress
	}
	servPort, err := validationPort(EnvNameServerPort, DefaultServerPort)
	if err != nil {
		return nil, err
	}
	nameDb := os.Getenv(EnvNameDbName)
	if nameDb == "" {
		return nil, fmt.Errorf("db_name cannot be empty")
	}
	dbSsl := os.Getenv(EnvNameDbSSlMode)
	if dbSsl == "" {
		dbSsl = DefaultDbSSLMode
	}

	dbHost := os.Getenv(EnvNameDbHost)
	if dbHost == "" {
		dbHost = DefaultDbHost
	}
	dbPort, err := validationPort(EnvNameDbPort, DefaultDbPort)
	if err != nil {
		return nil, err
	}
	dbUser := os.Getenv(EnvNameDbUser)
	if dbUser == "" {
		return nil, fmt.Errorf("db_user cannot be empty")
	}

	dbPassword := os.Getenv(EnvNameDbPassword)
	if dbPassword == "" {
		return nil, fmt.Errorf("db_password cannot be empty")
	}
	webhookURL := os.Getenv(EnvNameWebHookURL)
	if webhookURL == "" {
		webhookURL = DefaultWebhookURL
	}
	webhookMethod := os.Getenv(EnvNameWebhookMethod)
	if webhookMethod != http.MethodPost && webhookMethod != http.MethodGet {
		log.Printf("invalid WEBHOOK_METHOD if env: <%s>, change to defaul: %s\n", webhookMethod, DefaultWebhookMethod)
		webhookMethod = DefaultWebhookMethod
	}
	conf := &Config{
		ConnectionStr: fmt.Sprintf("user=%s port=%s password=%s dbname=%s host=%s sslmode=%s", dbUser, dbPort, dbPassword, nameDb, dbHost, dbSsl),
		ServerAddr:    serverAddr,
		ServerPort:    servPort,
		WebhookURL:    webhookURL,
		WebhookMethod: webhookMethod,
		// убрать проверку радиуса из бд + добавить в конфиг
		DefaultRadius: 5000,
		MaxRadius:     50000,
	}
	return conf, nil
}

func findEnv() (string, error) {
	path := "./"
	for {
		files, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		for _, enrty := range files {
			if enrty.Name() == GoModName || enrty.Name() == GoSumName {
				res := path + NameEnv
				_, err := os.Open(res)
				if err != nil {
					return "", fmt.Errorf("no .env file")
				}
				return res, nil
			}
		}
		path += "../"
	}
}

func validationPort(key, defaultValue string) (string, error) {
	port := os.Getenv(key)
	if port == "" {
		port = defaultValue
	}
	port = strings.TrimSpace(port)
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid port '%s': %w", port, err)
	}

	if portNum <= 0 || portNum > 65535 {
		return "", fmt.Errorf("port %d out of range (1-65535)", portNum)
	}

	return port, nil
}
