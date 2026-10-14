# Windows дома: Docker и тот же домен

Прод без аренды. Машина — твой ПК с Windows. Яндекс остановлен, в этой схеме его нет. Белого IP нет и не нужно: наружу ходит только named tunnel Cloudflare.

Пока ПК спит, сайт и бот молчат.

```
Telegram  --webhook-->  Cloudflare, домен local-ai-assist.ru
                            |
                       named tunnel (исходящий с Windows)
                            |
                       127.0.0.1:8080  Docker, LifeOS

LifeOS --ответ бота-->  host.docker.internal:8081
                            |
                       прокси на Windows
                            |
                       Cloudflare Worker
                            |
                       api.telegram.org
```

Вход: Telegram → Cloudflare → туннель → Docker. Ответ: Docker → прокси → Cloudflare → Telegram. A-запись на домашний адрес не добавлять, порты на роутере не открывать. Public Hostname в Zero Trust не менять: HTTP, `127.0.0.1:8080`.

Команды ниже выполняются на этом Windows. Агент в изолированной среде до ПК не дотягивается.

## 1. Docker Desktop

1. [Docker Desktop для Windows](https://www.docker.com/products/docker-desktop/). Движок WSL2.
2. Дождись Running.
3. В PowerShell: `docker info`. Должна быть строка Server Version.

## 2. Секреты

Токен бота, JWT, секрет webhook и токен туннеля в чат, в git и в скриншот не класть. Каталог вне репозитория:

```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\.config\lifeos" | Out-Null
```

Файл `.env` в корне клона. Он уже в `.gitignore`. Compose читает его как `../.env`. Если копия со старой машины сохранилась — положи её сюда и поправь одну строку прокси (шаг 4). Если копии нет:

- токен бота заново у @BotFather, это тот же бот;
- туннель: Zero Trust → Networks → Tunnels → тот же туннель → новый token коннектора. Старый умер вместе с Яндексом;
- `LIFEOS_JWT_SECRET` и `LIFEOS_TELEGRAM_WEBHOOK_SECRET` — новые, `openssl` не обязателен: 32+ случайных байта. Старые сессии Mini App при новом JWT отвалятся;
- `LIFEOS_TELEGRAM_MODE=webhook`;
- `LIFEOS_MINIAPP_URL=https://local-ai-assist.ru/app/`;
- `LIFEOS_TELEGRAM_WEBHOOK_URL=https://local-ai-assist.ru/webhook/telegram`.

База в новом Docker пустая. Старые записи лежат на диске остановленной ВМ, пока диск не удалят. Этот файл их не забирает.

## 3. Compose

Из корня репозитория, который открыт в Cursor. `make` на Windows не нужен. Файл `docker-compose.override.yml` от ВМ сюда не копировать.

```powershell
docker compose -f deployments/docker-compose.yml up -d --build
curl.exe -fsS http://127.0.0.1:8080/health
```

Слушают только `127.0.0.1:8080` и `127.0.0.1:5433`. Образ собирается на этом ПК. Первая сборка качает базовые образы, домашний интернет это умеет.

## 4. Ответ бота через Cloudflare

Прокси — `deployments/vps/tg-proxy.py`. URL воркера и его секрет лежат в файле вне git, например `%USERPROFILE%\.config\lifeos\tg-proxy.env`. В PowerShell их в окружение, не на экран:

```powershell
Get-Content "$env:USERPROFILE\.config\lifeos\tg-proxy.env" | ForEach-Object {
  if ($_ -match '^(LIFEOS_TG_PROXY_WORKER_URL|LIFEOS_TG_PROXY_SECRET)=(.*)$') {
    Set-Item -Path "Env:$($Matches[1])" -Value $Matches[2]
  }
}
py -3 deployments/vps/tg-proxy.py
```

Окно не закрывать. В `.env` приложения:

```env
LIFEOS_HTTP_PROXY=http://host.docker.internal:8081
```

`127.0.0.1` внутри контейнера — это сам контейнер, не Windows. Имя `host.docker.internal` Docker Desktop уже резолвит. После правки `.env`:

```powershell
docker compose -f deployments/docker-compose.yml up -d
```

## 5. Туннель

`cloudflared-windows-amd64.exe` с релизов GitHub, переименовать в `cloudflared.exe` и положить в PATH, не в клон. Токен — одна строка `TUNNEL_TOKEN=...` в `%USERPROFILE%\.config\lifeos\tunnel.env`.

Git Bash, если он есть от Git for Windows:

```bash
LIFEOS_TUNNEL_ENV="$USERPROFILE/.config/lifeos/tunnel.env" bash deployments/cloudflared/run-tunnel.sh
```

Без Bash, в том же духе, что прокси: прочитать строку в `$env:TUNNEL_TOKEN` и не печатать её.

```powershell
cloudflared --no-autoupdate --protocol http2 --edge-ip-version 4 tunnel run
```

`install-named-tunnel.sh` — это systemd для Linux. На Windows он не нужен. В Zero Trust сервис hostname остаётся HTTP `http://127.0.0.1:8080`. Второй коннектор не поднимать.

## 6. Проверка

```powershell
curl.exe -fsS http://127.0.0.1:8080/health
curl.exe -fsS https://local-ai-assist.ru/health
```

Браузер: `https://local-ai-assist.ru/app/`. В Telegram: `/start` и кнопка Mini App. Оба curl должны ответить, и бот должен ответить, пока открыты окна `cloudflared` и прокси.

## Сон

Сон Windows гасит Docker и туннель. На ночь у розетки: Параметры → Система → Питание → экран можно гасить, сон — никогда, пока это прод.

## Позже: модель на этом же ПК

Не включать, пока бот не отвечает. Выбор один из двух, оба ставятся на Windows, не в контейнер LifeOS.

Ollama, порт 11434:

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_PROVIDER=ollama
LIFEOS_OLLAMA_URL=http://host.docker.internal:11434
LIFEOS_OLLAMA_MODEL=llama3.2
```

LM Studio, локальный сервер, OpenAI-совместимый порт 1234. Ключ для LM Studio любой непустой, это не секрет провайдера:

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_PROVIDER=openai
LIFEOS_LLM_API_KEY=lm-studio
LIFEOS_LLM_BASE_URL=http://host.docker.internal:1234/v1
LIFEOS_LLM_MODEL=имя-модели-из-LM-Studio
```

`LIFEOS_LLM_AGENT_ENABLED=true` только вместе с `LIFEOS_LLM_ENABLED=true`. Пока оба флага `false`.
