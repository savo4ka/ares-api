# Ares API

Backend сервис для безопасной передачи секретов с шифрованием и ограниченным временем жизни.

## Описание

Ares API - это backend приложение на Golang для безопасной передачи конфиденциальной информации (паролей, токенов, учётных данных). Каждый секрет:
- Шифруется с помощью AES-128-CBC
- Имеет ограниченный срок жизни (24, 48 или 72 часа)
- Может быть прочитан только один раз
- Автоматически удаляется после истечения срока

## Технологии

- **Язык**: Go 1.21+
- **База данных**: PostgreSQL
- **Шифрование**: AES-128-CBC (crypto/cipher)
- **HTTP роутер**: gorilla/mux
- **PostgreSQL драйвер**: pgx/v5

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
│   ├── models/          # Модели данных
│   └── repository/      # Слой работы с БД
├── migrations/          # SQL миграции
└── .env.example        # Пример файла конфигурации
```

## Установка и запуск

### Предварительные требования

- Go 1.21 или выше
- PostgreSQL 12 или выше
- Git

### Шаг 1: Клонирование репозитория

```bash
git clone https://github.com/savo4ka/ares-api.git
cd ares-api
```

### Шаг 2: Настройка базы данных

Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE ares_db;
```

Примените миграцию для создания таблицы:

```sql
psql -U postgres -d ares_db -f migrations/001_create_secrets_table.sql
```

### Шаг 3: Настройка конфигурации

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

### Шаг 4: Установка зависимостей

```bash
go mod download
```

### Шаг 5: Запуск сервера

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
go build -o bin/ares-api.exe cmd/server/main.go
```

Запуск:
```bash
.\bin\ares-api.exe
```

## Production deployment

1. Сгенерируйте надёжный ключ шифрования (16 символов)
2. Используйте SSL/TLS для PostgreSQL (`sslmode=require`)
3. Настройте CORS для конкретных доменов
4. Используйте reverse proxy (nginx, Caddy)
5. Настройте мониторинг и логирование

## Лицензия

MIT

## Контакты

GitHub: [savo4ka/ares-api](https://github.com/savo4ka/ares-api)