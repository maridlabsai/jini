package app

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Secret storage — the PRD requires BYO credentials live in the OS keychain,
// "never plaintext dotfiles" (number-one-platform-prd.md §P0). This is the seam:
// secret-looking config keys (API keys/tokens) go to the OS keychain when one is
// available; everything else (endpoints, model ids) stays in provider.json. On a
// platform without a supported keychain the store reports unavailable and the
// caller keeps the 0600 dotfile fallback, so behavior degrades honestly rather
// than losing the key.

var errSecretStoreUnavailable = errors.New("no OS secret store available")

type secretStore interface {
	get(name string) (string, bool)
	set(name, value string) error
	delete(name string) error
	available() bool
}

// providerSecretStore is overridable in tests so the real keychain is never
// touched by the suite (which would prompt or mutate the developer's keychain).
var providerSecretStore secretStore = defaultSecretStore()

func defaultSecretStore() secretStore {
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("security"); err == nil {
			return &macKeychainStore{account: "jini", cache: map[string]cachedSecret{}}
		}
	}
	return unavailableSecretStore{}
}

// isSecretConfigKey reports whether a config key holds a credential that must
// not sit in a plaintext dotfile.
func isSecretConfigKey(name string) bool {
	up := strings.ToUpper(strings.TrimSpace(name))
	if up == "" {
		return false
	}
	for _, marker := range []string{"API_KEY", "SECRET", "TOKEN", "PASSWORD"} {
		if strings.Contains(up, marker) {
			return true
		}
	}
	return false
}

// macKeychainStore backs secrets with the macOS login keychain via the
// `security` CLI. Each secret is a generic password keyed by (account="jini",
// service=<config key>). A process-lifetime cache (including negative results)
// keeps `configValue` — a hot path — from spawning `security` on every call for
// an unset secret.
type cachedSecret struct {
	value string
	found bool
}

type macKeychainStore struct {
	account string
	mu      sync.Mutex
	cache   map[string]cachedSecret
}

func (*macKeychainStore) available() bool { return true }

func (m *macKeychainStore) get(name string) (string, bool) {
	m.mu.Lock()
	if c, ok := m.cache[name]; ok {
		m.mu.Unlock()
		return c.value, c.found
	}
	m.mu.Unlock()

	out, err := exec.Command("security", "find-generic-password", "-a", m.account, "-s", name, "-w").Output()
	result := cachedSecret{}
	if err == nil {
		if value := strings.TrimRight(string(out), "\r\n"); value != "" {
			result = cachedSecret{value: value, found: true}
		}
	}
	m.mu.Lock()
	m.cache[name] = result
	m.mu.Unlock()
	return result.value, result.found
}

func (m *macKeychainStore) set(name, value string) error {
	// -U updates the item in place when it already exists.
	if err := exec.Command("security", "add-generic-password", "-a", m.account, "-s", name, "-w", value, "-U").Run(); err != nil {
		return err
	}
	m.mu.Lock()
	m.cache[name] = cachedSecret{value: value, found: true}
	m.mu.Unlock()
	return nil
}

func (m *macKeychainStore) delete(name string) error {
	// A missing item is not an error for our purposes.
	_ = exec.Command("security", "delete-generic-password", "-a", m.account, "-s", name).Run()
	m.mu.Lock()
	m.cache[name] = cachedSecret{}
	m.mu.Unlock()
	return nil
}

type unavailableSecretStore struct{}

func (unavailableSecretStore) get(string) (string, bool) { return "", false }
func (unavailableSecretStore) set(string, string) error  { return errSecretStoreUnavailable }
func (unavailableSecretStore) delete(string) error       { return nil }
func (unavailableSecretStore) available() bool           { return false }
