// Live check: sign initData with bot token, POST auth, print telegram_id from response.
// Invoked by scripts/verify-webapp-auth.sh when env is set.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/auth"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	base := os.Getenv("BASE_URL")
	if token == "" || base == "" {
		fmt.Fprintln(os.Stderr, "TELEGRAM_BOT_TOKEN and BASE_URL required")
		os.Exit(1)
	}
	const wantID int64 = 424242
	now := time.Now().UTC()
	initData := auth.SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      fmt.Sprintf(`{"id":%d,"first_name":"Verify","username":"verify_bot"}`, wantID),
	}, token)

	body, _ := json.Marshal(map[string]string{"init_data": initData})
	resp, err := http.Post(base+"/api/v1/auth/telegram-webapp", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "status=%d body=%s\n", resp.StatusCode, raw)
		os.Exit(1)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		TelegramID  int64  `json:"telegram_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		os.Exit(1)
	}
	if out.AccessToken == "" || out.TelegramID != wantID {
		fmt.Fprintf(os.Stderr, "unexpected response: %+v\n", out)
		os.Exit(1)
	}
	fmt.Printf("OK live auth: telegram_id=%d (from signed initData.user.id) → JWT issued\n", out.TelegramID)
}
