# Telegram

Бот — основной интерфейс. Mini App — второй клиент на том же API. По умолчанию бот забирает апдейты сам (`LIFEOS_TELEGRAM_MODE=polling`) и для этого должен достучаться до `api.telegram.org`.

## Бот

1. В [@BotFather](https://t.me/BotFather) создай бота и возьми токен.
2. Положи его в `TELEGRAM_BOT_TOKEN`. Старое имя `LIFEOS_TELEGRAM_BOT_TOKEN` тоже читается.
3. Запусти приложение ([getting-started.md](getting-started.md)) и напиши боту `/start`.

Без публичного HTTPS Mini App из Telegram не откроется. Сам бот в режиме polling работает.

## Mini App на ноутбуке

Короткий туннель Cloudflare подходит для разработки на машине, где открыт `api.trycloudflare.com`. Адрес меняется при перезапуске.

```bash
make miniapp-build
make docker-up
make tunnel
make docker-up
```

`make tunnel` пишет `LIFEOS_MINIAPP_URL` в `.env`. Кнопка меню в боте берёт этот адрес. Он должен заканчиваться на `/app/`.

`make stack-up` делает сборку фронта, подъём стека и туннель одной целью.

Ошибка Cloudflare 1033 значит, что имя туннеля есть, а `http://127.0.0.1:8080/health` не отвечает.

На части сетей (в том числе на типичной ВМ в Yandex Cloud) quick-tunnel не создаётся: закрыт `api.trycloudflare.com:443`. Постоянный адрес на домашнем Windows, тем же доменом: [deploy/windows.md](deploy/windows.md).

## Webhook

Webhook включай, когда есть стабильный HTTPS.

```env
LIFEOS_TELEGRAM_MODE=webhook
LIFEOS_TELEGRAM_WEBHOOK_URL=https://<hostname>/webhook/telegram
LIFEOS_TELEGRAM_WEBHOOK_SECRET=<openssl rand -hex 32>
LIFEOS_MINIAPP_URL=https://<hostname>/app/
```

При старте процесс сам вызывает `setWebhook` и ставит кнопку меню. Если прямой выход к Telegram закрыт, исходящие вызовы идут через локальный прокси. Это описано в разделе туннеля на ВМ.

Проверка входа Mini App с живым initData: `make verify-webapp-auth`.
