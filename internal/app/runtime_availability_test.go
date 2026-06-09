package app

import "testing"

func withRuntimeConnectivityProbe(t *testing.T, result runtimeConnectivityProbeResult) {
	t.Helper()
	previous := runtimeConnectivityProbe
	runtimeConnectivityProbe = func() runtimeConnectivityProbeResult {
		return result
	}
	t.Cleanup(func() {
		runtimeConnectivityProbe = previous
	})
}
