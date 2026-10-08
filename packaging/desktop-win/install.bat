@echo off
setlocal EnableExtensions
set "DEST=%LOCALAPPDATA%\LifeOS\desktop"
mkdir "%DEST%" 2>nul
copy /Y "%~dp0lifeos-desktop.exe" "%DEST%\lifeos-desktop.exe" >nul
copy /Y "%~dp0LifeOS.bat" "%DEST%\LifeOS.bat" >nul
copy /Y "%~dp0README.txt" "%DEST%\README.txt" >nul
powershell -NoProfile -Command "$s = (New-Object -ComObject WScript.Shell).CreateShortcut([Environment]::GetFolderPath('Desktop') + '\LifeOS.lnk'); $s.TargetPath = $env:LOCALAPPDATA + '\LifeOS\desktop\LifeOS.bat'; $s.WorkingDirectory = $env:LOCALAPPDATA + '\LifeOS\desktop'; $s.WindowStyle = 7; $s.Description = 'LifeOS'; $s.Save()"
echo Installed to %DEST%
echo Shortcut: Desktop\LifeOS.lnk
echo Sign in with your Telegram nick and the code the bot sends.
exit /b 0
