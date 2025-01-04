package grpc

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
	tctest "github.com/dnsoftware/mpm-miners-processor/test/testcontainers"
)

const bufSize = 1024 * 1024 // Размер буфера для соединений в памяти

var lis *bufconn.Listener

func bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return lis.Dial() // Возвращает соединение внутри процесса
}

func setup(t *testing.T) {
	// Создаем буферизованный listener
	lis = bufconn.Listen(bufSize)
	serverReady := make(chan struct{})

	filePath, err := logger.GetLoggerTestLogPath()
	require.NoError(t, err)
	logger.InitLogger(logger.LogLevelDebug, filePath)

	// Подготовка Postgres контейнера
	ctxPG := context.Background()
	postgresContainer, err := tctest.NewPostgresTestcontainer(t)
	if err != nil {
		t.Fatalf(err.Error())
	}

	dsn, err := postgresContainer.ConnectionString(ctxPG, "sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}

	// Миграции
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	// Укажите путь к миграциям и строку подключения к базе данных
	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir,
		dsn,
	)
	require.NoError(t, err)

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}
	log.Println("Миграции успешно применены")

	// Поднимаем gRPC-сервер в фоновом процессе
	go func() {
		grpcServer := grpc.NewServer()
		minersServer, err := pb.NewGRPCServer(pool)
		require.NoError(t, err)
		pb.RegisterMinersServiceServer(grpcServer, minersServer)
		close(serverReady) // Уведомляем, что сервер готов
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	<-serverReady // Ждем, пока сервер отправит сигнал готовности (вычитываем пустое значение после закрытия канала)
}

func TestGRPCServer(t *testing.T) {

	setup(t)

	// Создаем контекст с тайм-аутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Создаем gRPC соединение через `NewClient`
	conn, err := grpc.DialContext(ctx,
		"bufnet",                          // Адрес символический, используется только для идентификации (т.к. используем bufconn в тестах)
		grpc.WithContextDialer(bufDialer), // Указываем кастомный диалер
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	// Создаем клиента
	client := pb.NewMinersServiceClient(conn)

	// проверяем реальный запрос/ответ
	// Coin
	req := pb.GetCoinIDByNameRequest{Coin: "ALPH"}
	resp2, err := client.GetCoinIDByName(ctx, &req)
	require.NoError(t, err)
	require.Equal(t, resp2.GetId(), int64(4))

	req = pb.GetCoinIDByNameRequest{Coin: "NONAME"}
	resp2, err = client.GetCoinIDByName(ctx, &req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no rows in result set")

	// Wallet
	res, err := client.CreateWallet(ctx, &pb.CreateWalletRequest{
		CoinId:       4,
		Name:         "wallet",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Id)

	res2, err := client.GetWalletIDByName(ctx, &pb.GetWalletIDByNameRequest{
		Wallet:       "wallet",
		CoinId:       4,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res2.Id)

	// Worker
	res3, err := client.CreateWorker(ctx, &pb.CreateWorkerRequest{
		CoinId:       4,
		Workerfull:   "wallet.worker",
		Wallet:       "wallet",
		Worker:       "worker",
		ServerId:     "SERV",
		Ip:           "127.0.0.1",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res3.Id)

	res4, err := client.GetWorkerIDByName(ctx, &pb.GetWorkerIDByNameRequest{
		Workerfull:   "wallet.worker",
		CoinId:       4,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res4.Id)

}
