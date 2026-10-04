# Деплой на Yandex Cloud VM: как заливать секреты

Инструкция без реальных значений, IP и ключей. Все `<...>` — плейсхолдеры, которые вы заполняете сами. Секреты **никогда не коммитятся в git** — только на виртуалку через SSH.

---

## 0. Правила безопасности (обязательно)

1. **Никаких секретов в git.** `.env`, `*.key`, `id_rsa`, `credentials.json` и т.п. закрыты в `.gitignore`. Перед каждым push: `git status` — в staged не должно быть файлов с токенами/паролями.
2. Приватный SSH-ключ хранится **только** на вашей машине (`~/.ssh/…`, права `600`). На ВМ — только **публичная** часть (`~/.ssh/authorized_keys`).
3. Все секреты (bot token, JWT secret, DB password, API keys) — в файле на ВМ вне репозитория, либо в Yandex Lockbox / Secret Store (см. §7).
4. Скомпрометированный секрет = немедленно заменить, затем ротация остальных.
5. После замены ключа/токена — обновить файл на ВМ и перезапустить сервис.

---

## 1. Подключение к ВМ

```bash
# Одноразово: положить свою публичную клавишу на ВМ (пароль от ВМ — из панели YC / от админа)
ssh-copy-id <USER>@<VM_IP>

# Дальше — по ключу
ssh <USER>@<VM_IP>
```

Рекомендуемый `~/.ssh/config` на локальной машине (файл НЕ попадает в git):

```
Host lifeos-yc
    HostName <VM_IP>
    User <USER>
    IdentityFile ~/.ssh/<your_private_key_file>
    ServerAliveInterval 30
```

Проверка: `ssh lifeos-yc 'echo ok'`.

---

## 2. Где живут секреты на ВМ

```
/opt/lifeos/
├── repo/                 # git clone (код, без секретов)
├── .env                  # ← ВСЕ секреты здесь. chmod 600, владелец — деплоятильщик
└── backups/              # дампы БД (chmod 600)
```

`.env` на ВМ — копия структуры `.env.example` из репозитория, но с боевыми значениями.

---

## 3. Первичный деплой ВМ

```bash
ssh lifeos-yc <<'EOF'
set -e
sudo mkdir -p /opt/lifeos/backups
cd /opt/lifeos
git clone <REPO_URL> repo || (cd repo && git pull --ff-only)
cd repo
docker compose -f deployments/docker-compose.yml up -d --build
EOF
```

## 4. Заливка/обновление `.env` с секретами (без копирования в git)

Способ А — пересылка файла по SSH (файл лежит локально, например `~/secrets/lifeos.env`, права 600):

```bash
scp -p ~/secrets/lifeos.env <USER>@<VM_IP>:/opt/lifeos/.env
ssh <USER>@<VM_IP> 'chmod 600 /opt/lifeos/.env'
```

Способ Б — одна строка без промежуточного файла (если нужно поменять один ключ):

```bash
ssh <USER>@<VM_IP> "sed -i 's/^TELEGRAM_BOT_TOKEN=.*/TELEGRAM_BOT_TOKEN=<NEW_TOKEN>/' /opt/lifeos/.env"
```

Способ В — интерактивно на ВМ:

```bash
ssh <USER>@<VM_IP>
nano /opt/lifeos/.env      # вписать значения
chmod 600 /opt/lifeos/.env
```

После любых изменений секретов — рестарт (см. §5/§6).

## 5. Приложение через Docker Compose (рекомендуется)

Compose читает `.env` через `env_file: ../.env` (см. `deployments/docker-compose.yml`). Для продакшена создайте **не-committable** override:

```bash
cp deployments/docker-compose.override.example.yml deployments/docker-compose.override.yml
# отредактируйте image tag, домены — сам .env не трогайте
docker compose -f deployments/docker-compose.yml up -d --build
```

Makefile уже подхватывает `deployments/docker-compose.override.yml` автоматически (переменная `COMPOSE`). Override-файл на ВМ тоже вне git; в репо лежит только `*.example.yml` без секретов.

## 6. Обновление кода без даунтайма

```bash
cd /opt/lifeos/repo
git pull --ff-only
go run ./cmd/lifeos migrate up     # или docker exec app /app/lifeos migrate up
docker compose -f deployments/docker-compose.yml up -d --build
docker compose logs -f app         # убедиться, что поднялось
```

## 7. Продвинутый уровень: Yandex Lockbox / KMS (опционально)

Чтобы вообще не хранить `.env` в открытом виде на диске:

1. Создать секрет в [Yandex Cloud Console](https://console.yandex.cloud/) → каталог LifeOS → Secret `lifeos-env`.
2. Выдать сервисному аккаунту ВМ роль `lockbox.viewer`.
3. На ВМ скрипт запуска тянет секрет перед стартом:

```bash
yc lockbox payload get <SECRET_ID> --folder-id <FOLDER_ID> \
  > /opt/lifeos/.env && chmod 600 /opt/lifeos/.env
systemctl restart lifeos   # или docker compose up -d
```

Ротация секрета в Lockbox + рестарт сервиса — код и файлы не трогаются.

## 8. Чек-лист после деплоя

- [ ] `curl -fsS http://localhost:8080/health` — OK
- [ ] `GET /metrics` отвечает
- [ ] Бот отвечает в Telegram (polling запущен, токен подхвачен)
- [ ] Mini App открывается по `LIFEOS_MINIAPP_URL` (HTTPS)
- [ ] `/opt/lifeos/.env` имеет права `600` и **нет** его копии в git (`git ls-files | grep -i '\.env$'` пусто)
- [ ] Резервная копия: `make backup` на ВМ → `/opt/lifeos/backups/`

## 9. Чего делать нельзя

- ❌ Коммитить `.env`, ключи, токены — даже «временно».
- ❌ Хранить приватный SSH-ключ внутри репозитория или на ВМ.
- ❌ Передавать секреты по HTTP/в чатах/в описаниях задач агентам.
- ❌ Делать `docker cp` секретов в образ.
- ❌ Публиковать вывод `cat /opt/lifeos/.env` в отчётах/скриншотах.
