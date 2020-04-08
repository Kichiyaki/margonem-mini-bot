@ECHO OFF
set /p machineID="Enter MachineID: "
rm -rf build\windows
go build -ldflags "-X main.Mode=production -X main.MachineID=%machineID%" -o build\windows\bot.exe
copy %cd%\default_config.json %cd%\build\windows\config.json
copy %cd%\readme.md %cd%\build\windows\readme.md