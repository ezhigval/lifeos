LifeOS window (Windows)
======================

This zip is the desktop window: the same screens as the Mac app, with the
assistant docked on the right. It is not the older Start.bat server bundle
from `make package-win`.

What is inside
--------------
  lifeos-desktop.exe   local window server (Go, windows/amd64, no console)
  LifeOS.bat           starts that server and opens Edge or Chrome in app mode
  install.bat          copies the folder to %LOCALAPPDATA%\LifeOS\desktop
                       and adds a Desktop shortcut
  README.txt           this file

There is no signed install.exe. This Linux-built zip is unsigned.

First run
---------
  1. Unzip.
  2. Double-click install.bat, or run LifeOS.bat from this folder.
  3. Sign in with your Telegram nick. The bot sends a code in the chat.
     Press /start in the bot once if it does not know you yet.
  4. The right-hand column talks to POST /api/v1/assistant/chat on the same
     account as Telegram and the Mini App.

The window talks to https://local-ai-assist.ru unless you set
LIFEOS_API_ORIGIN before launching. The listen address defaults to
127.0.0.1:47321 (LIFEOS_DESKTOP_ADDR). Logs:
%LOCALAPPDATA%\LifeOS\desktop.log

Edge or Chrome is required for the app window. If neither is installed,
LifeOS.bat opens the default browser at the same local URL.

Build
-----
  bash scripts/build-win-desktop.sh
  LIFEOS_DESKTOP_VERSION=desktop-v0.4.0 bash scripts/build-win-desktop.sh

The GitHub workflow .github/workflows/desktop-release.yml publishes the Mac
.app only. It has no Windows runner, so this zip is not attached there.
