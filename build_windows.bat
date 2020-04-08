@ECHO OFF
set /p machineID="Enter MachineID: "
rm -rf build\windows
cd .\cmd\bot
windres -O coff -o bot.syso bot.rc
go build -ldflags "-X main.Mode=production -X main.MachineID=%machineID%" -o .\..\..\build\windows\bot.exe
cd .\..\..
cd .\cmd\maps_exporter
windres -O coff -o maps_exporter.syso maps_exporter.rc
go build -ldflags "-X main.Mode=production -X main.MachineID=%machineID%" -o .\..\..\build\windows\maps_exporter.exe
cd .\..\..
copy %cd%\default_config.json %cd%\build\windows\config.json
copy %cd%\readme.md %cd%\build\windows\readme.md