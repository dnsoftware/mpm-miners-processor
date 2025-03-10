package constants

const (
	ProjectRootAnchorFile = ".env"
	AppLogFile            = "app.log"
	TestLogFile           = "test.log"

	StartConfigFilename  = "/startconf.yaml"             // название файла стартового конфига (с доступами к etcd основного конфига)
	LocalConfigPath      = "/config.yaml"                // Путь к локальному файлу конфига (сюда сохраняется удаленный конфиг)
	ServiceConfigPath    = "/service_config/services"    // Папка в etcd где хранятся конфиги микросервисов
	ServiceDiscoveryPath = "/service_discovery/services" // Папка в etcd где хранятся текущие адреса микросервисов

	CaPath      = "/certs/ca.crt"     // путь к корневому сертификату
	PublicPath  = "/certs/client.crt" // путь к сертификату
	PrivatePath = "/certs/client.key" // путь к приватному ключу

)

const (
	ApiBaseUrlGrpc = "grpc" // дополнительный суффикс для идентификации службы в ServiceDiscovery
	ApiBaseUrlRest = "rest" // дополнительный суффикс для идентификации службы в ServiceDiscovery
)

// Postgresql
const (
	QueryDealine = 5 // время в секундах, после которого прерывать контекст выполнения Postgresql запроса
)

const MigrationDir = "migration" // папка с миграциями относительно корня проекта
