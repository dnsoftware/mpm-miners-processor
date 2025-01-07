#!/bin/bash

# Создаём приватный ключ для корневого сертификата
openssl genrsa -out ca.key 4096

# Генерируем самоподписанный корневой сертификат
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt \
  -subj "/C=RU/ST=State/L=City/O=Organization/OU=OrgUnit/CN=MinerRootCA"