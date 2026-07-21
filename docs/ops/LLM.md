# LLM (опционально)

Rule-based resolver всегда первый. LLM вызывается только для `unknown` интентов и для prose-обзоров. При ошибке/таймауте (~8s) — деградация на rule-based / template **с логом** `llm … degraded`.

`dispatchIntent` после резолва вызывает use cases — LLM не обходит домен.

## Включение

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_PROVIDER=openai
LIFEOS_LLM_API_KEY=...
LIFEOS_LLM_BASE_URL=https://api.groq.com/openai/v1
LIFEOS_LLM_MODEL=llama-3.3-70b-versatile
```

Alias ключа: `LIFEOS_OPENAI_API_KEY` (если `LIFEOS_LLM_API_KEY` пуст).

Провайдер по умолчанию: `openai` (любой OpenAI-compatible endpoint). Локально: `LIFEOS_LLM_PROVIDER=ollama`.

`serve` **fail-closed**:
- пустой `TELEGRAM_BOT_TOKEN` / `LIFEOS_JWT_SECRET` → ошибка старта (обход: `LIFEOS_ALLOW_NO_TELEGRAM` / `LIFEOS_ALLOW_NO_API`)
- openai без API key → ошибка старта
- `scripts/mock_ollama.go` (модель `lifeos_mock` в `/api/tags`) → ошибка, пока нет `LIFEOS_ALLOW_MOCK_LLM=true`

---

## Бесплатные API

### #1 — Groq (рекомендуется для Telegram)

- Ключ: https://console.groq.com/keys (без карты)
- `BASE_URL=https://api.groq.com/openai/v1`
- `MODEL=llama-3.3-70b-versatile`
- Free: ~30 RPM, ~1000 RPD
- Плюс: низкая латентность

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_PROVIDER=openai
LIFEOS_LLM_API_KEY=gsk_...
LIFEOS_LLM_BASE_URL=https://api.groq.com/openai/v1
LIFEOS_LLM_MODEL=llama-3.3-70b-versatile
```

### #2 — Google AI Studio (Gemini)

- Ключ: https://aistudio.google.com/apikey
- OpenAI-compat: `https://generativelanguage.googleapis.com/v1beta/openai/`
- Модель: `gemini-2.0-flash` (или актуальная flash — проверить в AI Studio)
- Щедрый free tier

```env
LIFEOS_LLM_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai/
LIFEOS_LLM_MODEL=gemini-2.0-flash
LIFEOS_LLM_API_KEY=...
```

### #3 — OpenRouter free

- Ключ: https://openrouter.ai/keys
- `BASE=https://openrouter.ai/api/v1`
- Модель с суффиксом `:free`, напр. `meta-llama/llama-3.3-70b-instruct:free`
- Лимит ниже (~50 req/day free)

```env
LIFEOS_LLM_BASE_URL=https://openrouter.ai/api/v1
LIFEOS_LLM_MODEL=meta-llama/llama-3.3-70b-instruct:free
LIFEOS_LLM_API_KEY=...
```

### #4 — Local Ollama

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_PROVIDER=ollama
LIFEOS_OLLAMA_URL=http://localhost:11434
LIFEOS_OLLAMA_MODEL=llama3.2
```

Compose по умолчанию Ollama не поднимает.

### Dev stub (только локально)

Если реальный Ollama/llama.cpp падает на inference:

```bash
go run ./scripts/mock_ollama.go
# + в .env:
LIFEOS_ALLOW_MOCK_LLM=true
LIFEOS_LLM_PROVIDER=ollama
```

Без `LIFEOS_ALLOW_MOCK_LLM` `serve` откажется стартовать при обнаружении `lifeos_mock`.

---

## Поток

```
текст → rulebased → known? → dispatchIntent (use cases)
                   → unknown + LLM on? → openaiapi/ollama → known? → dispatchIntent
                                                              → unknown → UX «не понял»
обзор → LLM (sanitize <b>) → template fallback (logged)
```

## Диалоговый агент

См. [docs/ai/AGENT.md](../ai/AGENT.md) — multi-turn, tools, память, anon learning.
`LIFEOS_LLM_AGENT_ENABLED=true` (default).

## Speech-to-text (голос / кружочки)

Telegram `voice` / `video_note` / `audio` → download → Whisper → тот же agent path.

```env
LIFEOS_STT_ENABLED=true
# optional overrides (defaults reuse LLM key + Groq Whisper):
# LIFEOS_STT_API_KEY=...
# LIFEOS_STT_BASE_URL=https://api.groq.com/openai/v1
# LIFEOS_STT_MODEL=whisper-large-v3-turbo
```

Рекомендуется тот же Groq ключ, что и для LLM. Ollama LLM + cloud Whisper — нормальная связка.
