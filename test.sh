curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer secure-dev-token-123" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.mercadolivre.com.br"}'

export PORT=8080
export REDIS_ADDR="localhost:6379"
export MONGO_URI="mongodb://localhost:27017"
export URL_PREFIX="http://localhost:8080/"
export AUTH_TOKEN="testtoken123"


http://localhost:9090
http://localhost:3000