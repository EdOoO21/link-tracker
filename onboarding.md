1. перейдите в корневую папку
2. добавьте свой токен в .env (APP_TELEGRAM_TOKEN=..)
3. добавьте в .env
APP_TELEGRAM_COMMANDS_PATH=commands.json
APP_SCRAPPER_BASE_URL=http://localhost:9080
APP_BOT_BASE_URL=http://localhost:9090
APP_POSTGRES_DB=db_example
APP_DATABASE_USER=user_example
APP_DATABASE_PASSWORD=pass_example
APP_DATABASE_URL=postgres://user_example:pass_example@localhost:5432/db_example?sslmode=disable
4. set -a
5. source ./.env
6. set +a
7. go run ./cmd/scrapper
8. go run ./cmd/bot
9.  интеграционный тест `RUN_TESTCONTAINERS=1 go test ./internal/integration -count=1 -v` - надо запустить docker daemon
10. тесты без gen `go test $(go list ./... | grep -v '/proto/gen$' | grep -v '/cmd/') -count=1 -coverprofile=coverage.out`
11. cover `go tool cover -func=coverage.out`
12. линтер `golangci-lint run -c .golangci.yml`