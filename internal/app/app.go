package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-miners-processor/config"
	pb "github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc"
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

	// Создаем gRPC-сервер
	grpcServer := grpc.NewServer()
	minersServer, err := pb.NewGRPCServer(pool)
	if err != nil {
		logger.Log().Fatal("Error create NewGRPCServer: " + err.Error())
	}

	// Регистрируем сервис
	pb.RegisterMinersServiceServer(grpcServer, minersServer)

	// Запускаем сервер на определенном порту
	lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
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
