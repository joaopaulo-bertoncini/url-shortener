package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Contador de requisições de encurtamento
	ShortenCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_shorten_requests_total",
			Help: "Total number of shorten URL requests",
		},
	)

	// Contador de requisições de redirecionamento
	RedirectCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_redirect_requests_total",
			Help: "Total number of redirect requests",
		},
	)

	// Monitore o tempo de resposta das requisições HTTP por rota:
	UrlRequests = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP por rota",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	// Monitore o total de requisições com erro (4xx, 5xx):
	HttpErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors by path and status code",
		},
		[]string{"path", "status"},
	)

	ResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Tamanho das respostas HTTP por rota",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10), // 100, 200, 400, ..., ~50k
		},
		[]string{"path"},
	)

	RedisCacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total Redis cache hits",
		},
	)

	RedisCacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total Redis cache misses",
		},
	)

	MongoOpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mongodb_operation_duration_seconds",
			Help:    "Duration of MongoDB operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	RedisOpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Duration of Redis operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)

	InvalidTokens = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_invalid_tokens_total",
			Help: "Total number of invalid or expired tokens",
		},
	)
)

func InitCustomMetrics() {
	prometheus.MustRegister(ShortenCounter)
	prometheus.MustRegister(RedirectCounter)
	prometheus.MustRegister(UrlRequests)
	prometheus.MustRegister(HttpErrors)
	prometheus.MustRegister(ResponseSize)
	prometheus.MustRegister(RedisCacheHits)
	prometheus.MustRegister(RedisCacheMisses)
	prometheus.MustRegister(MongoOpDuration)
	prometheus.MustRegister(RedisOpDuration)
	prometheus.MustRegister(InvalidTokens)
}
