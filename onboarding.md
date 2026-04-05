1. перейдите в корневую папку
2. добавьте в .env
APP_TELEGRAM_TOKEN=your_token
APP_TELEGRAM_COMMANDS_PATH=commands.json
APP_SCRAPPER_BASE_URL=http://localhost:9080
APP_BOT_BASE_URL=http://localhost:9090
APP_POSTGRES_DB=db_example
APP_DATABASE_USER=user_example
APP_DATABASE_PASSWORD=pass_example
APP_DATABASE_URL=postgres://user_example:pass_example@localhost:5432/db_example?sslmode=disable
APP_DATABASE_ACCESS_TYPE=SQL
3. поднимите postgres `docker compose up`
4. ```set -a source ./.env set +a```
5. go run ./cmd/scrapper
6. go run ./cmd/bot
7.  интеграционный тест `RUN_TESTCONTAINERS=1 go test ./internal/e2e -count=1 -v` - надо запустить docker daemon
8.  тесты без gen `go test $(go list ./... | grep -v '/proto/gen$' | grep -v '/cmd/'| grep -v '/internal/e2e') -count=1 -coverprofile=coverage.out`
9.   cover `go tool cover -func=coverage.out`
10.   линтер `golangci-lint run -c .golangci.yml`
