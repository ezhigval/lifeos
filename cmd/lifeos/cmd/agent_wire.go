package cmd

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/agent"
	"github.com/valentinezhov/lifeos/internal/ai/dialogue"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/openaiapi"
	learningapp "github.com/valentinezhov/lifeos/internal/learning/app"
	learninginfra "github.com/valentinezhov/lifeos/internal/learning/infra"
	memoryapp "github.com/valentinezhov/lifeos/internal/memory/app"
	memoryinfra "github.com/valentinezhov/lifeos/internal/memory/infra"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func newChatModel(cfg config.Config, log *slog.Logger) ai.ChatModel {
	if cfg.LLMProvider == "ollama" {
		log.Info("llm chat model", "provider", "ollama", "base_url", cfg.OllamaURL, "model", cfg.OllamaModel)
		return ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	}
	log.Info("llm chat model", "provider", "openai", "base_url", cfg.LLMBaseURL, "model", cfg.LLMModel)
	return openaiapi.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
}

func newAgentService(cfg config.Config, log *slog.Logger, p *pgxpool.Pool, tools toolDeps) *dialogue.Service {
	if !cfg.LLMEnabled || !cfg.LLMAgentEnabled {
		return nil
	}
	memRepo := memoryinfra.NewRepository(p)
	learnRepo := learninginfra.NewRepository(p)
	reg := agent.NewRegistry()
	tools.upsertMemory = memoryapp.NewUpsertMemory(memRepo)
	tools.recallMemory = memoryapp.NewRecall(memRepo)
	registerAgentTools(reg, tools)

	model := newChatModel(cfg, log)
	modelName := cfg.LLMModel
	if cfg.LLMProvider == "ollama" {
		modelName = cfg.OllamaModel
	}
	log.Info("conversational agent enabled", "provider", cfg.LLMProvider, "model", modelName)
	return &dialogue.Service{
		Agent:        agent.New(model, reg),
		ListMemories: memoryapp.NewListMemories(memRepo),
		Privacy:      memRepo,
		RecordLearn:  learningapp.NewRecordEvent(learnRepo),
		LearningSalt: cfg.LearningSalt,
		ModelName:    modelName,
	}
}
