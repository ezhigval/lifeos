@echo off
cd /d "%~dp0"
if not exist settings.env copy /Y settings.env.example settings.env >nul
notepad settings.env
