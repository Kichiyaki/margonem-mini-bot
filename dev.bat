@echo off
IF NOT EXIST "config.json" COPY "default_config.json" "config.json"
set MODE=development
go run ./cmd/bot/main.go