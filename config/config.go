package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AppName     string `yaml:"app_name" envconfig:"APP_NAME"    required:"true"`
	AppVersion  string `yaml:"app_version" envconfig:"APP_VERSION" required:"true"`
	PostgresDSN string `yaml:"postgres_dsn" envconfig:"POSTGRES_DSN" required:"true"`
	JWTSecret   string `yaml:"jwt_secret" envconfig:"JWT_SECRET" required:"true"`
	GrpcPort    string `yaml:"grpc_port" envconfig:"GRPC_PORT" required:"true"`
}

func New(filePath string, envFile string) (Config, error) {
	var config Config
	var err error

	// 1. Читаем из config.yaml. Самый низкий приоритет
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		decoder := yaml.NewDecoder(file)
		if decodeErr := decoder.Decode(&config); decodeErr != nil {
			log.Fatalf("Ошибка при чтении config.yaml: %v", decodeErr)
		}
	} else {
		log.Printf("config.yaml не найден, используются значения по умолчанию: %v", err)
	}

	// 2.1 Загрузка переменных окружения из .env
	err = godotenv.Load(envFile)
	if err != nil {
		return config, fmt.Errorf("godotenv.Load: %w", err)
	}

	// 2.2 Переопределяем переменные, полученные из конфиг файла
	err = envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	// 3. Чтение параметров командной строки
	// ... TODO добавить по необходимости

	return config, nil
}
