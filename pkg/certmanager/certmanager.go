package certmanager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// CertManager Структура для работы с сертификатами
// файлы сертификатов должны быть предварительно сгененрированы по пути certsDir
type CertManager struct {
	certsDir string
	certPool *x509.CertPool
}

func NewCertManager(certsDir string) (*CertManager, error) {

	// Загрузка корневого сертификата
	caCert, err := os.ReadFile(certsDir + "/ca.crt")
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Не удалось загрузить CA сертификат: %v", err))
	}

	// Создание пула корневых сертификатов
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("не удалось добавить CA сертификат в пул")
	}

	sm := &CertManager{
		certsDir: certsDir,
		certPool: certPool,
	}

	return sm, nil
}

// GetServerCredentials Получение полномочий для TLS сервера
func (m *CertManager) GetServerCredentials() (*credentials.TransportCredentials, error) {
	// Загрузка сертификата сервера
	serverCert, err := tls.LoadX509KeyPair(m.certsDir+"/server.crt", m.certsDir+"/server.key")
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Не удалось загрузить серверный сертификат: %v", err))
	}

	// Настройка TLS
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    m.certPool,
	})

	return &creds, nil
}

// GetClientCredentials Получение полномочий для TLS клиента
func (m *CertManager) GetClientCredentials() (*credentials.TransportCredentials, error) {
	// Загрузка сертификата клиента
	clientCert, err := tls.LoadX509KeyPair(m.certsDir+"/client.crt", m.certsDir+"/client.key")
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Не удалось загрузить клиентский сертификат: %v", err))
	}

	// Настройка TLS
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      m.certPool,
	})

	return &creds, nil
}
