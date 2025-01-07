#!/bin/bash

# Создаём приватный ключ для сервера
openssl genrsa -out server.key 2048

# Создаём запрос на подпись сертификата (CSR)
openssl req -new -key server.key -out server.csr \
  -subj "/C=RU/ST=State/L=City/O=Organization/OU=Server/CN=miners-processor-server"

# Создаём серверный сертификат (Подписываем серверный сертификат с использованием корневого сертификата.)
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 3650 -sha256 -extfile server.ext

# Возможная ошибка: tls: failed to verify certificate: x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs
# означает, что сертификат сервера не содержит IP-адрес 127.0.0.1 в поле Subject Alternative Name (SAN).
# Для современных проверок TLS обязательно наличие SAN, и если вы используете IP-адрес в качестве хоста, его нужно явно указать в SAN.
# В общем в файл server.ext в раздел [alt_names] добавляем все IP и хосты, на которых будет использоваться сертификат