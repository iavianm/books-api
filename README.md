# books-api

**[English](#english) · [Русский](#русский)**

---

## English

A CRUD REST API for a book catalogue, written in Go on top of the standard library.

### Stack

| Component | Choice |
| --- | --- |
| Routing | `net/http.ServeMux` (method patterns and path values, Go 1.22+) |
| Database | PostgreSQL 18 + `database/sql` with the `pgx/v5` driver |
| Migrations | `golang-migrate`, SQL files embedded into the binary via `embed.FS` |
| Logging | `log/slog`, JSON output |
| Caching | in-memory cache with TTL and single-flight loading |
| Config | environment variables |

No web framework and no ORM: routing, middleware, JSON and SQL are written directly against the standard library.

### Quick start

Requires Docker and Docker Compose.

```sh
git clone https://github.com/iavianm/books-api.git
cd books-api
cp .env.example .env
make up
```

The API is now available at `http://localhost:8080`. Migrations run automatically on startup.

```sh
curl localhost:8080/health
make logs    # follow application logs
make down    # stop everything
```

#### Running locally without Docker

Start only the database, then run the application on the host:

```sh
cp .env.example .env
make db-up
make run
```

`make run` loads `.env` through the Makefile, so `DB_HOST=localhost` from that file is used. Inside Compose the same variables are overridden with `DB_HOST=postgres`.

### Configuration

All settings come from environment variables; `.env.example` is a ready-to-use template.

| Variable | Example | Description |
| --- | --- | --- |
| `HTTP_PORT` | `8080` | port the HTTP server listens on |
| `DB_HOST` | `localhost` | database host (`postgres` inside Compose) |
| `DB_PORT` | `5432` | database port |
| `DB_USER` | `postgres_user` | database user |
| `DB_PASSWORD` | `postgres_password` | database password |
| `DB_NAME` | `books` | database name |
| `DB_SSLMODE` | `disable` | SSL mode for the connection |
| `CACHE_TTL` | `30s` | cache entry lifetime, any `time.ParseDuration` value |

Every variable is mandatory: the application fails fast at startup if one is missing, instead of running with silent defaults.

### API

| Method | Path | Description | Success |
| --- | --- | --- | --- |
| `GET` | `/health` | health check | `200` |
| `GET` | `/books` | list all books | `200` |
| `GET` | `/books/{id}` | get a book by id | `200` |
| `POST` | `/books` | create a book | `201` |
| `PUT` | `/books/{id}` | update a book by id | `200` |
| `DELETE` | `/books/{id}` | delete a book by id | `204` |

#### Book

```json
{
  "id": 1,
  "title": "War and Peace",
  "author": "Leo Tolstoy",
  "year": 1869,
  "genre": "novel",
  "created_at": "2026-08-16T01:00:00Z",
  "updated_at": "2026-08-16T01:00:00Z"
}
```

`id`, `created_at` and `updated_at` are managed by the server and ignored in request bodies.

#### Examples

```sh
# create
curl -X POST localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"title":"War and Peace","author":"Leo Tolstoy","genre":"novel","year":1869}'

# list
curl localhost:8080/books

# get by id
curl localhost:8080/books/1

# update
curl -X PUT localhost:8080/books/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"War and Peace","author":"Leo Tolstoy","genre":"epic","year":1869}'

# delete
curl -X DELETE localhost:8080/books/1
```

#### Errors

Errors share one JSON shape:

```json
{ "message": "validation failed: title is required" }
```

| Status | When |
| --- | --- |
| `400` | malformed JSON, non-positive or non-numeric `id`, failed validation |
| `404` | book does not exist, or unknown route |
| `405` | method not allowed for this path |
| `500` | internal error — details go to the log, never to the client |

Validation rules: `title`, `author` and `genre` must be non-blank; `year` must be between 1450 and next year. Surrounding whitespace is trimmed before saving.

### Project layout

```
cmd/api/            entry point: wiring and graceful shutdown
internal/
  config/           environment variables into a Config struct
  model/            Book and request structs
  repository/       SQL queries, nothing else
  cache/            caching decorator around the repository
  service/          validation and business rules
  handler/          HTTP: routing, JSON, status codes, middleware
  database/         connection pool and migration runner
migrations/         SQL migrations, embedded into the binary
```

Dependencies point one way: `handler → service → cache → repository → PostgreSQL`. Each layer declares the interface it needs from the one below, so the caching layer was inserted without touching the service or the repository.

### Caching

`GET /books` and `GET /books/{id}` are served from an in-memory cache with a `CACHE_TTL` lifetime. Writes invalidate the affected keys: `POST` drops the list, `PUT` and `DELETE` drop both the list and that book.

Concurrent misses for the same key are collapsed into a single database query (single-flight), so an expiring entry cannot turn a burst of requests into a burst of queries. Failed loads are never cached.

### Development

```sh
make build             # build the binary into ./.bin/api
make run               # build and run against a local database
make test              # unit tests, no database required
make test-integration  # repository tests against a running PostgreSQL
make cover             # total test coverage
make fmt               # gofmt + goimports
make lint              # golangci-lint
make db-psql           # psql shell into the database container
make migrate-create name=add_something   # new migration pair
make migrate-down                        # roll back the last migration
```

Integration tests are guarded by the `integration` build tag, so `make test` runs without a database. They need one started with `make db-up`.

### Graceful shutdown

On `SIGINT` or `SIGTERM` the server stops accepting new connections, gives in-flight requests up to 10 seconds to finish, then closes the cache and the database pool. `docker compose stop` triggers the same path.

---

## Русский

CRUD REST API для каталога книг на Go, написанный на стандартной библиотеке.

### Стек

| Компонент | Выбор |
| --- | --- |
| Роутинг | `net/http.ServeMux` (паттерны с методами и path-параметрами, Go 1.22+) |
| База данных | PostgreSQL 18 + `database/sql` с драйвером `pgx/v5` |
| Миграции | `golang-migrate`, SQL-файлы вшиты в бинарник через `embed.FS` |
| Логирование | `log/slog`, вывод в JSON |
| Кэширование | in-memory кэш с TTL и single-flight загрузкой |
| Конфигурация | переменные окружения |

Без веб-фреймворка и без ORM: роутинг, middleware, JSON и SQL написаны напрямую на стандартной библиотеке.

### Быстрый старт

Нужны Docker и Docker Compose.

```sh
git clone https://github.com/iavianm/books-api.git
cd books-api
cp .env.example .env
make up
```

API доступен на `http://localhost:8080`. Миграции применяются автоматически при старте.

```sh
curl localhost:8080/health
make logs    # смотреть логи приложения
make down    # остановить всё
```

#### Запуск без Docker

Поднять только базу, а приложение запустить на хосте:

```sh
cp .env.example .env
make db-up
make run
```

`make run` подгружает `.env` через Makefile, поэтому используется `DB_HOST=localhost` из этого файла. Внутри Compose те же переменные переопределяются на `DB_HOST=postgres`.

### Конфигурация

Все настройки берутся из переменных окружения, `.env.example` — готовый шаблон.

| Переменная | Пример | Описание |
| --- | --- | --- |
| `HTTP_PORT` | `8080` | порт HTTP-сервера |
| `DB_HOST` | `localhost` | хост базы (`postgres` внутри Compose) |
| `DB_PORT` | `5432` | порт базы |
| `DB_USER` | `postgres_user` | пользователь базы |
| `DB_PASSWORD` | `postgres_password` | пароль |
| `DB_NAME` | `books` | имя базы |
| `DB_SSLMODE` | `disable` | режим SSL для подключения |
| `CACHE_TTL` | `30s` | время жизни записи в кэше, любое значение для `time.ParseDuration` |

Все переменные обязательны: при отсутствии любой из них приложение падает на старте, а не работает с молчаливыми умолчаниями.

### API

| Метод | Путь | Описание | Успех |
| --- | --- | --- | --- |
| `GET` | `/health` | проверка живости | `200` |
| `GET` | `/books` | список всех книг | `200` |
| `GET` | `/books/{id}` | книга по id | `200` |
| `POST` | `/books` | создать книгу | `201` |
| `PUT` | `/books/{id}` | обновить книгу по id | `200` |
| `DELETE` | `/books/{id}` | удалить книгу по id | `204` |

#### Книга

```json
{
  "id": 1,
  "title": "Война и мир",
  "author": "Лев Толстой",
  "year": 1869,
  "genre": "роман",
  "created_at": "2026-08-16T01:00:00Z",
  "updated_at": "2026-08-16T01:00:00Z"
}
```

Поля `id`, `created_at` и `updated_at` ставит сервер — в теле запроса они игнорируются.

#### Примеры

```sh
# создать
curl -X POST localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"title":"Война и мир","author":"Лев Толстой","genre":"роман","year":1869}'

# список
curl localhost:8080/books

# получить по id
curl localhost:8080/books/1

# обновить
curl -X PUT localhost:8080/books/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Война и мир","author":"Лев Толстой","genre":"эпопея","year":1869}'

# удалить
curl -X DELETE localhost:8080/books/1
```

#### Ошибки

У всех ошибок единый формат:

```json
{ "message": "validation failed: title is required" }
```

| Код | Когда |
| --- | --- |
| `400` | битый JSON, неположительный или нечисловой `id`, непройденная валидация |
| `404` | книги не существует либо неизвестный маршрут |
| `405` | метод не разрешён для этого пути |
| `500` | внутренняя ошибка — подробности уходят в лог, но не клиенту |

Правила валидации: `title`, `author` и `genre` не должны быть пустыми; `year` — от 1450 до следующего года. Пробелы по краям обрезаются перед сохранением.

### Структура проекта

```
cmd/api/            точка входа: сборка зависимостей и graceful shutdown
internal/
  config/           переменные окружения в структуру Config
  model/            Book и структуры запросов
  repository/       SQL-запросы, и ничего больше
  cache/            кэширующий декоратор над репозиторием
  service/          валидация и бизнес-правила
  handler/          HTTP: роутинг, JSON, коды ответов, middleware
  database/         пул соединений и запуск миграций
migrations/         SQL-миграции, вшитые в бинарник
```

Зависимости направлены в одну сторону: `handler → service → cache → repository → PostgreSQL`. Каждый слой объявляет интерфейс, который ему нужен от нижележащего, — благодаря этому кэш удалось вставить в цепочку, не меняя ни сервис, ни репозиторий.

### Кэширование

`GET /books` и `GET /books/{id}` отдаются из in-memory кэша со временем жизни `CACHE_TTL`. Операции записи сбрасывают затронутые ключи: `POST` — список, `PUT` и `DELETE` — список и саму книгу.

Одновременные промахи по одному ключу схлопываются в один запрос к базе (single-flight), поэтому истёкшая запись не превращает наплыв запросов в наплыв SQL-запросов. Неуспешные загрузки в кэш не попадают.

### Разработка

```sh
make build             # собрать бинарник в ./.bin/api
make run               # собрать и запустить с локальной базой
make test              # unit-тесты, база не нужна
make test-integration  # тесты репозитория против запущенного PostgreSQL
make cover             # общее покрытие тестами
make fmt               # gofmt + goimports
make lint              # golangci-lint
make db-psql           # psql внутри контейнера с базой
make migrate-create name=add_something   # новая пара миграций
make migrate-down                        # откатить последнюю миграцию
```

Интеграционные тесты закрыты build-тегом `integration`, поэтому `make test` работает без базы. Для них нужна база, поднятая через `make db-up`.

### Graceful shutdown

По `SIGINT` или `SIGTERM` сервер перестаёт принимать новые соединения, даёт запросам в работе до 10 секунд на завершение, после чего закрывает кэш и пул соединений с базой. `docker compose stop` запускает тот же сценарий.
