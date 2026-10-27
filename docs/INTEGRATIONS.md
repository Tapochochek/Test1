# Интеграции (Д2)

Токены и вебхуки создаёт федерация на странице «Процессы». Права токена: `competitions:read`, `rating:read`, `results:write`.

```bash
# календарь и рейтинг
curl -H "Authorization: Bearer fsp_..." http://localhost:8080/api/ext/competitions
curl -H "Authorization: Bearer fsp_..." "http://localhost:8080/api/ext/rating?discipline=..."
# импорт результатов из внешней платформы (протокол формируется автоматически, только одобренные заявки)
curl -X POST -H "Authorization: Bearer fsp_..." -H "Content-Type: application/json" \
  -d '{"items":[{"fsp_id":"FSP-001","place":1,"score":100}]}' http://localhost:8080/api/ext/competitions/1/results
```

Вебхуки: `competition.published`, `application.approved`, `protocol.published`. Тело `{"event","data","sent_at"}`, заголовки `X-FSP-Event` и `X-FSP-Signature` = hex(HMAC-SHA256(secret, тело)). Попытки доставки видны в журнале на странице «Процессы». Проверка быстро: `https://webhook.site` или локальный `nc -l`.
