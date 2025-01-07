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
