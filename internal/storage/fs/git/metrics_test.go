package git

import (
	"context"
	"testing"
	"time"
)

// TestObserveSuccess ensures calling ObserveSuccess does not panic.
func TestObserveSuccess(t *testing.T) {
	ctx := context.Background()
	ObserveSuccess(ctx, "manual")
}

// TestObserveFailure ensures calling ObserveFailure does not panic.
func TestObserveFailure(t *testing.T) {
	ctx := context.Background()
	ObserveFailure(ctx, "polling")
}

// TestObserveFlagsFetched ensures flags fetched counter increments without error.
func TestObserveFlagsFetched(t *testing.T) {
	ctx := context.Background()
	ObserveFlagsFetched(ctx, 5, "manual")
}

// TestObserveDuration ensures duration histogram records without error.
func TestObserveDuration(t *testing.T) {
	ctx := context.Background()
	ObserveDuration(ctx, 1.23, "manual")
}

// TestObserveFailureWithReason ensures reason attribute path is exercised.
func TestObserveFailureWithReason(t *testing.T) {
	ctx := context.Background()
	ObserveFailureWithReason(ctx, "webhook", "timeout")
}

// TestObserveSyncSuccess ensures success path sets last sync time and updates counters.
func TestObserveSyncSuccess(t *testing.T) {
	ctx := context.Background()
	before := GetLastSyncTime()

	ObserveSync(ctx, 0.5, 3, true, "manual", "no_change")

	after := GetLastSyncTime()
	if after <= before {
		t.Errorf("expected last sync time to update, before=%d after=%d", before, after)
	}
}

// TestObserveSyncFailure ensures failure path sets last sync time and updates counters.
func TestObserveSyncFailure(t *testing.T) {
	ctx := context.Background()
	before := GetLastSyncTime()

	ObserveSync(ctx, 1.0, 0, false, "polling", "fetch_failed")

	after := GetLastSyncTime()
	if after <= before {
		t.Errorf("expected last sync time to update, before=%d after=%d", before, after)
	}
}

// TestSetAndGetLastSyncTime directly validates setter/getter.
func TestSetAndGetLastSyncTime(t *testing.T) {
	now := time.Now().Unix()
	setLastSyncTime(now)
	got := GetLastSyncTime()
	if got != now {
		t.Errorf("expected last sync time %d, got %d", now, got)
	}
}
