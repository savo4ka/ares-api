# Ares API

Backend сервис для безопасной передачи секретов с шифрованием и ограниченным временем жизни.

## Описание

Ares API - это backend приложение на Golang для безопасной передачи конфиденциальной информации (паролей, токенов, учётных данных). Каждый секрет:
- Шифруется с помощью AES-128-CBC
- Имеет ограниченный срок жизни (24, 48 или 72 часа)
- Может быть прочитан только один раз
- Автоматически удаляется после истечения срока

## Технологии

- **Язык**: Go 1.25+
- **База данных**: PostgreSQL
- **Шифрование**: AES-128-CBC (crypto/cipher)
- **HTTP роутер**: gorilla/mux
- **PostgreSQL драйвер**: pgx/v5
- **Метрики**: Prometheus

## Структура проекта

```
ares-api/
├── cmd/
│   └── server/           # Точка входа приложения
│       └── main.go
├── internal/
│   ├── config/          # Конфигурация приложения
│   ├── crypto/          # Сервис шифрования AES-128-CBC
│   ├── database/        # Подключение к PostgreSQL
│   ├── handlers/        # HTTP обработчики
│   ├── metrics/         # Prometheus метрики
│   ├── models/          # Модели данных
│   └── repository/      # Слой работы с БД
├── migrations/          # SQL миграции
└── .env.example        # Пример файла конфигурации
```

## Установка и запуск

### Предварительные требования

- Go 1.25 или выше
- PostgreSQL 12 или выше
- Git

### Шаг 1: Клонирование репозитория

```bash
git clone https://github.com/savo4ka/ares-api.git
cd ares-api
```

### Шаг 2: Установка инструмента миграций

Установите `golang-migrate`:

```bash
# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# macOS
brew install golang-migrate

# Windows (через Scoop)
scoop install migrate

# Или через Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Шаг 3: Настройка базы данных

Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE ares_db;
```

Примените миграции:

```bash
# Применить все миграции
migrate -database "postgresql://postgres:password@localhost:5432/ares_db?sslmode=disable" -path migrations up

# Откатить последнюю миграцию
migrate -database "postgresql://postgres:password@localhost:5432/ares_db?sslmode=disable" -path migrations down 1

# Проверить версию миграции
migrate -database "postgresql://postgres:password@localhost:5432/ares_db?sslmode=disable" -path migrations version
```

### Шаг 4: Настройка конфигурации

Создайте файл `.env` на основе `.env.example`:

```bash
copy .env.example .env
```

Отредактируйте `.env` и укажите свои значения:

```env
SERVER_PORT=8080
DATABASE_URL=postgresql://postgres:password@localhost:5432/ares_db?sslmode=disable
ENCRYPTION_KEY=your16charkey123  # ВАЖНО: Должен быть ровно 16 символов!
ALLOWED_ORIGINS=http://localhost:3000
```

**ВАЖНО:** Для production сгенерируйте уникальный ключ шифрования!

### Шаг 5: Установка зависимостей

```bash
go mod download
```

### Шаг 6: Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустится на `http://localhost:8080`

## API Endpoints

### 1. Создание секрета

**POST** `/api/secrets`

Создаёт новый зашифрованный секрет.

**Request Body:**
```json
{
  "content": "Текст секрета",
  "expiration_hours": 24
}
```

Параметры:
- `content` (string, обязательный) - текст секрета
- `expiration_hours` (int, обязательный) - время жизни: 24, 48 или 72 часа

**Response (201 Created):**
```json
{
  "id": "uuid",
  "url": "http://localhost:8080/secret/uuid",
  "expires_at": "2025-10-31T12:00:00Z"
}
```

### 2. Получение секрета

**GET** `/api/secrets/{id}`

Получает и расшифровывает секрет. **Можно получить только один раз!**

**Response (200 OK):**
```json
{
  "content": "Расшифрованный текст секрета",
  "expires_at": "2025-10-31T12:00:00Z",
  "created_at": "2025-10-30T12:00:00Z"
}
```

**Возможные ошибки:**
- `404 Not Found` - секрет не найден
- `410 Gone` - секрет уже был прочитан или истёк срок действия

### 3. Health Check

**GET** `/health`

Проверка здоровья сервиса.

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

### 4. Метрики Prometheus

**GET** `/metrics`

Эндпоинт для сбора метрик в формате Prometheus.

**Доступные метрики:**

**HTTP метрики:**
- `ares_http_requests_total` - Общее количество HTTP запросов (по методу, пути, статусу)
- `ares_http_request_duration_seconds` - Длительность HTTP запросов в секундах

**Бизнес-метрики секретов:**
- `ares_secrets_created_total` - Количество созданных секретов
- `ares_secrets_read_total` - Количество успешно прочитанных секретов
- `ares_secrets_already_read_total` - Попытки прочитать уже прочитанный секрет
- `ares_secrets_expired_read_total` - Попытки прочитать истекший секрет
- `ares_secrets_cleaned_up_total` - Количество удалённых истекших секретов
- `ares_active_secrets` - Текущее количество активных секретов (gauge)

**Метрики шифрования:**
- `ares_encryption_errors_total` - Ошибки шифрования
- `ares_decryption_errors_total` - Ошибки расшифровки

**Пример использования:**
```bash
curl http://localhost:8080/metrics
```

**Интеграция с Prometheus:**

Добавьте в `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'ares-api'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
```

## Безопасность

### Шифрование

- Используется **AES-128-CBC** с уникальным IV для каждого секрета
- Мастер-ключ хранится в переменной окружения (не в коде!)
- Для каждого секрета генерируется уникальный IV (initialization vector)
- Зашифрованные данные кодируются в base64 для хранения

### Защита данных

- Секрет можно прочитать **только один раз**
- Автоматическое удаление истёкших секретов каждый час
- CORS настраивается через переменную окружения
- Все пароли БД хранятся в `.env` (не коммитится в git)

## Примеры использования

### Создание секрета с помощью curl

```bash
curl -X POST http://localhost:8080/api/secrets \
  -H "Content-Type: application/json" \
  -d "{\"content\":\"Мой секретный пароль\",\"expiration_hours\":24}"
```

### Получение секрета

```bash
curl http://localhost:8080/api/secrets/{id}
```

## Разработка

### Работа с миграциями

```bash
# Создать новую миграцию
migrate create -ext sql -dir migrations -seq название_миграции

# Применить все миграции
migrate -database $DATABASE_URL -path migrations up

# Откатить все миграции
migrate -database $DATABASE_URL -path migrations down

# Откатить N миграций
migrate -database $DATABASE_URL -path migrations down N

# Применить/откатить до конкретной версии
migrate -database $DATABASE_URL -path migrations goto VERSION

# Форсировать версию (использовать осторожно!)
migrate -database $DATABASE_URL -path migrations force VERSION
```

### Запуск с hot-reload (опционально)

Установите `air`:
```bash
go install github.com/cosmtrek/air@latest
```

Запустите:
```bash
air
```

### Сборка бинарника

```bash
go build -o bin/ares-api cmd/server/main.go
```

Запуск:
```bash
./bin/ares-api
```

## Docker

### Запуск с Docker Compose

Самый простой способ запустить весь стек (приложение + PostgreSQL):

```bash
# Запуск всех сервисов
docker-compose up -d

# Запуск с мониторингом (Prometheus + Grafana)
docker-compose --profile monitoring up -d

# Остановка
docker-compose down

# Остановка с удалением volumes
docker-compose down -v
```

После запуска:
- API: http://localhost:8080
- Prometheus: http://localhost:9090 (если включен профиль monitoring)
- Grafana: http://localhost:3000 (если включен профиль monitoring, логин: admin/admin)

### Сборка Docker образа

```bash
# Сборка образа
docker build -t ares-api:latest .

# Запуск контейнера
docker run -d \
  --name ares-api \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/ares_db?sslmode=disable" \
  -e ENCRYPTION_KEY="your16charkey123" \
  -e ALLOWED_ORIGINS="*" \
  ares-api:latest
```

### Использование из GitHub Container Registry

Образы автоматически публикуются в GitHub Container Registry при пуше в main или создании тега:

```bash
# Последняя версия из main
docker pull ghcr.io/savo4ka/ares-api:latest

# Конкретная версия (тег)
docker pull ghcr.io/savo4ka/ares-api:v1.0.0

# Запуск
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="..." \
  -e ENCRYPTION_KEY="..." \
  ghcr.io/savo4ka/ares-api:latest
```

### Применение миграций в Docker

```bash
# Установите golang-migrate в контейнер или используйте отдельный контейнер
docker run --rm \
  --network ares-network \
  -v $(pwd)/migrations:/migrations \
  migrate/migrate \
  -path=/migrations \
  -database "postgresql://postgres:postgres@postgres:5432/ares_db?sslmode=disable" \
  up
```

## Production deployment

### Рекомендации по безопасности

1. **Ключ шифрования**: Сгенерируйте надёжный случайный ключ (16 символов)
   ```bash
   openssl rand -base64 16 | cut -c1-16
   ```

2. **База данных**: Используйте SSL/TLS (`sslmode=require` в DATABASE_URL)

3. **CORS**: Настройте конкретные разрешённые origins вместо `*`

4. **Reverse Proxy**: Используйте nginx/Caddy для HTTPS и rate limiting

5. **Secrets**: Используйте Docker secrets или переменные окружения (не hardcode)

6. **Мониторинг**: Настройте Prometheus + Grafana для отслеживания метрик

### Пример docker-compose для production

```yaml
version: '3.8'

services:
  app:
    image: ghcr.io/savo4ka/ares-api:latest
    restart: always
    environment:
      SERVER_PORT: 8080
      DATABASE_URL: postgresql://user:${DB_PASSWORD}@postgres:5432/ares_db?sslmode=require
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      ALLOWED_ORIGINS: https://yourdomain.com
    depends_on:
      - postgres
    networks:
      - internal

  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ares_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - internal

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - app
    networks:
      - internal

networks:
  internal:

volumes:
  postgres_data:
```

### CI/CD

Проект использует GitHub Actions для автоматической сборки и публикации Docker образов:

- **Push в main**: создаёт образ с тегом `latest`
- **Создание тега `v*`**: создаёт образы с версионными тегами (v1.0.0, v1.0, v1)
- **Pull Request**: собирает образ для тестирования (не публикует)

Workflow автоматически:
- Собирает multi-platform образы (amd64, arm64)
- Публикует в GitHub Container Registry
- Использует build cache для ускорения сборки
- Добавляет метаданные и labels

## Лицензия

MIT

## Контакты

GitHub: [savo4ka/ares-api](https://github.com/savo4ka/ares-api)