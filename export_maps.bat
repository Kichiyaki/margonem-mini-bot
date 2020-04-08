@echo off
IF NOT EXIST "config.json" COPY "default_config.json" "config.json"
cd ./cmd/maps_exporter
go run main.go