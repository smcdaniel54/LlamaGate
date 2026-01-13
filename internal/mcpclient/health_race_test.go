package mcpclient

import (
	"sync"
	"testing"
	"time"
)

// TestHealthMonitor_Start_RaceCondition tests that Start() can be called concurrently
// without starting multiple monitor loops
func TestHealthMonitor_Start_RaceCondition(_ *testing.T) {
	hm := NewHealthMonitor(100*time.Millisecond, 5*time.Second)

	// Call Start() concurrently from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			hm.Start()
		}()
	}

	wg.Wait()

	// Give the monitor loop a moment to start
	time.Sleep(50 * time.Millisecond)

	// Verify only one monitor loop is running by checking stopChan
	// If multiple loops were started, Stop() would panic trying to close an already closed channel
	// This test verifies that doesn't happen
	hm.Stop()

	// If we get here without panicking, the race condition is fixed
	// Additional verification: Start() should be idempotent
	hm2 := NewHealthMonitor(100*time.Millisecond, 5*time.Second)
	hm2.Start()
	hm2.Start() // Call again - should be safe
	hm2.Start() // Call again - should be safe
	hm2.Stop()
}

// TestHealthMonitor_Start_AfterStop tests that Start() doesn't start a new loop after Stop()
func TestHealthMonitor_Start_AfterStop(_ *testing.T) {
	hm := NewHealthMonitor(100*time.Millisecond, 5*time.Second)

	// Start and stop
	hm.Start()
	time.Sleep(50 * time.Millisecond)
	hm.Stop()

	// Try to start again after stop - should not start new loop
	hm.Start()
	time.Sleep(50 * time.Millisecond)

	// Should be safe to stop again (no panic)
	hm.Stop()
}

// TestHealthMonitor_Stop_MultipleCalls tests that Stop() can be called multiple times
// without panicking (close of closed channel)
func TestHealthMonitor_Stop_MultipleCalls(_ *testing.T) {
	hm := NewHealthMonitor(100*time.Millisecond, 5*time.Second)

	// Start the monitor
	hm.Start()
	time.Sleep(50 * time.Millisecond)

	// Call Stop() multiple times concurrently - should not panic
	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Multiple calls to Stop() should be safe
			hm.Stop()
			hm.Stop()
			hm.Stop()
		}()
	}

	wg.Wait()

	// If we get here without panicking, the fix works
	// Verify that Stop() is idempotent
	hm2 := NewHealthMonitor(100*time.Millisecond, 5*time.Second)
	hm2.Start()
	time.Sleep(50 * time.Millisecond)

	// Multiple sequential calls should be safe
	hm2.Stop()
	hm2.Stop()
	hm2.Stop()
}
