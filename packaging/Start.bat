@echo off
REM Start LifeOS on Windows — binary loads settings.env from this folder.
setlocal
cd /d "%~dp0"

if not exist "settings.env" (
  copy /Y settings.env.example settings.env >nul
  echo Created settings.env — fill TELEGRAM_BOT_TOKEN, JWT, DATABASE_URL
  start notepad settings.env
  echo Save the file and run Start.bat again.
  pause
  exit /b 0
)

if not exist "logs" mkdir logs
if not exist "run" mkdir run

if not exist "bin\lifeos.exe" (
  echo Missing bin\lifeos.exe
  pause
  exit /b 1
)

set "LIFEOS_STATIC_DIR=%cd%\web\miniapp\dist"
set "LIFEOS_MIGRATIONS_DIR=%cd%\migrations"

echo Migrating...
bin\lifeos.exe migrate up
if errorlevel 1 (
  echo migrate failed — check DATABASE_URL in settings.env
  pause
  exit /b 1
)

echo Starting… log: logs\lifeos.log
start "LifeOS logs" cmd /k "cd /d ""%~dp0"" && Logs.bat"
bin\lifeos.exe serve >> logs\lifeos.log 2>&1
echo Server stopped.
pause
