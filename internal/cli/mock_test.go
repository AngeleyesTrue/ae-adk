package cli

import (
	"context"

	"github.com/AngeleyesTrue/ae-adk/internal/update"
)

// --- Mock implementations for CLI dependency testing ---

// mockUpdateChecker implements update.Checker for testing.
type mockUpdateChecker struct {
	checkLatestFunc   func(ctx context.Context) (*update.VersionInfo, error)
	isUpdateAvailFunc func(current string) (bool, *update.VersionInfo, error)
}

func (m *mockUpdateChecker) CheckLatest(ctx context.Context) (*update.VersionInfo, error) {
	if m.checkLatestFunc != nil {
		return m.checkLatestFunc(ctx)
	}
	return &update.VersionInfo{Version: "1.0.0", URL: "https://example.com/ae-binary"}, nil
}

func (m *mockUpdateChecker) IsUpdateAvailable(current string) (bool, *update.VersionInfo, error) {
	if m.isUpdateAvailFunc != nil {
		return m.isUpdateAvailFunc(current)
	}
	return false, nil, nil
}

// mockUpdateOrchestrator implements update.Orchestrator for testing.
type mockUpdateOrchestrator struct {
	updateFunc func(ctx context.Context) (*update.UpdateResult, error)
}

func (m *mockUpdateOrchestrator) Update(ctx context.Context) (*update.UpdateResult, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx)
	}
	return &update.UpdateResult{PreviousVersion: "v0.0.0", NewVersion: "v0.0.1"}, nil
}

