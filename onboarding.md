1. перейдите в корневую папку
2. добавьте свой токен в .env (APP_TELEGRAM_TOKEN=..)
3. добавьте в .env<br>
APP_TELEGRAM_COMMANDS_PATH=commands.json<br>
APP_SCRAPPER_BASE_URL=http://localhost:9080<br>
APP_BOT_BASE_URL=http://localhost:9090<br>
APP_POSTGRES_DB=db_example<br>
APP_DATABASE_USER=user_example<br>
APP_DATABASE_PASSWORD=pass_example<br>
APP_DATABASE_URL=postgres://user_example:pass_example@localhost:5432/db_example?sslmode=disable<br>
APP_DATABASE_ACCESS_TYPE=SQL<br>
4. поднимите postgres `docker compose up`
5. `set -a`<br>
`source ./.env`<br>
`set +a`
6. go run ./cmd/scrapper
7. go run ./cmd/bot
8.  e2e тест `RUN_TESTCONTAINERS=1 go test ./internal/e2e -count=1 -v` - надо запустить docker daemon
9.  интеграционные тесты бд `RUN_TESTCONTAINERS=1 go test ./internal/infrastructure/postgres/sql ./internal/infrastructure/postgres/orm -count=1 -v` - нужен docker daemon
10.  unit тесты `go test $(go list ./... | grep -v '/proto/gen$' | grep -v '/cmd/'| grep -v '/internal/e2e') -count=1 -coverprofile=coverage.out`
11.   cover `go tool cover -func=coverage.out`
12.   линтер `golangci-lint run -c .golangci.yml`
