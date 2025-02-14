package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Etcd struct {
	Endpoints string
	Username  string
	Password  string
}

type ApiBaseUrls struct {
	Grps string `yaml:"grpc"` // host:port для доступа к gRPC API (пустая строка, если нет)
	Rest string `yaml:"rest"` // host:port для доступа к REST API (пустая строка, если нет)
}

type GRPCConfig struct {
	SharesProcessor string `yaml:"shares_processor"` // ServiceDiscovery ID для адреса сервиса процессинга шар
}

type Config struct {
	AppID                string
	ApiBaseUrls          ApiBaseUrls `yaml:"api_base_urls"`
	EtcdConfig           Etcd
	AppName              string            `yaml:"app_name" envconfig:"APP_NAME"    required:"false"`
	AppVersion           string            `yaml:"app_version" envconfig:"APP_VERSION" required:"false"`
	Dependencies         []string          `yaml:"dependencies"` // Зависимости от других микросервисов (будет ожидать их запуска, отслеживание через Service Discovery)
	ServiceDiscoveryList map[string]string // список текущих сервисов из Service Discovery
	PostgresDSN          string            `yaml:"postgres_dsn" envconfig:"POSTGRES_DSN" required:"false"`
	//GrpcPort             string            `yaml:"grpc_port" envconfig:"GRPC_PORT" required:"false"`
	JWTServiceName   string     `yaml:"jwt_service_name" envconfig:"JWT_SERVICE_NAME" required:"false"`     // Название сервиса (для сверки с JWTValidServices при авторизаии)
	JWTSecret        string     `yaml:"jwt_secret" envconfig:"JWT_SECRET" required:"false"`                 // JWT секрет
	JWTValidServices []string   `yaml:"jwt_valid_services" envconfig:"JWT_VALID_SERVICES" required:"false"` // список микросервисов (через запятую), которым разрешен доступ
	GRPCConfig       GRPCConfig `yaml:"grpc"`
}

func New(filePath string, envFile string) (Config, error) {
	var config Config
	var err error

	config.ServiceDiscoveryList = make(map[string]string)

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
