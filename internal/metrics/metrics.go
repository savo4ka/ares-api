package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	// Бизнес-метрики секретов
	SecretsCreatedTotal      prometheus.Counter
	SecretsReadTotal         prometheus.Counter
	SecretsAlreadyReadTotal  prometheus.Counter
	SecretsExpiredReadTotal  prometheus.Counter
	SecretsCleanedUpTotal    prometheus.Counter
	ActiveSecretsGauge       prometheus.Gauge

	// Метрики шифрования
	EncryptionErrorsTotal prometheus.Counter
	DecryptionErrorsTotal prometheus.Counter
}

// New создаёт и регистрирует все метрики
func New() *Metrics {
	return &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ares_http_requests_total",
				Help: "Общее количество HTTP запросов",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ares_http_request_duration_seconds",
				Help:    "Длительность HTTP запросов в секундах",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		// Бизнес-метрики секретов
		SecretsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_secrets_created_total",
				Help: "Общее количество созданных секретов",
			},
		),
		SecretsReadTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_secrets_read_total",
				Help: "Общее количество успешно прочитанных секретов",
			},
		),
		SecretsAlreadyReadTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_secrets_already_read_total",
				Help: "Количество попыток прочитать уже прочитанный секрет",
			},
		),
		SecretsExpiredReadTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_secrets_expired_read_total",
				Help: "Количество попыток прочитать истекший секрет",
			},
		),
		SecretsCleanedUpTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_secrets_cleaned_up_total",
				Help: "Общее количество удалённых истекших секретов при cleanup",
			},
		),
		ActiveSecretsGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ares_active_secrets",
				Help: "Текущее количество активных (не прочитанных и не истекших) секретов",
			},
		),

		// Метрики шифрования
		EncryptionErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_encryption_errors_total",
				Help: "Общее количество ошибок шифрования",
			},
		),
		DecryptionErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ares_decryption_errors_total",
				Help: "Общее количество ошибок расшифровки",
			},
		),
	}
}
