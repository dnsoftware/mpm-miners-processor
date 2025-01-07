### Генерация TLS сертификатов со своим центром сертицикации

смотри описание https://laradrom.ru/languages/golang/golang-generacziya-sertifikatov-s-sobstvennym-kornevym-sertifikatom-dlya-ispolzovaniya-mtls-dlya-raboty-mikroservisov/

--------

**generate_ca.sh** - генерация самоподписанного корневого сертификата

**ca.key**: закрытый ключ корневого сертификата.

**ca.crt**: самоподписанный корневой сертификат.

---

**generate_server.sh** - генерация серверного сертификата

**server.key**: приватный ключ сервера.

**server.csr**: запрос на подпись сертификата.

**server.crt**: подписанный сертификат сервера.

---

**generate_client.sh** - генерация клиентского сертификата

**client.key**: приватный ключ сервера.

**client.csr**: запрос на подпись сертификата.

**client.crt**: подписанный сертификат сервера.

---

# Возможная ошибка: tls: failed to verify certificate: x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs
# означает, что сертификат сервера не содержит IP-адрес 127.0.0.1 в поле Subject Alternative Name (SAN).
# Для современных проверок TLS обязательно наличие SAN, и если вы используете IP-адрес в качестве хоста, его нужно явно указать в SAN.
# В общем в файл server.ext в раздел [alt_names] добавляем все IP и хосты, на которых будет использоваться сертификат
