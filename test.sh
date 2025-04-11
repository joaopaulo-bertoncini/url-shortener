curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer secure-dev-token-123" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.mercadolivre.com.br"}'