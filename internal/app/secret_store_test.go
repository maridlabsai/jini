package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSecretStore is an in-memory store so tests never touch the real keychain.
type fakeSecretStore struct {
	data map[string]string
	up   bool
}

func (f *fakeSecretStore) get(name string) (string, bool) { v, ok := f.data[name]; return v, ok }
func (f *fakeSecretStore) set(name, value string) error   { f.data[name] = value; return nil }
func (f *fakeSecretStore) delete(name string) error       { delete(f.data, name); return nil }
func (f *fakeSecretStore) available() bool                { return f.up }

func withFakeSecretStore(t *testing.T, up bool) *fakeSecretStore {
	t.Helper()
	prev := providerSecretStore
	fake := &fakeSecretStore{data: map[string]string{}, up: up}
	providerSecretStore = fake
	t.Cleanup(func() { providerSecretStore = prev })
	return fake
}

func TestIsSecretConfigKey(t *testing.T) {
	secret := []string{"ANTHROPIC_API_KEY", "XAI_API_KEY", "AWS_SECRET_ACCESS_KEY", "SOME_TOKEN", "DB_PASSWORD"}
	notSecret := []string{"OPENAI_BASE_URL", "XAI_MODEL", "AZURE_OPENAI_ENDPOINT", "JINI_PROVIDER", ""}
	for _, k := range secret {
		if !isSecretConfigKey(k) {
			t.Errorf("%q should be a secret key", k)
		}
	}
	for _, k := range notSecret {
		if isSecretConfigKey(k) {
			t.Errorf("%q should NOT be a secret key", k)
		}
	}
}

// A saved secret must go to the keychain and NOT into the plaintext dotfile;
// a non-secret setting must go to the dotfile as before.
func TestSaveProviderSettingsKeepsSecretsOutOfDotfile(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	// Ensure env doesn't shadow configValue lookups.
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("XAI_MODEL", "")
	fake := withFakeSecretStore(t, true)

	if err := saveProviderSettings(map[string]string{
		"XAI_API_KEY": "sk-secret-123",
		"XAI_MODEL":   "grok-2-latest",
	}); err != nil {
		t.Fatal(err)
	}

	if got := fake.data["XAI_API_KEY"]; got != "sk-secret-123" {
		t.Fatalf("secret not stored in keychain: %q", got)
	}
	// The dotfile must contain the non-secret but never the secret value.
	raw, err := os.ReadFile(filepath.Join(sessionStateRoot(), "provider.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "sk-secret-123") {
		t.Fatalf("secret leaked into plaintext dotfile:\n%s", raw)
	}
	if !strings.Contains(string(raw), "grok-2-latest") {
		t.Fatalf("non-secret setting missing from dotfile:\n%s", raw)
	}
	// configValue resolves the secret from the keychain and the setting from the dotfile.
	if v := configValue("XAI_API_KEY"); v != "sk-secret-123" {
		t.Fatalf("configValue(secret) = %q, want from keychain", v)
	}
	if v := configValue("XAI_MODEL"); v != "grok-2-latest" {
		t.Fatalf("configValue(setting) = %q", v)
	}
}

// Deleting a secret (empty value) removes it from the keychain.
func TestSaveProviderSettingsDeletesSecret(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("XAI_API_KEY", "")
	fake := withFakeSecretStore(t, true)
	fake.data["XAI_API_KEY"] = "old"
	if err := saveProviderSettings(map[string]string{"XAI_API_KEY": ""}); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.data["XAI_API_KEY"]; ok {
		t.Fatal("secret should have been deleted from the keychain")
	}
}

// With no keychain available, secrets fall back to the dotfile (honest degrade).
func TestSecretsFallBackToDotfileWhenNoKeychain(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("XAI_API_KEY", "")
	withFakeSecretStore(t, false) // unavailable
	if err := saveProviderSettings(map[string]string{"XAI_API_KEY": "sk-fallback"}); err != nil {
		t.Fatal(err)
	}
	if v := configValue("XAI_API_KEY"); v != "sk-fallback" {
		t.Fatalf("configValue = %q, want dotfile fallback", v)
	}
}
