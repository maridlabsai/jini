package app

import (
	"errors"
	"net/http"
	"testing"
)

type localDiscoveryRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn localDiscoveryRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func init() {
	localSLMDiscoveryHTTPClient = &http.Client{Transport: localDiscoveryRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("no test local runtime")
	})}
	powerProfileProbeEnabled = false
	runProbeCommand = func(name string, args ...string) (string, error) {
		switch name {
		case "sw_vers":
			return "15.7.7", nil
		case "sysctl":
			return "17179869184", nil
		case "uname":
			return "test-kernel", nil
		default:
			return "", errors.New("test subprocess probe disabled")
		}
	}
}

func resetLocalSLMAutoDiscoveryForTest(t *testing.T) {
	t.Helper()
	localSLMAutoDiscoveryMu.Lock()
	previousCached := localSLMAutoDiscoveryCached
	previousKey := localSLMAutoDiscoveryKey
	previousValue := localSLMAutoDiscoveryValue
	localSLMAutoDiscoveryCached = false
	localSLMAutoDiscoveryKey = ""
	localSLMAutoDiscoveryValue = localSLMRuntimeDiscovery{}
	localSLMAutoDiscoveryMu.Unlock()
	t.Cleanup(func() {
		localSLMAutoDiscoveryMu.Lock()
		localSLMAutoDiscoveryCached = previousCached
		localSLMAutoDiscoveryKey = previousKey
		localSLMAutoDiscoveryValue = previousValue
		localSLMAutoDiscoveryMu.Unlock()
	})
}

func withLocalSLMDiscoveryHTTPClient(t *testing.T, fn localDiscoveryRoundTripFunc) {
	t.Helper()
	previous := localSLMDiscoveryHTTPClient
	localSLMDiscoveryHTTPClient = &http.Client{Transport: fn}
	t.Cleanup(func() { localSLMDiscoveryHTTPClient = previous })
}
