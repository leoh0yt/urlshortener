# URL Shortener Service

Сервис для сокращения URL-адресов с поддержкой двух типов хранилищ (in-memory и PostgreSQL).
- При запуске принимает два флага port(порт сервера) и storage(тип хранилища)
- Поддерживает чтение конфигурации из .env файла
- Реализация логгирования и panic recovery с помощью middleware


## Оглавление

- [Требования](#требования)
- [Запуск](#запуск)
- [API](#api)

## Требования

- Docker
- Docker Compose



## Запуск
```bash
git clone git@github.com:leoh0yt/urlshortener.git
cd urlshortener
```
Необходимо добавление env файла:
```bash
POSTGRE_HOST=postgres_db
POSTGRE_PORT=5432
POSTGRE_DB_NAME=shortener
POSTGRE_USER=postgres
POSTGRE_PASSWORD=postgres
LOGS_LEVEL=info
```
docker compose build --no-cache && docker compose up --force-recreate

## API
### POST запрос

**Тело запроса:**

```http
POST /shorten
Content-Type: application/json

{
    "url": "https://example.com/very-long-url"
}
```
**Ответы:**

- **201** - успех
```json
{
    "short_url": "http://localhost:8080/0000000001",
    "original_url": "https://example.com/very-long-url",

}
```

- **400** ошибка: невалидный URL или пустой запрос

- **500** ошибка: внутренняя ошибка сервера

### GET запрос
**Запрос**
```http
GET /{short_code}
```
**Ответы**
- 301 успех
- 400 при пустом short_code
- 404 при отсутсвии ассоциированного original_url