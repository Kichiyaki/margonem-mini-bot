@echo off
IF NOT EXIST "config.json" COPY "default_config.json" "config.json"
cd ./cmd/bot
go run main.go