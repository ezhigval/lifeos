# ВМ: секреты, автодеплой, туннель Telegram

Один способ поднять LifeOS на виртуальной машине. Проверено на Ubuntu и Yandex Cloud. Quick-tunnel здесь не используется: с такой ВМ не открывается `api.trycloudflare.com`, а адрес всё равно сбрасывается при рестарте.

```
Telegram  --webhook-->  Cloudflare (твой домен)
                            |
                       named tunnel, TCP 7844, с ВМ наружу
                            |
                       127.0.0.1:8080  LifeOS

LifeOS --ответ бота--> 127.0.0.1:8081  локальный прокси
                            |
                       Cloudflare Worker
                            |
                       api.telegram.org
```

Туннель принимает Mini App, `/health` и webhook. Ответы бота идут через Worker, потому что до Telegram с ВМ напрямую часто нет маршрута. Имя хоста не меняется, когда перезапускается `cloudflared` или контейнер.

A-запись на IP машины не нужна. Входящие 80/8080 снаружи могут быть закрыты, и сертификат на самой ВМ это не лечит.

## Деньги

| Что | Зачем |
|-----|--------|
| Домен | стабильное имя. `.ru` на первый год часто 169–199 ₽, продление обычно 700–850 ₽. `.com` около 1500–1800 ₽/год |
| Cloudflare DNS, Tunnel, Worker | NS, HTTPS, туннель, исходящий прокси. На плане Free — 0 |
| Cloud DNS и Certificate Manager в Yandex | не нужны, пока NS и TLS на Cloudflare |

## Где что лежит

```
/opt/lifeos/
├── repo/            клон git, без секретов
├── .env             секреты приложения, chmod 600
├── secrets/         chmod 700
│   ├── tunnel.env   TUNNEL_TOKEN
│   └── tg-proxy.env URL воркера и секрет прокси
└── backups/
```

| Значение | Куда | В git |
|----------|------|-------|
| Домен и hostname | публично | можно |
| URL воркера | `/opt/lifeos/.env` | только плейсхолдер |
| Токен туннеля | `/opt/lifeos/secrets/tunnel.env` | никогда |
| `PROXY_SECRET` | воркер и `secrets/tg-proxy.env` | никогда |
| Токен бота, JWT, пароль БД | `/opt/lifeos/.env` | никогда |

Перед push: `bash scripts/check-no-secrets.sh`.

## 1. Автодеплой

На ВМ, от root. Для приватного репозитория сначала deploy-ключ только на чтение:

```bash
ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N "" -C "lifeos-deploy@vm"
cat /root/.ssh/lifeos_deploy.pub
```

Ключ добавляется в GitHub → Settings → Deploy keys. Галочка write не нужна.

```bash
sudo LIFEOS_REPO_URL=git@github.com:ezhigval/lifeos.git bash deployments/autodeploy.sh
nano /opt/lifeos/.env
sudo systemctl start lifeos-deploy.service
curl -fsS http://127.0.0.1:8080/health
```

В `.env` обязательны `TELEGRAM_BOT_TOKEN` и `LIFEOS_JWT_SECRET` (от 32 байт). Таймер `lifeos-deploy.timer` примерно раз в 5 минут делает `git fetch`. Образ пересобирается только если SHA в `origin/main` отличается от `/opt/lifeos/.last_deploy_sha`. Миграции выполняются в контейнере до рестарта приложения.

Повторный запуск того же скрипта безопасен. Снести установку и поставить заново, сохранив дамп БД: `sudo bash deployments/vm-reset.sh`. Полный проход с переносом старого тома: `sudo bash deployments/vm-bootstrap.sh`.

Ручной цикл:

```bash
sudo systemctl start lifeos-deploy.service
journalctl -u lifeos-deploy -n 80 --no-pager
```

## 2. Домен и NS

1. Купи домен.
2. Добавь сайт в Cloudflare, план Free.
3. Пропиши у регистратора два NS, которые покажет Cloudflare.
4. Дождись статуса Active.

## 3. Named tunnel

Создаётся с компьютера, где открыт API Cloudflare. С ВМ этот API на 443 обычно закрыт.

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → Networks → Tunnels → Create.
2. Имя, например `lifeos`. Коннектор: cloudflared.
3. Скопируй токен со шага Install connector. Это длинная строка, чаще всего с `eyJ`.
4. Public Hostname: тип **HTTP**, URL **`http://127.0.0.1:8080`**.
5. Hostname без пути и без слэша на конце, например `https://lifeos.example.com`.

На ВМ токен не должен попасть в историю шелла и в git:

```bash
sudo install -d -m 700 /opt/lifeos/secrets
read -rs TUNNEL_TOKEN
printf 'TUNNEL_TOKEN=%s\n' "$TUNNEL_TOKEN" | sudo tee /opt/lifeos/secrets/tunnel.env >/dev/null
unset TUNNEL_TOKEN
sudo chmod 600 /opt/lifeos/secrets/tunnel.env
sudo bash /opt/lifeos/repo/deployments/cloudflared/install-named-tunnel.sh
```

Скрипт ставит `cloudflared`, если его нет, и запускает `lifeos-tunnel.service`. Протокол HTTP/2, только IPv4, TCP 7844. Токен читается из файла окружения и не передаётся аргументом процесса.

`install-named-tunnel.sh` отказывается писать токен внутрь git-клона.

## 4. Исходящий Bot API

Код воркера: `deployments/vps/tg-proxy-worker.js`. Деплой с ноутбука:

```bash
cd deployments/vps
npx wrangler login
npx wrangler deploy
```

Запомни `https://<имя>.<аккаунт>.workers.dev`.

Общий секрет, чтобы воркером не пользовались чужие:

```bash
npx wrangler secret put PROXY_SECRET
```

На ВМ:

```bash
cd /opt/lifeos/repo
sudo LIFEOS_TG_PROXY_WORKER_URL='https://<worker>' \
  LIFEOS_TG_PROXY_SECRET='<тот же секрет>' \
  bash deployments/vps/tg-proxy.sh install
sudo bash deployments/vps/tg-proxy.sh test
```

`test` печатает `ok`, id и username. Токен бота не печатает. Секрет остаётся в `/opt/lifeos/secrets/tg-proxy.env`. В `.env` приложения попадает только `LIFEOS_HTTP_PROXY`. Для Docker это `http://host.docker.internal:8081`: прокси слушает localhost и адрес моста `lifeos_default`, не публичный интерфейс.

Клиент Telegram при непустом `LIFEOS_HTTP_PROXY` ходит на `http://api.telegram.org`. Иначе библиотека открыла бы CONNECT, и прокси не увидел бы запрос.

Если `test` пишет, что воркер недоступен, с ВМ закрыт 443 до `*.workers.dev`. Туннель и Mini App при этом живы. Для ответов бота нужен релей с тем же контрактом `GET/POST /fetch?path=` на хосте, который ВМ открывает по 443. Его URL подставляется вместо URL воркера.

## 5. Публичный origin

После того как туннель в статусе Healthy:

```bash
sudo LIFEOS_PUBLIC_ORIGIN='https://lifeos.example.com' \
  bash /opt/lifeos/repo/deployments/apply-public-origin.sh
```

Скрипт выставляет webhook и `LIFEOS_MINIAPP_URL`, при пустом секрете webhook генерирует новый, пересоздаёт контейнер и регистрирует webhook через локальный прокси. Секрет в вывод не попадает. Файл `.env` внутри git-клона скрипт не редактирует.

## Проверка

```bash
curl -fsS http://127.0.0.1:8080/health
systemctl is-active lifeos-tunnel lifeos-tg-proxy lifeos-deploy.timer
sudo bash /opt/lifeos/repo/deployments/vps/tg-proxy.sh test
curl -fsS https://lifeos.example.com/health
```

В Telegram отправь `/start` и открой Mini App новой кнопкой меню. Старая ссылка в истории чата может вести на умерший адрес.

С самой ВМ публичный `/health` иногда не открывается из-за hairpin через Cloudflare. Это нормально, если с другой машины тот же URL отвечает.
