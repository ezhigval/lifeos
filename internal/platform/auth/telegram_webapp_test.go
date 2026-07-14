package auth

import (
	"strconv"
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
