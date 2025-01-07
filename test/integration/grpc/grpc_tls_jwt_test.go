package grpc

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc/proto"
	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
	jwt2 "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
	tctest "github.com/dnsoftware/mpm-miners-processor/test/testcontainers"
)

func TestTLSJWTTest(t *testing.T) {

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

	jwt := jwt2.NewJWTServiceSymmetric("normalizer", []string{"normalizer"}, "jwtsecret")

	// Поднимаем gRPC-сервер в фоновом процессе
	go func() {
		interceptor := jwt.GetValidateInterceptor()
		grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
		minersServer, err := pb.NewGRPCServer(pool)
		require.NoError(t, err)
		proto.RegisterMinersServiceServer(grpcServer, minersServer)
		close(serverReady) // Уведомляем, что сервер готов
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	<-serverReady // Ждем, пока сервер отправит сигнал готовности (вычитываем пустое значение после закрытия канала)

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
	client := proto.NewMinersServiceClient(conn)

	// Генерация токена
	token, err := jwt.GenerateJWT()
	if err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}

	// Добавление токена в метаданные
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	// проверяем реальный запрос/ответ
	// Coin
	req := proto.GetCoinIDByNameRequest{Coin: "ALPH"}
	resp2, err := client.GetCoinIDByName(ctx, &req)
	require.NoError(t, err)
	require.Equal(t, resp2.GetId(), int64(4))

	req = proto.GetCoinIDByNameRequest{Coin: "NONAME"}
	resp2, err = client.GetCoinIDByName(ctx, &req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no rows in result set")

	// Wallet
	res, err := client.CreateWallet(ctx, &proto.CreateWalletRequest{
		CoinId:       4,
		Name:         "wallet",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Id)

	res2, err := client.GetWalletIDByName(ctx, &proto.GetWalletIDByNameRequest{
		Wallet:       "wallet",
		CoinId:       4,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res2.Id)

	// Проверка на повторную вставку
	resD, err := client.CreateWallet(ctx, &proto.CreateWalletRequest{
		CoinId:       4,
		Name:         "wallet",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, res.Id, resD.Id)

	// Worker
	res3, err := client.CreateWorker(ctx, &proto.CreateWorkerRequest{
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

	res4, err := client.GetWorkerIDByName(ctx, &proto.GetWorkerIDByNameRequest{
		Workerfull:   "wallet.worker",
		CoinId:       4,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res4.Id)

	// Проверка на повторную вставку
	res5, err := client.CreateWorker(ctx, &proto.CreateWorkerRequest{
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
	require.Equal(t, res3.Id, res5.Id)

}
