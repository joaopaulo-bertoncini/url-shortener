# 🔗 URL Shortener API

Este projeto é uma API de encurtador de URLs escalável e instrumentada com **OpenTelemetry**, **Prometheus**, **Grafana**, **Jaeger**, **Elasticsearch** e **Kibana**. Ele utiliza as tecnologias: Go, Redis, MongoDB, Docker e muito mais.

---

## 🚀 Tecnologias

- [x] **Go**: linguagem principal do projeto
- [x] **Gin**: web framework
- [x] **MongoDB**: banco para persistência das URLs
- [x] **Redis**: cache de URLs curtas
- [x] **Docker** + **docker-compose**: para orquestração
- [x] **Prometheus** + **Grafana**: métricas
- [x] **OpenTelemetry** + **Jaeger**: tracing
- [x] **Elasticsearch** + **Kibana** + **Filebeat**: logs centralizados

---

## 🛠️ Executando o projeto

### 1. Clonar o repositório

```bash
git clone https://github.com/seu-usuario/url-shortener.git
cd url-shortener
docker-compose up --build

