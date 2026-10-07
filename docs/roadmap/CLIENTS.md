# Клиенты

Карта того, что уже совпадает с видением владельца, и что ещё резать. Стадийный роман `docs/roadmap/ROADMAP.md` с дерева снят в `cd7ebf9` (альфа 1.0); этот файл его не восстанавливает. Живые документы: [оглавление](../README.md), [агент](../ai/AGENT.md), [Telegram](../telegram.md).

## Совпадает

| Видение | Где |
|---------|-----|
| Telegram — диалоговый агент с tools (задачи, финансы, привычки, календарь и остальные) | [docs/ai/AGENT.md](../ai/AGENT.md), `internal/ai`, `cmd/lifeos/cmd/agent_wire.go` |
| Mini App показывает продукт | маршруты задач, сфер, привычек, календаря, заметок, здоровья, долгов, напоминаний, аналитики, настроек; финансы на главной |
| Один аккаунт. Вне Telegram вход — код в чат бота | `migrations/00037_telegram_login_codes.sql`, `internal/identity/app/login_code.go`, `POST /api/v1/auth/telegram-login/*`, `web/miniapp/src/components/WebLogin.tsx` |
| Mac: окно с теми же экранами | `desktop/macos` (Swift `.app`), `cmd/lifeos-desktop`, `web/miniapp/desktop/DesktopFrame.tsx`. Релиз: тег `desktop-v*` и [`.github/workflows/desktop-release.yml`](../../.github/workflows/desktop-release.yml) на `macos-14`, команда `bash scripts/build-mac-app.sh` → `dist/LifeOS-mac.zip` |
| HTTP того же агента | `POST /api/v1/assistant/chat` вызывает `dialogue.Service`, не второго агента |

`make package` / `make package-win` — старый архив сервера (`Start.command` / `Start.bat`, логи, settings). Это не окно десктопа.

## Этот срез

- Колонка ассистента справа в десктоп-окне. Тот же `POST /api/v1/assistant/chat` и тот же JWT после кода из Telegram.
- Windows: `bash scripts/build-win-desktop.sh` → `dist/LifeOS-win.zip` (`lifeos-desktop.exe` + `LifeOS.bat` + `install.bat`). Окно — Edge/Chrome `--app` на локальный сервер UI. Подписи нет. В workflow нет Windows-раннера, zip туда не вешается. `.app` на этой Linux-машине Swift не собирает: `build-mac-app.sh` здесь останавливается после Go-бинарей.

## Дальше, по порядку

1. UI-чат внутри Mini App. В [AGENT.md](../ai/AGENT.md) он всё ещё отложен; десктоп его не заменяет.
2. Экраны, которых нет в Mini App, хотя tools агента есть: карьера (UI снимали), личная память.
3. Подписанный Windows-инсталлятор и публикация zip в том же релизе, когда появится Windows-раннер. До этого команда сборки — `scripts/build-win-desktop.sh`.
4. Workspaces (Stage 4) не начинать в этом срезе.
