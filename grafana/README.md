# Grafana Dashboard для Ares API

Этот каталог содержит готовый дашборд Grafana для мониторинга Ares API.

## Структура

```
grafana/
├── dashboards/
│   └── ares-api-dashboard.json    # JSON конфигурация дашборда
└── provisioning/
    ├── dashboards/
    │   └── dashboards.yml          # Конфигурация для автозагрузки дашборда
    └── datasources/
        └── datasources.yml         # Конфигурация datasource Prometheus
```

## Автоматическая загрузка

При запуске с Docker Compose и профилем `monitoring`, дашборд автоматически загружается в Grafana:

```bash
docker-compose --profile monitoring up -d
```

После запуска:
1. Откройте Grafana: http://localhost:3000
2. Войдите (логин: `admin`, пароль: `admin`)
3. Дашборд "Ares API Dashboard" будет доступен сразу

## Содержимое дашборда

### Overview (Обзор)
- **Secrets Created (24h)** - количество созданных секретов за 24 часа
- **Secrets Read (24h)** - количество прочитанных секретов за 24 часа
- **Active Secrets** - текущее количество активных секретов (gauge)
- **Secrets Cleaned Up (24h)** - количество удалённых истёкших секретов за 24 часа

### HTTP Metrics
- **HTTP Requests Rate** - частота HTTP запросов по методам и endpoints
- **HTTP Request Duration (p50, p95)** - перцентили длительности запросов
- **HTTP Status Codes** - распределение HTTP статус кодов (2xx, 4xx, 5xx)

### Business Metrics - Secrets
- **Secrets Creation & Read Rate** - скорость создания и чтения секретов
- **Active Secrets Over Time** - история изменения количества активных секретов
- **Failed Read Attempts** - попытки прочитать уже прочитанные или истёкшие секреты
- **Secrets Cleanup Rate** - скорость очистки истёкших секретов

### Errors & Health
- **Encryption/Decryption Errors** - ошибки при шифровании/расшифровке
- **Application Status** - статус приложения (up/down)

## Ручная установка

Если вы хотите импортировать дашборд вручную:

1. Откройте Grafana
2. Перейдите в **Dashboards → Import**
3. Загрузите файл `dashboards/ares-api-dashboard.json`
4. Выберите Prometheus datasource
5. Нажмите **Import**

## Настройка Prometheus Datasource

Если datasource не создался автоматически:

1. Перейдите в **Configuration → Data sources**
2. Нажмите **Add data source**
3. Выберите **Prometheus**
4. Укажите URL: `http://prometheus:9090` (для Docker) или `http://localhost:9090` (для локальной установки)
5. Нажмите **Save & test**

## Кастомизация

Дашборд можно редактировать через интерфейс Grafana:
- Добавлять новые панели
- Изменять запросы
- Настраивать пороговые значения
- Добавлять алерты

После редактирования можно экспортировать дашборд обратно в JSON через:
**Dashboard settings → JSON Model → Copy to Clipboard**

## Метрики

Все метрики собираются с endpoint `/metrics` Ares API.

Полный список метрик доступен в [internal/metrics/metrics.go](../internal/metrics/metrics.go).
