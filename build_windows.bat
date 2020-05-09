@ECHO OFF
rm -rf build\windows
cd .\cmd\bot
windres -O coff -o bot.syso bot.rc
go build -o .\..\..\build\windows\bot.exe
cd .\..\..
cd .\cmd\maps_exporter
windres -O coff -o maps_exporter.syso maps_exporter.rc
go build -o .\..\..\build\windows\maps_exporter.exe
cd .\..\..
copy %cd%\default_config.json %cd%\build\windows\config.json
copy %cd%\readme.md %cd%\build\windows\readme.md