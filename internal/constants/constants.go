package constants

const (
	ProjectRootAnchorFile = ".env"
	AppLogFile            = "app.log"
	TestLogFile           = "test.log"
)

// Postgresql
const (
	QueryDealine = 5 // время в секундах, после которого прерывать контекст выполнения Postgresql запроса
)

const MigrationDir = "migration" // папка с миграциями относительно корня проекта
