package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/dnsoftware/mpmslib/pkg/servicediscovery"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/dnsoftware/mpm-miners-processor/config"
	pb "github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc/proto"
	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
	"github.com/dnsoftware/mpm-miners-processor/pkg/certmanager"
	jwtauth "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
)

type Dependencies struct {
}

func Run(ctx context.Context, cfg config.Config) {
	var deps Dependencies
	_ = deps

	/********* Инициализация трассировщика **********/
	/********* КОНЕЦ Инициализация трассировщика **********/

	pool, err := pgxpool.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		logger.Log().Fatal("Base path error: " + err.Error())
	}

	etcdConf, err := servicediscovery.NewEtcdConfig(servicediscovery.EtcdConfig{
		Nodes:       strings.Split(cfg.EtcdConfig.Endpoints, ","),
		Username:    cfg.EtcdConfig.Username,
		Password:    cfg.EtcdConfig.Password,
		CertCaPath:  basePath + constants.CaPath,
		CertPath:    basePath + constants.PublicPath,
		CertKeyPath: basePath + constants.PrivatePath,
	})
	if err != nil {
		log.Fatalf("NewEtcdConfig error: %s", err.Error())
	}

	// Регистрируемся в ServiceDiscovery
	serviceKey := cfg.AppID + ":" + constants.ApiBaseUrlGrpc
	serviceAddr := cfg.ApiBaseUrls.Grps
	sd, err := servicediscovery.NewServiceDiscovery(*etcdConf, constants.ServiceDiscoveryPath, serviceKey, serviceAddr, 5, 10)
	if err != nil {
		log.Fatalf("NewServiceDiscovery error: %s", err.Error())
	}

	err = sd.RegisterService(cfg.AppID+":"+constants.ApiBaseUrlRest, cfg.ApiBaseUrls.Rest)
	if err != nil {
		log.Fatalf("Rest service register error: %s", err.Error())
	}

	sd.WaitDependencies(cfg.Dependencies)
	baseUrls, err := sd.DiscoverAllServices()
	if err != nil {
		log.Fatalf("DiscoverAllServices error: %s", err.Error())
	}
	_ = baseUrls
	logger.Log().Info("All services discovered")

	// инициализируем BaseURLs доступа к API внешних сервисов
	cfg.GRPCConfig.SharesProcessor = cfg.ServiceDiscoveryList[cfg.GRPCConfig.SharesProcessor]

	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir,
		cfg.PostgresDSN,
	)
	if err != nil {
		logger.Log().Fatal("Bad migration: " + err.Error())
	}

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}

	jwt := jwtauth.NewJWTServiceSymmetric(cfg.JWTServiceName, cfg.JWTValidServices, cfg.JWTSecret, 60)

	certManager, err := certmanager.NewCertManager(basePath + "/certs")
	if err != nil {
		logger.Log().Fatal("NewCertManager error: " + err.Error())
	}

	// Создаем gRPC-сервер
	serverCreds, err := certManager.GetServerCredentials()
	interceptor := jwt.GetValidateInterceptor()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor), grpc.Creds(*serverCreds))
	minersServer, err := pb.NewGRPCServer(pool)
	if err != nil {
		logger.Log().Fatal("Error create NewGRPCServer: " + err.Error())
	}

	// Регистрируем сервис
	proto.RegisterMinersServiceServer(grpcServer, minersServer)

	// Запускаем сервер на определенном порту
	addrParts := strings.Split(cfg.ApiBaseUrls.Grps, ":")
	lis, err := net.Listen("tcp", ":"+addrParts[1])
	if err != nil {
		logger.Log().Fatal(fmt.Sprintf("Failed to listen: %v", err))
	}

	// Поднимаем gRPC-сервер в фоновом процессе
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log().Fatal(fmt.Sprintf("Server exited with error: %v", err))
		}
	}()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down gRPC server...")

	// Останавливаем сервер
	grpcServer.GracefulStop()
	logger.Log().Info("gRPC server stopped")
}
