package app

import "testing"

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
