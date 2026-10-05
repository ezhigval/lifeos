# Стабильный HTTPS: домен + named Cloudflare Tunnel

Инструкция для текущей ВМ LifeOS. Секретов в этом файле нет. Плейсхолдеры вида `<...>` заполняешь ты. В git уходят только скрипты и пустые примеры.

Quick-tunnel `*.trycloudflare.com` здесь не используем: с ВМ не открывается `api.trycloudflare.com:443`, а URL всё равно меняется при рестарте.

## Схема

```
Telegram  --webhook-->  Cloudflare (HTTPS, твой домен)
                              |
                         named tunnel (исходящий с ВМ, TCP 7844)
                              |
                         127.0.0.1:8080  LifeOS

LifeOS  --ответ бота-->  127.0.0.1:8081 tg-proxy
                              |
                         Cloudflare Worker
                              |
                         api.telegram.org
```

Туннель только входящий: Mini App, `/health`, webhook. Исходящий Bot API через воркер, потому что с ВМ до Telegram напрямую нет маршрута. Имя хоста не меняется при рестарте `cloudflared` и при пересборке контейнера.

A-запись на IP ВМ не делаем. Входящие 80/8080 снаружи нестабильны, а сертификат на самой ВМ эту проблему не лечит.

## Сколько стоит

| Что | Зачем | Деньги |
|-----|--------|--------|
| Домен `.ru` | стабильное имя | первый год часто 169–199 ₽, продление обычно 700–850 ₽ |
| Домен `.com` | то же | порядка 1500–1800 ₽/год |
| Cloudflare DNS + Tunnel + Worker | NS, HTTPS, туннель, исходящий прокси | 0 ₽ на free |
| Зона Cloud DNS в Yandex | не нужна, NS будут на Cloudflare | не покупать (~40 ₽/мес) |
| Certificate Manager | не нужен, TLS терминирует Cloudflare | 0 ₽ |

Регистратор любой нормальный (Yandex 360 / Рег.ру / Timeweb). Важно только, чтобы можно было сменить NS.

## Что секрет, а что можно написать в чат

| Значение | Куда класть | В чат Cursor | В git / PR / issue |
|----------|-------------|--------------|--------------------|
| Имя домена, hostname | публично | да | да, это не секрет |
| URL воркера `https://….workers.dev` | `/opt/lifeos/.env` | да | нет, только плейсхолдер |
| Токен коннектора туннеля | `/opt/lifeos/secrets/tunnel.env` `chmod 600` | лучше не надо, см. ниже | никогда |
| `PROXY_SECRET` | воркер + `/opt/lifeos/secrets/tg-proxy.env` | лучше не надо | никогда |
| Токен бота, JWT, пароль БД | уже в `/opt/lifeos/.env` | не присылать повторно | никогда |

Предпочтительный путь для токена туннеля — записать его на ВМ самому и написать в чат только «токен лежит в `/opt/lifeos/secrets/tunnel.env`». Если передашь токен в чат агенту, он имеет право записать его только в этот файл и в окружение сервиса. Коммитить, класть в PR, в issue и в пример в репозитории нельзя.

Проверка перед любым push:

```bash
bash scripts/check-no-secrets.sh
git status
```

CI гоняет тот же скрипт. `.gitignore` закрывает `.env`, `secrets/`, `tunnel.env`, `tg-proxy.env`, ключи и `docker-compose.override.yml`.

## 1. Домен и NS

1. Купи домен.
2. Заведи бесплатный аккаунт Cloudflare и добавь сайт (план Free).
3. Cloudflare покажет два NS. Пропиши их у регистратора вместо прежних.
4. Подожди, пока статус зоны станет Active. Обычно минуты, иногда несколько часов.

Пока NS не переключились, public hostname туннеля не зарезолвится.

## 2. Named tunnel с ноутбука

Создавать туннель нужно с машины, где открыт HTTPS к API Cloudflare. С ВМ этот API на 443 не открывается, quick-tunnel тоже.

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → Networks → Tunnels → Create a tunnel.
2. Имя, например `lifeos`. Тип коннектора: cloudflared.
3. На шаге Install connector скопируй токен. Это длинная строка, обычно начинается с `eyJ`. Сохрани её в менеджер паролей.
4. Не ставь коннектор на ноутбук как единственную копию. Коннектор будет на ВМ.
5. Public Hostname:
   - Subdomain / domain: тот hostname, который хочешь, например `lifeos.<твой-домен>` или корень домена
   - Type: **HTTP**
   - URL: `http://127.0.0.1:8080`
6. Сохрани. Дополнительный catch-all не нужен: у remotely-managed tunnel он уже есть.

Hostname должен совпасть с тем, что потом пойдёт в `LIFEOS_PUBLIC_ORIGIN`. Без пути, без слэша на конце: `https://lifeos.example.com`.

Туннель ходит на edge по TCP 7844, HTTP/2, только IPv4. Так и запускает `deployments/cloudflared/run-tunnel.sh`. QUIC и IPv6 с этой ВМ не используем.

## 3. Воркер (исходящий Bot API)

Код: `deployments/vps/tg-proxy-worker.js`. Его надо задеплоить заново: старая копия пускает только пути `/bot…` и отрезает скачивание файлов `/file/bot…`.

С ноутбука:

```bash
cd deployments/vps
npx --yes wrangler login
npx --yes wrangler deploy
```

Либо вставь файл в дашборде Workers → твой воркер → Deploy. Запомни URL `https://<имя>.<аккаунт>.workers.dev`.

Рекомендуется общий секрет, чтобы воркером не пользовались чужие:

```bash
npx wrangler secret put PROXY_SECRET
```

То же значение потом попадёт в `/opt/lifeos/secrets/tg-proxy.env` как `LIFEOS_TG_PROXY_SECRET`. В репозиторий его не клади. Если секрет на воркере есть, а на ВМ нет, `test` получит 401.

## 4. Что прислать после покупки

Напиши в чат:

1. Публичный origin, например `https://lifeos.example.com`.
2. URL воркера.
3. Одно из двух:
   - «токен туннеля уже в `/opt/lifeos/secrets/tunnel.env`, права 600»;
   - или сам токен в чат, если записать на ВМ сам не хочешь.
4. Включён ли `PROXY_SECRET`. Если да — либо «секрет уже в `/opt/lifeos/secrets/tg-proxy.env`», либо пришли его отдельно от git.

Команды, если токен кладёшь сам (значение не остаётся в истории, если вызвать через `read -rs`):

```bash
ssh <USER>@<VM>
sudo install -d -m 700 /opt/lifeos/secrets
read -rs TUNNEL_TOKEN
printf 'TUNNEL_TOKEN=%s\n' "$TUNNEL_TOKEN" | sudo tee /opt/lifeos/secrets/tunnel.env >/dev/null
unset TUNNEL_TOKEN
sudo chmod 600 /opt/lifeos/secrets/tunnel.env
```

`printf` в этой сессии видит токен. В git он не попадает: каталог `/opt/lifeos/secrets` не является клоном репозитория.

## 5. Что ставится на ВМ

Код должен быть на `main`, чтобы автодеплой пересобрал образ. В этом образе клиент Telegram при `LIFEOS_HTTP_PROXY` ходит на `http://api.telegram.org`. Иначе Go открывает CONNECT, локальный прокси запрос не видит, а в лог ошибки попадает URL с токеном бота. Ошибки клиента токен больше не печатают.

После merge таймер `lifeos-deploy.timer` сам подхватит коммит. Пока SHA в `/opt/lifeos/.last_deploy_sha` не равен новому `main`, образ старый.

Дальше на ВМ, из `/opt/lifeos/repo`:

```bash
sudo bash deployments/cloudflared/install-named-tunnel.sh
sudo LIFEOS_TG_PROXY_WORKER_URL='https://<worker>' \
  bash deployments/vps/tg-proxy.sh install
sudo bash deployments/vps/tg-proxy.sh test
sudo LIFEOS_PUBLIC_ORIGIN='https://lifeos.example.com' \
  bash deployments/apply-public-origin.sh
```

Если секрет воркера включён, добавь его в окружение только на время install:

```bash
sudo LIFEOS_TG_PROXY_WORKER_URL='https://<worker>' \
  LIFEOS_TG_PROXY_SECRET='<секрет>' \
  bash deployments/vps/tg-proxy.sh install
```

Скрипт пишет секрет в `/opt/lifeos/secrets/tg-proxy.env` и убирает его из своего окружения. В argv python он не передаётся.

Что делают скрипты:

| Скрипт | Действие |
|--------|----------|
| `install-named-tunnel.sh` | `cloudflared` при отсутствии, unit `lifeos-tunnel.service`, старт только если токен уже в файле |
| `tg-proxy.sh install` | локальный прокси, `LIFEOS_HTTP_PROXY` в `/opt/lifeos/.env`, `extra_hosts: host.docker.internal` в `/opt/lifeos/docker-compose.override.yml` (этот файл вне git) |
| `tg-proxy.sh test` | `getMe` через `127.0.0.1:8081`, в вывод только ok / id / username |
| `apply-public-origin.sh` | `MODE=webhook`, Mini App `/app/`, webhook `/webhook/telegram`, секрет webhook генерируется, если его ещё нет; контейнер пересоздаётся; `setWebhook` и кнопка меню идут через локальный прокси |

Прокси слушает `127.0.0.1` и, если есть сеть `lifeos_default`, IP шлюза этой сети. Публичный интерфейс ВМ он не занимает. Контейнер ходит на `http://host.docker.internal:8081`.

`apply-public-origin.sh` отказывается править env-файл, если его реальный путь лежит внутри git-клона.

## 6. Проверка

На ВМ:

```bash
curl -fsS http://127.0.0.1:8080/health
systemctl is-active lifeos-tunnel lifeos-tg-proxy
sudo bash /opt/lifeos/repo/deployments/vps/tg-proxy.sh test
```

С любой машины, где резолвится домен:

```bash
curl -fsS https://lifeos.example.com/health
curl -fsS -o /dev/null -w '%{http_code}\n' https://lifeos.example.com/app/
```

В Telegram отправь боту `/start`. Mini App открывай кнопкой Menu или новой кнопкой внизу, не старой ссылкой из истории чата.

`apply-public-origin.sh` в конце печатает `webhook_url`, `pending` и `last_error`. `webhook_url` должен быть `https://<hostname>/webhook/telegram`. `last_error` пустой после первого успешного апдейта.

Локальный health обязателен. Публичный health с самой ВМ может не открыться (hairpin через Cloudflare). Это не провал, если с ноутбука `curl` на тот же URL отвечает.

## 7. Если воркер с ВМ не открывается

`tg-proxy.sh test` пишет `getMe via local proxy failed` и не показывает токен. Тогда с ВМ нет HTTPS до `*.workers.dev` (часть адресов Cloudflare на 443 режется, при этом TCP 7844 до edge туннеля живой).

Туннель и Mini App при этом всё равно работают. Не работают ответы бота. Следующий шаг — релей на хосте, который ВМ открывает по 443 и который сам достаёт до `api.telegram.org` или до воркера. URL релея подставляется вместо URL воркера в `LIFEOS_TG_PROXY_WORKER_URL`, контракт тот же: `GET/POST /fetch?path=`. Пока тест не упал, релей не нужен.

## 8. Чего не делать

- Не коммитить `/opt/lifeos/.env`, `secrets/`, `tunnel.env`, `docker-compose.override.yml`.
- Не вставлять токен туннеля в issue, PR, скриншот `journalctl` и пример в репозитории.
- Не возвращать quick-tunnel в `LIFEOS_MINIAPP_URL`.
- Не логировать `docker compose config` и `docker inspect` наружу: там видны переменные приложения.
- Не запускать скрипты с `bash -x`: они делают `set +x`, но родительский trace всё равно может напечатать окружение.
