# ADR-004: Telegram Long Polling for MVP

## Status
Accepted (revisit for production webhook)

## Context
Домашний Mac со статическим IP. Webhook требует HTTPS, port forward, сертификат. Polling проще для старта.

## Decision
MVP использует **long polling**. Webhook — Phase 1.5 после стабилизации.

## Consequences
**+** Нет возни с TLS/port forward на старте  
**+** Работает за NAT  
**−** Чуть выше latency, постоянное соединение  
**−** Менее «production classic» чем webhook

## Migration Path
`LIFEOS_TELEGRAM_MODE` env переключает adapter. Handler layer идентичен.

## Note
Статический IP сохраняем для будущего webhook + возможного REST API.
