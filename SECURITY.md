# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.0.x (alpha) | ✅ |
| < 1.0 | ❌ |

LifeOS — персональное приложение с моделью угроз «один доверенный пользователь на своём железе». Тем не менее публичные инстансы (Telegram Mini App, REST API) требуют соблюдения правил ниже.

## Reporting a Vulnerability

**Не создавайте публичные issue с уязвимостями.** Сообщайте о проблемах безопасности приватно:

- GitHub Security Advisory: [Report a vulnerability](https://github.com/ezhigval/lifeos/security/advisories/new)
- E-mail maintainer'а указан в репозитории.

Мы постараемся подтвердить проблему в течение 72 часов и выпустить исправление в разумный срок. Не раскрывайте детали до публикации патча.

## Deployment Hardening Checklist

1. **Секреты никогда не коммитятся.** Используйте `.env` (вне VCS) или `/opt/lifeos/secrets`. Проверка CI: `bash scripts/check-no-secrets.sh`.
2. **JWT**: `LIFEOS_JWT_SECRET` — случайные байты длиной ≥ 32 (`openssl rand -base64 48`). Токены короткоживущие; refresh через Telegram login codes.
3. **Прослушивание**: по умолчанию сервер слушает `127.0.0.1:8080`. Публичный доступ — только через TLS-терминатор (Caddy / Cloudflare Tunnel), см. [docs/deploy/EDGE.md](docs/deploy/EDGE.md).
4. **PostgreSQL** не публикуется наружу; в Docker Compose порт проброшен только на localhost (`127.0.0.1:5433`).
5. **LLM-провайдер**: при включённом `LIFEOS_LLM_ENABLED` данные уходят во внешний API. Анонимное обучение (`learning_opt_in`) отключено по умолчанию и пишет только HMAC-субъект и грубые флаги без PII ([DOMAIN_MODEL.md](docs/architecture/DOMAIN_MODEL.md#learning)).
6. **Telegram webhook-режим**: проверяйте `secret_token` (middleware) и ограничивайте источник, если possible.
7. Обновляйте зависимости: автоматические PR на обновления — `.github/dependabot.yml`.

## Scope

В зоне безопасности: аутентификация (JWT, Telegram initData, login codes), авторизация (single-tenant `user_id` isolation на уровне запросов), обработка вложений (STT/Vision), логирование (маскирование токенов).
