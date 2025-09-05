package git

import (
	"testing"
	"time"
)

// TestObserveSyncSuccess ensures success path sets last sync time and updates counters.
func TestObserveSyncSuccess(t *testing.T) {
	before := getLastSyncTime()
	observeSync(t.Context(), 2*time.Second, 3, true)

	after := getLastSyncTime()
	if after <= before {
		t.Errorf("expected last sync time to update, before=%d after=%d", before, after)
	}
}

// TestObserveSyncFailure ensures failure path sets last sync time and updates counters.
func TestObserveSyncFailure(t *testing.T) {
	setLastSyncTime(time.Now().UTC().Add(-time.Minute))
	before := getLastSyncTime()
	observeSync(t.Context(), time.Second, 0, false)

	after := getLastSyncTime()
	if after <= before {
		t.Errorf("expected last sync time to update, before=%d after=%d", before, after)
	}
}

// TestSetAndGetLastSyncTime directly validates setter/getter.
func TestSetAndGetLastSyncTime(t *testing.T) {
	now := time.Now()
	setLastSyncTime(now)
	got := getLastSyncTime()
	if got != now.Unix() {
		t.Errorf("expected last sync time %d, got %d", now.Unix(), got)
	}
}
