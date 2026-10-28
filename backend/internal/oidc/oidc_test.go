package oidc

import (
	"crypto/sha256"
	"strings"
	"testing"
)

func TestVerifyPKCE(t *testing.T) {
	verifier := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(verifier))
	challenge := b64(sum[:])
	if !VerifyPKCE(verifier, challenge) {
		t.Fatal("верный verifier должен проходить")
	}
	if VerifyPKCE(strings.Repeat("b", 64), challenge) {
		t.Fatal("чужой verifier не должен проходить")
	}
	if VerifyPKCE("short", challenge) {
		t.Fatal("слишком короткий verifier не должен проходить")
	}
}

func TestNormalizeScope(t *testing.T) {
	got, err := normalizeScope("sport openid profile")
	if err != nil || got != "openid profile sport" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := normalizeScope("profile"); err == nil {
		t.Fatal("без openid должна быть ошибка")
	}
	if _, err := normalizeScope("openid admin"); err == nil {
		t.Fatal("неизвестный scope должен отклоняться")
	}
}

func TestValidRedirect(t *testing.T) {
	for uri, want := range map[string]bool{
		"https://partner.ru/cb": true, "http://localhost:3000/cb": true,
		"http://evil.com/cb": false, "https://x.ru/cb#frag": false, "not a url": false,
	} {
		if validRedirect(uri) != want {
			t.Errorf("%s: ожидалось %v", uri, want)
		}
	}
}
