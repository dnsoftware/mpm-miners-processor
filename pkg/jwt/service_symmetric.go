package jwt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClaimsSymmetric Набор утвержданий для симметричных межмикросервисных токенов
type ClaimsSymmetric struct {
	jwt.RegisteredClaims
	ServiceName string `json:"servicename"` // имя сервиса, который делает запрос
}

// ServiceSymmetric -симметричное шифрование (у клиента и сервера один секретный ключ,
// взаимодействуют сервисы из определенного списка)
type ServiceSymmetric struct {
	serviceName        string
	validServicesList  []string      // список названий сервисов, от которых принимаем запросы
	secret             string        // секрет для проверки и генерации подписи
	validityPeriod     time.Duration // Период действия в минутах
	mu                 sync.Mutex
	currentClientToken string // текущий клиентский токен в строковой форме
}

func NewJWTServiceSymmetric(serviceName string, validServicesList []string, secret string) *ServiceSymmetric {
	s := &ServiceSymmetric{
		serviceName:        serviceName,
		validServicesList:  validServicesList,
		secret:             secret,
		validityPeriod:     60, // 60 минут (TODO вынести в конфиг)
		currentClientToken: "",
	}

	return s
}

// GetActualToken Сгенерировать новый токен, если еще не создан или если действие старого скоро истечет
// Если еще не истекло - возвращаем текущий токен
func (s *ServiceSymmetric) GetActualToken() (string, error) {

	if s.currentClientToken == "" {
		// Генерируем новый токен
		return s.generateClientJWT()
	}

	claims, err := s.GetClaims(s.currentClientToken)
	if err != nil {
		return "", err
	}

	// Проверяем, нужен ли новый токен
	// Если текущее время перед временем истечения срока действия - возвращаем текущий токен
	// Добавляем минуту, чтобы немного заранее обновить токен
	if time.Now().Add(1 * time.Minute).Before(claims.ExpiresAt.Time) {
		return s.currentClientToken, nil
	}

	// Генерируем новый токен
	return s.generateClientJWT()
}

// GenerateClientJWT создает JWT токен
func (s *ServiceSymmetric) generateClientJWT() (string, error) {
	claims := ClaimsSymmetric{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.validityPeriod * time.Minute)),
		},
		ServiceName: s.serviceName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return tokenStr, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentClientToken = tokenStr

	return s.currentClientToken, err
}

// GetClaims получение утверждений из токена
func (s *ServiceSymmetric) GetClaims(tokenStr string) (*ClaimsSymmetric, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &ClaimsSymmetric{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ClaimsSymmetric)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// IsServiceValid проверка валидности сервиса от которого идет запрос на выполнение удаленной процедуры
func (s *ServiceSymmetric) IsServiceValid(claims *ClaimsSymmetric) bool {
	for _, v := range s.validServicesList {
		if v == claims.ServiceName {
			return true
		}
	}
	return false
}

// GetValidateInterceptor - gRPC серверный интерсептор для проверки JWT
func (s *ServiceSymmetric) GetValidateInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Извлечение метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("missing metadata")
		}

		// Проверка заголовка авторизации
		authHeader, exists := md["authorization"]
		if !exists || len(authHeader) == 0 {
			return nil, fmt.Errorf("missing authorization token")
		}

		// Извлечение токена
		token := strings.TrimSpace(authHeader[0])

		// Валидация токена
		claims, err := s.GetClaims(token)
		if err != nil {
			return nil, fmt.Errorf("invalid token: %v", err)
		}
		if !s.IsServiceValid(claims) {
			return nil, fmt.Errorf("invalid service: %v", err)
		}

		// Проверяем, истек ли срок действия
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token has expired")
		}

		// Добавление данных из токена в контекст
		ctx = context.WithValue(ctx, "claims", claims)

		// Продолжение выполнения запроса
		return handler(ctx, req)
	}
}

// GetClientInterceptor Unary Interceptor для добавления JWT-токена
func (s *ServiceSymmetric) GetClientInterceptord() grpc.UnaryClientInterceptor {

	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		token, err := s.GetActualToken()
		if err != nil {
			return err
		}

		// Добавляем токен в метаданные
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

		// Выполняем основной запрос
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
