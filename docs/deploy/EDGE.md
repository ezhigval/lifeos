# Край: Cloudflare, не nginx

Публичный вход один — Cloudflare. nginx на ВМ не стоит перед приложением и не терминирует HTTPS. Второй HTTPS-стек на ВМ не поднимается.

```
Mini App браузер  -- GET /app/... --------------> Cloudflare
   кэш края только для /app/assets/*
Telegram          -- POST /webhook/telegram ----> Cloudflare  (HTTPS, оранжевое облако)
Desktop           -- не на ВМ и не на этот хост -> GitHub Releases
                                                      |
                                                 named tunnel
                                                 исходящий TCP 7844 с ВМ
                                                      |
                                                 http://127.0.0.1:8080
                                                 Go-приложение
                                                      |
                                                 Postgres 127.0.0.1:5433
                                                 системный Postgres уже на 127.0.0.1:5432

Приложение -- Bot API --> 127.0.0.1:8081  lifeos-tg-proxy
                                |
                           Cloudflare Worker
                           deployments/vps/tg-proxy-worker.js
                                |
                           api.telegram.org
```

Воркер нужен только потому, что Yandex Cloud не пускает на `api.telegram.org`. Это не путь Mini App. `wrangler.toml` деплоит только его. Правила кэша — отдельный файл `deployments/vps/cache-rules.json`, wrangler его не выкладывает.

## Что где живёт

| | ВМ | Cloudflare | GitHub |
|---|----|------------|--------|
| Go и Postgres | да | нет | нет |
| HTTPS и DNS | нет | да | нет |
| Статика Mini App | origin на `:8080` | кэш `/app/assets/*` | нет |
| Входящий webhook | `POST /webhook/telegram` | проксирует в туннель | нет |
| Исходящий Bot API | прокси `:8081` | Worker | нет |
| Сборки desktop | нет | нет, бакета R2 в workflow нет | Releases |

Снаружи на ВМ открыт только SSH 22. Порты 80, 443, 5432, 5433 и 8080 в firewall и в группе безопасности не открывать. Туннель сам устанавливает исходящее соединение.

Приложение уже принимает webhook: маршрута `POST /webhook/telegram` нет только если процесс запущен без него. На ВМ режим задаёт `LIFEOS_TELEGRAM_MODE=webhook`, URL собирает `deployments/apply-public-origin.sh`. Предпочтительный путь: Telegram → Cloudflare → туннель → этот POST. Токен бота не ротировать. `lifeos-tg-proxy` не удалять.

Desktop-обновлятор читает `https://github.com/ezhigval/lifeos/releases/latest/download/latest.json`. Zip в этом JSON лежит на том же хосте `github.com`, иначе проверка источника его отбросит. Файл появится в релизе со следующим тегом `desktop-v*` (workflow `.github/workflows/desktop-release.yml`). Пока в текущем релизе только zip. `releases/latest` должен указывать на такой тег. Уже установленные сборки ходят на старый URL ВМ, пока их один раз не обновят с Releases.

## Откат

Если после остановки nginx публичный `/health` пропал: `sudo systemctl enable --now nginx`, в Zero Trust у Public Hostname верни сервис на `http://127.0.0.1:80`, проверь `https://<hostname>/health`. Чтобы снова уйти с nginx: поставь сервис `http://127.0.0.1:8080`, дождись `curl -fsS http://127.0.0.1:8080/health` и того же ответа с публичного имени, затем `sudo systemctl disable --now nginx`.

## Клики в панели

Токенов и паролей здесь нет. Имена — твои.

1. Public hostname туннеля. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → Networks → Tunnels → туннель LifeOS → Configure → Public Hostname. Тип **HTTP**, URL **`127.0.0.1:8080`**. Не `:80` и не nginx. Hostname без пути, например `lifeos.example.com`.
2. Оранжевое облако. [dash.cloudflare.com](https://dash.cloudflare.com/) → сайт → DNS → Records. CNAME, который создал туннель, должен быть **Proxied**. A-запись на IP ВМ не добавлять.
3. Кэш Mini App. Тот же сайт → Rules → Cache Rules → Create rule. Три правила, обход выше кэша ассетов. Выражения и действия лежат в `deployments/vps/cache-rules.json`.
   - `/app/assets/*` — cache, edge TTL один год. Origin уже шлёт `public, max-age=31536000, immutable`.
   - `/app`, `/app/`, `/app/index.html`, `/app/sw.js` — bypass. Origin уже шлёт `no-store`. `/app/` — это адрес, который открывает Telegram.
   - путь с `/api` или `/webhook` — bypass.
4. Исходящий воркер, если его ещё нет в панели: Workers → тот, что собран из `tg-proxy-worker.js`. Маршрут воркера на Mini App не вешать.
