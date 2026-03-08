1. перейдите в корневую папку
2. добавьте свой токен в .env (APP_TELEGRAM_TOKEN=..)
3. добавьте в .env APP_TELEGRAM_COMMANDS_PATH=commands.json
4. set -a
5. source .env
6. set +a
7. go run cmd/bot/main.go
