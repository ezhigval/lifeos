# Край: Cloudflare

Публичный вход один — named tunnel Cloudflare на `http://127.0.0.1:8080`. Отдельного HTTP-сервера на ВМ нет, HTTPS на машине не терминируется.

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
| Образ приложения | pull и запуск | нет | Actions: GHCR и prerelease `vm-image` |

Снаружи на ВМ открыт только SSH 22. Порты 80, 443, 5432, 5433 и 8080 в firewall и в группе безопасности не открывать. Туннель сам устанавливает исходящее соединение.

`/opt/lifeos/docker-compose.override.yml` перекрывает порты из compose и в git не входит. Postgres в нём — `127.0.0.1:5433:5432`. Приложение — `127.0.0.1:8080:8080`. Выкладка сама переписывает `8080:8080` и `0.0.0.0:8080:8080` на loopback. Порт 8080 в группу безопасности не открывать: снаружи к нему ходит только туннель.

Приложение уже принимает webhook: маршрута `POST /webhook/telegram` нет только если процесс запущен без него. На ВМ режим задаёт `LIFEOS_TELEGRAM_MODE=webhook`, URL собирает `deployments/apply-public-origin.sh`. Предпочтительный путь: Telegram → Cloudflare → туннель → этот POST. Токен бота не ротировать. `lifeos-tg-proxy` не удалять.

Образ приложения собирает workflow CI на push в `main`. ВМ его не компилирует. Prerelease `vm-image` не должен становиться Latest.

Desktop-обновлятор читает `https://github.com/ezhigval/lifeos/releases/latest/download/latest.json`. Zip в этом JSON лежит на том же хосте `github.com`, иначе проверка источника его отбросит. Файл появится в релизе со следующим тегом `desktop-v*` (workflow `.github/workflows/desktop-release.yml`). Пока в текущем релизе только zip. `releases/latest` должен указывать на такой тег. Уже установленные сборки ходят на старый URL ВМ, пока их один раз не обновят с Releases.

## Если публичный /health пропал

Смотри приложение и туннель.

```bash
curl -fsS http://127.0.0.1:8080/health
systemctl is-active lifeos-tunnel
```

В Zero Trust у Public Hostname сервис должен остаться `http://127.0.0.1:8080`. Отдельный прокси на `:80` не поднимать.

## Клики в панели

Токенов и паролей здесь нет. Имена — твои.

1. Public hostname туннеля. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → Networks → Tunnels → туннель LifeOS → Configure → Public Hostname. Тип **HTTP**, URL **`127.0.0.1:8080`**. Не `:80`. Hostname без пути, например `lifeos.example.com`.
2. Оранжевое облако. [dash.cloudflare.com](https://dash.cloudflare.com/) → сайт → DNS → Records. CNAME, который создал туннель, должен быть **Proxied**. A-запись на IP ВМ не добавлять.
3. Кэш Mini App. Тот же сайт → Rules → Cache Rules → Create rule. Три правила, обход выше кэша ассетов. Выражения и действия лежат в `deployments/vps/cache-rules.json`.
   - `/app/assets/*` — cache, edge TTL один год. Origin уже шлёт `public, max-age=31536000, immutable`.
   - `/app`, `/app/`, `/app/index.html`, `/app/sw.js` — bypass. Origin уже шлёт `no-store`. `/app/` — это адрес, который открывает Telegram.
   - путь с `/api` или `/webhook` — bypass.
4. Исходящий воркер, если его ещё нет в панели: Workers → тот, что собран из `tg-proxy-worker.js`. Маршрут воркера на Mini App не вешать.
