1. перейдите в корневую папку проекта

2. создайте `.env` одним блоком:

```bash
cat > .env <<'ENV'
APP_TELEGRAM_TOKEN=your_telegram_token_here
APP_TELEGRAM_COMMANDS_PATH=commands.json
APP_SCRAPPER_BASE_URL=http://localhost:8080
APP_BOT_BASE_URL=http://localhost:8090

APP_POSTGRES_DB=link_tracker
APP_DATABASE_USER=example_user
APP_DATABASE_PASSWORD=example_password
APP_DATABASE_URL=postgres://example_user:example_password@localhost:5432/link_tracker?sslmode=disable

APP_DATABASE_ACCESS_TYPE=SQL
ENV
```

3. поднимите postgres:

```bash
docker compose up -d postgres
```

4. загрузите переменные окружения:

```bash
set -a
source .env
set +a
```

5. запустите сервисы в двух отдельных терминалах:

```bash
go run ./cmd/bot
```

```bash
go run ./cmd/scrapper
```

`example_user`, `example_password` и `APP_DATABASE_URL` в примере выше - это просто для примера

