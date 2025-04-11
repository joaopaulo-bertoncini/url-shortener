package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UrlShortenedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortened_total",
			Help: "Total de URLs encurtadas",
		})

	UrlRedirectedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_redirected_total",
			Help: "Total de redirecionamentos feitos",
		})

	UrlRequests = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP por rota",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
)

func Init() {
	prometheus.MustRegister(UrlShortenedTotal)
	prometheus.MustRegister(UrlRedirectedTotal)
	prometheus.MustRegister(UrlRequests)
}
