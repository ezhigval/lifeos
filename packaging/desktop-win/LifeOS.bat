@echo off
setlocal EnableExtensions
cd /d "%~dp0"

if not defined LIFEOS_API_ORIGIN set "LIFEOS_API_ORIGIN=https://local-ai-assist.ru"
if not defined LIFEOS_DESKTOP_DATA set "LIFEOS_DESKTOP_DATA=%LOCALAPPDATA%\LifeOS"
if not defined LIFEOS_DESKTOP_ADDR set "LIFEOS_DESKTOP_ADDR=127.0.0.1:47321"

if not exist "%LIFEOS_DESKTOP_DATA%" mkdir "%LIFEOS_DESKTOP_DATA%"

for /f "tokens=2 delims=:" %%P in ("%LIFEOS_DESKTOP_ADDR%") do set "PORT=%%P"
if not defined PORT set "PORT=47321"

call :listening
if not errorlevel 1 goto open

start "LifeOS" /MIN "%~dp0lifeos-desktop.exe"
set /a TRIES=0
:wait
set /a TRIES+=1
call :listening
if not errorlevel 1 goto open
if %TRIES% GEQ 40 goto fail
timeout /t 1 /nobreak >nul
goto wait

:open
call :browser
if defined BROWSER (
  start "" "%BROWSER%" --app=http://127.0.0.1:%PORT%/ --user-data-dir="%LIFEOS_DESKTOP_DATA%\webview"
) else (
  start "" "http://127.0.0.1:%PORT%/"
)
exit /b 0

:fail
echo LifeOS desktop did not open %LIFEOS_DESKTOP_ADDR%.
echo Log: %LIFEOS_DESKTOP_DATA%\desktop.log
exit /b 1

:listening
powershell -NoProfile -Command "try { $c = New-Object Net.Sockets.TcpClient; $c.Connect('127.0.0.1', %PORT%); $c.Close(); exit 0 } catch { exit 1 }"
exit /b %errorlevel%

:browser
set "BROWSER="
set "PF86=%ProgramFiles(x86)%"
if exist "%PF86%\Microsoft\Edge\Application\msedge.exe" set "BROWSER=%PF86%\Microsoft\Edge\Application\msedge.exe"
if not defined BROWSER if exist "%ProgramFiles%\Microsoft\Edge\Application\msedge.exe" set "BROWSER=%ProgramFiles%\Microsoft\Edge\Application\msedge.exe"
if not defined BROWSER if exist "%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe" set "BROWSER=%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe"
exit /b 0
