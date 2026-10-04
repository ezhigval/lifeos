package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/valentinezhov/lifeos/internal/platform/config"
	tg "github.com/valentinezhov/lifeos/internal/transport/telegram"
)

var telegramCmd = &cobra.Command{
	Use:   "telegram",
	Short: "Telegram bot management commands",
}

var telegramSetWebhookCmd = &cobra.Command{
	Use:   "set-webhook",
	Short: "Register Telegram webhook (works from a network where api.telegram.org is reachable)",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runTelegramSetWebhook())
	},
}

var telegramWebhookStatusCmd = &cobra.Command{
	Use:   "webhook-status",
	Short: "Show current Telegram webhook registration",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runTelegramWebhookStatus())
	},
}

var telegramDeleteWebhookCmd = &cobra.Command{
	Use:   "delete-webhook",
	Short: "Remove Telegram webhook and drop pending updates",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runTelegramDeleteWebhook())
	},
}

func init() {
	rootCmd.AddCommand(telegramCmd)
	telegramCmd.AddCommand(telegramSetWebhookCmd, telegramWebhookStatusCmd, telegramDeleteWebhookCmd)
}

func newTelegramClientFromConfig() (*tg.Client, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, config.Config{}, err
	}
	if cfg.TelegramBotToken == "" {
		return nil, cfg, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}
	return tg.NewClient(cfg.TelegramBotToken), cfg, nil
}

func runTelegramSetWebhook() error {
	client, cfg, err := newTelegramClientFromConfig()
	if err != nil {
		return err
	}
	url := cfg.TelegramWebhookURL
	secret := cfg.TelegramWebhookSecret
	if url == "" || secret == "" {
		return fmt.Errorf("LIFEOS_TELEGRAM_WEBHOOK_URL and LIFEOS_TELEGRAM_WEBHOOK_SECRET must be set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tg.RegisterWebhook(ctx, client, url, secret); err != nil {
		return err
	}
	info, err := client.GetWebhookInfo(ctx)
	if err != nil {
		return fmt.Errorf("webhook registered but status check failed: %w", err)
	}
	fmt.Fprintf(os.Stdout, "webhook set: %s (pending=%d)\n", info.URL, info.PendingUpdateCount)
	return nil
}

func runTelegramWebhookStatus() error {
	client, _, err := newTelegramClientFromConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	info, err := client.GetWebhookInfo(ctx)
	if err != nil {
		return err
	}
	if info.URL == "" {
		fmt.Fprintln(os.Stdout, "webhook: none (polling mode)")
		return nil
	}
	fmt.Fprintf(os.Stdout, "webhook url: %s\npending updates: %d\nlast error: %q\n", info.URL, info.PendingUpdateCount, info.LastError)
	return nil
}

func runTelegramDeleteWebhook() error {
	client, _, err := newTelegramClientFromConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tg.ClearWebhook(ctx, client); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "webhook deleted; polling will receive updates again")
	return nil
}
