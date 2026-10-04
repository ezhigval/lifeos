package auth

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestValidateWebAppInitDataOK(t *testing.T) {
	t.Parallel()
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	now := time.Unix(1_700_000_000, 0).UTC()
	userJSON := `{"id":42,"first_name":"Ada","last_name":"Lovelace","username":"ada","language_code":"en"}`
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"query_id":  "AAEAAAE",
		"user":      userJSON,
	}, token)

	got, err := ValidateWebAppInitData(initData, token, time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.User.ID != 42 || got.User.Username != "ada" {
		t.Fatalf("user=%+v", got.User)
	}
	if got.User.DisplayName() != "Ada Lovelace" {
		t.Fatalf("display=%q", got.User.DisplayName())
	}
	if got.QueryID != "AAEAAAE" {
		t.Fatalf("query_id=%q", got.QueryID)
	}
}

func TestValidateWebAppInitDataRejectsTamperedHash(t *testing.T) {
	t.Parallel()
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	now := time.Unix(1_700_000_000, 0).UTC()
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      `{"id":1,"first_name":"A"}`,
	}, token)
	initData += "x" // corrupt query string
	if _, err := ValidateWebAppInitData(initData, token, time.Hour, now); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateWebAppInitDataRejectsExpired(t *testing.T) {
	t.Parallel()
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	auth := time.Unix(1_700_000_000, 0).UTC()
	now := auth.Add(48 * time.Hour)
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(auth.Unix(), 10),
		"user":      `{"id":1,"first_name":"A"}`,
	}, token)
	if _, err := ValidateWebAppInitData(initData, token, 24*time.Hour, now); err == nil {
		t.Fatal("expected expired")
	}
}

func TestValidateWebAppInitDataRejectsWrongToken(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_700_000_000, 0).UTC()
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      `{"id":1,"first_name":"A"}`,
	}, "token-a")
	if _, err := ValidateWebAppInitData(initData, "token-b", time.Hour, now); err == nil {
		t.Fatal("expected hash mismatch")
	}
}

func TestValidateWebAppInitDataParsesTelegramUserID(t *testing.T) {
	t.Parallel()
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	now := time.Unix(1_700_000_000, 0).UTC()

	cases := []struct {
		name    string
		user    string
		wantID  int64
		wantErr string
	}{
		{
			name:   "realistic telegram id",
			user:   `{"id":279058397,"first_name":"Val","username":"val","language_code":"ru","is_premium":true,"photo_url":"https://x"}`,
			wantID: 279058397,
		},
		{
			name:   "id as json number only required",
			user:   `{"id":7}`,
			wantID: 7,
		},
		{
			name:    "missing user field",
			user:    "",
			wantErr: "missing user",
		},
		{
			name:    "zero id",
			user:    `{"id":0,"first_name":"X"}`,
			wantErr: "invalid user id",
		},
		{
			name:    "negative id",
			user:    `{"id":-1,"first_name":"X"}`,
			wantErr: "invalid user id",
		},
		{
			name:    "broken json",
			user:    `{id:1}`,
			wantErr: "invalid user json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fields := map[string]string{
				"auth_date": strconv.FormatInt(now.Unix(), 10),
			}
			if tc.user != "" {
				fields["user"] = tc.user
			}
			initData := SignWebAppInitData(fields, token)
			got, err := ValidateWebAppInitData(initData, token, time.Hour, now)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err=%v want contains %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.User.ID != tc.wantID {
				t.Fatalf("id=%d want %d", got.User.ID, tc.wantID)
			}
		})
	}
}

func TestValidateWebAppInitDataAcceptsURLEncodedUser(t *testing.T) {
	t.Parallel()
	// SignWebAppInitData uses url.Values.Encode which percent-encodes the user JSON —
	// same shape Telegram sends over the wire.
	const token = "123456:TEST"
	now := time.Unix(1_700_000_000, 0).UTC()
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      `{"id":900001,"first_name":"Test User"}`,
	}, token)
	if !strings.Contains(initData, "user=") {
		t.Fatalf("expected encoded initData, got %q", initData)
	}
	got, err := ValidateWebAppInitData(initData, token, time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.User.ID != 900001 || got.User.FirstName != "Test User" {
		t.Fatalf("user=%+v", got.User)
	}
}

func TestValidateWebAppInitDataIncludesSignatureInCheckString(t *testing.T) {
	t.Parallel()
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	now := time.Unix(1_700_000_000, 0).UTC()
	// Signature is present on modern Telegram clients; hash HMAC covers it.
	initData := SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      `{"id":99,"first_name":"Sig"}`,
		"signature": "third-party-ed25519-placeholder",
	}, token)
	if !strings.Contains(initData, "signature=") {
		t.Fatal("expected signature in initData")
	}
	got, err := ValidateWebAppInitData(initData, token, time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.User.ID != 99 {
		t.Fatalf("id=%d", got.User.ID)
	}
}
