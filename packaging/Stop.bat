@echo off
cd /d "%~dp0"
if exist run\lifeos.pid (
  for /f %%p in (run\lifeos.pid) do taskkill /PID %%p /F >nul 2>&1
  del /f /q run\lifeos.pid >nul 2>&1
)
taskkill /FI "IMAGENAME eq lifeos.exe" /F >nul 2>&1
echo Stopped.
pause
