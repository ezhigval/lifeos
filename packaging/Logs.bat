@echo off
cd /d "%~dp0"
if not exist logs mkdir logs
if not exist logs\lifeos.log type nul > logs\lifeos.log
powershell -NoProfile -Command "Get-Content -Wait -Tail 80 logs\lifeos.log"
