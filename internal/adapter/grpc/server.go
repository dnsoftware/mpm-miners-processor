package grpc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dnsoftware/mpm-miners-processor/internal/adapter/grpc/proto"
)

type GRPCServer struct {
	proto.UnimplementedMinersServiceServer
	pool *pgxpool.Pool
}

func NewGRPCServer(pool *pgxpool.Pool) (*GRPCServer, error) {
	s := &GRPCServer{
		pool: pool,
	}

	return s, nil
}

func (s *GRPCServer) GetCoinIDByName(ctx context.Context, req *proto.GetCoinIDByNameRequest) (*proto.GetCoinIDByNameResponse, error) {

	var id int64

	err := s.pool.QueryRow(ctx, `SELECT id FROM coins WHERE symbol = $1`, req.Coin).Scan(&id)

	// Ошибка
	if err != nil {
		// Возвращаем ошибку
		st := status.New(codes.Internal, err.Error())
		// Добавляем дополнительные детали, если нужно (тут сделано для примера, как это работает)
		detail := &proto.MPError{
			Method:      "GetCoinIDByName",
			Description: "QueryRow error: " + err.Error(),
		}
		st, _ = st.WithDetails(detail)

		return nil, st.Err() // возвращаем ошибку клиенту
	}

	resp := &proto.GetCoinIDByNameResponse{
		Id: id,
	}

	return resp, nil
}

func (s *GRPCServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	var newID int64

	// Проверка на существование
	check, err := s.GetWalletIDByName(ctx, &proto.GetWalletIDByNameRequest{
		Wallet:       req.Name,
		CoinId:       req.CoinId,
		RewardMethod: req.RewardMethod,
	})
	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}
	if check.Id > 0 {
		return &proto.CreateWalletResponse{Id: check.Id}, nil
	}

	// Вставка новой записи
	err = s.pool.QueryRow(ctx, `INSERT INTO wallets (coin_id, name, is_solo, reward_method) 
			VALUES ($1, $2, $3, $4) RETURNING id`,
		req.CoinId, req.Name, req.IsSolo, req.RewardMethod).Scan(&newID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	resp := &proto.CreateWalletResponse{
		Id: newID,
	}

	return resp, nil
}

func (s *GRPCServer) CreateWorker(ctx context.Context, req *proto.CreateWorkerRequest) (*proto.CreateWorkerResponse, error) {
	created_at := time.Now().Format("2006-01-02 15:04:05.000")
	updated_at := time.Now().Format("2006-01-02 15:04:05.000")

	var newID int64

	// Проверка на существование
	check, err := s.GetWorkerIDByName(ctx, &proto.GetWorkerIDByNameRequest{
		Workerfull:   req.Workerfull,
		CoinId:       req.CoinId,
		RewardMethod: req.RewardMethod,
	})
	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}
	if check.Id > 0 {
		return &proto.CreateWorkerResponse{Id: check.Id}, nil
	}

	// Вставка новой записи
	err = s.pool.QueryRow(ctx, `INSERT INTO workers (coin_id, workerfull, wallet, worker, server_id, created_at, updated_at, reward_method) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		req.CoinId, req.Workerfull, req.Wallet, req.Worker, req.ServerId, created_at, updated_at, req.RewardMethod).Scan(&newID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	resp := &proto.CreateWorkerResponse{
		Id: newID,
	}

	return resp, nil
}

func (s *GRPCServer) GetWalletIDByName(ctx context.Context, req *proto.GetWalletIDByNameRequest) (*proto.GetWalletIDByNameResponse, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM wallets WHERE name = $1 AND coin_id = $2 AND reward_method = $3`,
		req.Wallet, req.CoinId, req.RewardMethod).Scan(&id)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Если нет записей
			return &proto.GetWalletIDByNameResponse{
				Id: 0,
			}, nil
		} else {
			// Обработка других ошибок
			st := status.New(codes.Internal, err.Error())
			return &proto.GetWalletIDByNameResponse{
				Id: 0,
			}, st.Err()
		}
	}

	return &proto.GetWalletIDByNameResponse{
		Id: id,
	}, nil

}

func (s *GRPCServer) GetWorkerIDByName(ctx context.Context, req *proto.GetWorkerIDByNameRequest) (*proto.GetWorkerIDByNameResponse, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM workers WHERE workerfull = $1 AND coin_id = $2 AND reward_method = $3`,
		req.Workerfull, req.CoinId, req.RewardMethod).Scan(&id)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Если нет записей
			return &proto.GetWorkerIDByNameResponse{Id: 0}, nil
		} else {
			// Обработка других ошибок
			st := status.New(codes.Internal, err.Error())
			return &proto.GetWorkerIDByNameResponse{Id: 0}, st.Err()
		}
	}

	return &proto.GetWorkerIDByNameResponse{Id: id}, nil

}
