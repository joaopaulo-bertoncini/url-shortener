package handler

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ShortenCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_shorten_requests_total",
			Help: "Total number of shorten URL requests",
		},
	)

	RedirectCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_redirect_requests_total",
			Help: "Total number of redirect requests",
		},
	)
)

func InitCustomMetrics() {
	prometheus.MustRegister(ShortenCounter)
	prometheus.MustRegister(RedirectCounter)
}
