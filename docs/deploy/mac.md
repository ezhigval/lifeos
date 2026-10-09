# Mac дома: Docker и тот же домен

Прод без аренды машины: MacBook Air с Docker Desktop, домен `https://local-ai-assist.ru` и уже существующий named tunnel Cloudflare. Нового сервера нет. Пока Mac спит или крышка закрыта, сайт и бот недоступны.

Вход тот же, что в [EDGE.md](EDGE.md): Cloudflare → named tunnel → `http://127.0.0.1:8080`. Public Hostname в Zero Trust не меняй. Новый туннель не создавай. `make tunnel` не запускай: это короткий адрес для разработки, он подменит Mini App.

Туннель на ВМ Яндекса не трогай, пока на Mac не ответит `http://127.0.0.1:8080/health` и свой `cloudflared` не зарегистрирует соединение. Два коннектора с одним токеном делят запросы, поэтому публичный адрес проверяй уже после остановки туннеля на ВМ.

## 1. Docker Desktop

1. Скачай Docker Desktop для Apple Silicon: [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop/).
2. Открой `Docker.app` и дождись статуса Running.
3. В Терминале `docker info` должен показать Server Version.

## 2. Секреты с ВМ, не в чат

Токен бота, JWT и токен туннеля в чат, в git и в скриншот не попадают. Копируй файлами.

```bash
git clone https://github.com/ezhigval/lifeos.git
cd lifeos
mkdir -p "$HOME/.config/lifeos"
chmod 700 "$HOME/.config/lifeos"
# root — как в установке на ВМ. Вывод сразу в файл, в чат не вставляй.
ssh root@93.77.160.149 'cat /opt/lifeos/.env' > "$HOME/.config/lifeos/env"
ssh root@93.77.160.149 'cat /opt/lifeos/secrets/tunnel.env' > "$HOME/.config/lifeos/tunnel.env"
chmod 600 "$HOME/.config/lifeos/env" "$HOME/.config/lifeos/tunnel.env"
cp "$HOME/.config/lifeos/env" .env
chmod 600 .env
```

`.env` в корне уже в `.gitignore`. Compose читает его из `deployments/docker-compose.yml` (`env_file: ../.env`). `LIFEOS_MINIAPP_URL` и `LIFEOS_TELEGRAM_WEBHOOK_URL` оставь как на ВМ: `https://local-ai-assist.ru/app/` и `https://local-ai-assist.ru/webhook/telegram`. Режим — `webhook`.

Та же база, дамп тоже вне git:

```bash
ssh root@93.77.160.149 'docker exec "$(docker ps -qf name=lifeos-postgres)" pg_dump -U lifeos --clean --if-exists lifeos' > "$HOME/.config/lifeos/lifeos.sql"
```

## 3. Достаёт ли Mac до Telegram

Один раз:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' --max-time 15 https://api.telegram.org
```

Код `200`, `302` или `404` значит, что сеть до Telegram есть. В `.env` удали строки `LIFEOS_HTTP_PROXY` и `LIFEOS_TG_PROXY_WORKER_URL`. Прокси ВМ на Mac не нужен: контейнер сам ходит на `api.telegram.org`.

Таймаут или отказ соединения значит, что домашний провайдер режет Telegram. Тогда оставь `LIFEOS_HTTP_PROXY` как на ВМ (`http://host.docker.internal:8081`) и подними тот же исходящий прокси из [vm.md](vm.md), раздел «Исходящий Bot API». На Docker Desktop имя `host.docker.internal` уже резолвится.

## 4. Compose

Из корня репозитория:

```bash
make docker-up
curl -fsS http://127.0.0.1:8080/health
```

`make docker-up` — это `docker compose -f deployments/docker-compose.yml up -d --build`. Слушают только `127.0.0.1:8080` (приложение) и `127.0.0.1:5433` (Postgres). Образ собирается на Mac.

Если дамп есть и файл не пустой, останови приложение, залей дамп и подними его снова. Пустой файл значит, что имя контейнера другое: на ВМ `docker ps --format '{{.Names}}'` и подставь контейнер Postgres. После старта entrypoint снова прогоняет миграции.

```bash
docker compose -f deployments/docker-compose.yml stop app
docker compose -f deployments/docker-compose.yml exec -T postgres psql -U lifeos -d lifeos < "$HOME/.config/lifeos/lifeos.sql"
docker compose -f deployments/docker-compose.yml start app
curl -fsS http://127.0.0.1:8080/health
```

## 5. Тот же cloudflared

Токен — файл из шага 2. Скрипт его не печатает и отказывается читать токен из git-клона.

```bash
brew install cloudflared
LIFEOS_TUNNEL_ENV="$HOME/.config/lifeos/tunnel.env" bash deployments/cloudflared/run-tunnel.sh
```

Терминал не закрывай. `install-named-tunnel.sh` ставит systemd на Linux; на Mac он не нужен. В Zero Trust сервис hostname остаётся HTTP `http://127.0.0.1:8080`.

## 6. Доказательство, затем коннектор на ВМ

Пока оба `cloudflared` живы, публичный URL может отвечать с ВМ. Это не доказательство, что отвечает Mac.

1. На Mac: `curl -fsS http://127.0.0.1:8080/health`.
2. В окне `cloudflared` дождись строки `Registered tunnel connection`.
3. На ВМ останови только туннель. Машину не выключай:

```bash
sudo systemctl stop lifeos-tunnel
sudo systemctl disable lifeos-tunnel
```

4. Сразу с Mac:

```bash
curl -fsS https://local-ai-assist.ru/health
```

Открой `https://local-ai-assist.ru/app/`. В Telegram напиши боту `/start` и открой Mini App кнопкой меню. Бот должен ответить, страница Mini App — открыться.

Если публичный `/health` или бот молчат, на ВМ верни коннектор и разберись с Mac, не гася машину:

```bash
sudo systemctl enable --now lifeos-tunnel
```

Саму ВМ в этот вечер не выключай. Диск пусть полежит, пока Mac не ответит на `/health` и бот не ответит на `/start`. Выключение — вручную, из консоли Яндекса, когда возвращаться уже не нужно. Скрипта, который гасит ВМ, нет.

## Сон

Закрытая крышка останавливает Docker и `cloudflared`. Сайт и бот лежат, пока Mac не проснётся. На вечер у розетки: Системные настройки → Аккумулятор → не спать при питании от сети. В рюкзаке это не сервер.
