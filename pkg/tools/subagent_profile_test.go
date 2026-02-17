package tools

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/stretchr/testify/assert"
)

func TestSubagentProfile_Defaults(t *testing.T) {
	profile := &SubagentProfile{}

	// All fields should be zero values initially
	assert.Empty(t, profile.Model)
	assert.Equal(t, 0.0, profile.Temperature)
	assert.Equal(t, 0, profile.MaxTokens)
	assert.Equal(t, 0, profile.MaxIterations)
	assert.Empty(t, profile.SystemPrompt)
	assert.Empty(t, profile.Workspace)
	assert.False(t, profile.RestrictToWorkspace)
}

func TestSubagentProfile_CustomValues(t *testing.T) {
	profile := &SubagentProfile{
		Model:               "claude-opus-4",
		Temperature:         0.2,
		MaxTokens:           8192,
		MaxIterations:       30,
		SystemPrompt:        "You are a coding expert",
		Workspace:           "/custom/workspace",
		RestrictToWorkspace: true,
	}

	assert.Equal(t, "claude-opus-4", profile.Model)
	assert.Equal(t, 0.2, profile.Temperature)
	assert.Equal(t, 8192, profile.MaxTokens)
	assert.Equal(t, 30, profile.MaxIterations)
	assert.Equal(t, "You are a coding expert", profile.SystemPrompt)
	assert.Equal(t, "/custom/workspace", profile.Workspace)
	assert.True(t, profile.RestrictToWorkspace)
}

func TestSubagentManager_SpawnWithProfile(t *testing.T) {
	// This is a basic test - full integration would need a mock provider
	// For now, we just verify the method signature works
	provider := &mockProvider{}
	manager := NewSubagentManager(provider, "default-model", "/workspace", nil)

	// Use context to avoid nil dereference
	ctx := context.Background()
	_, err := manager.Spawn(ctx, "test task", "test-label", "cli", "direct", nil)
	assert.NoError(t, err)
}

// mockProvider implements a minimal provider for testing
type mockProvider struct{}

func (m *mockProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	return &providers.LLMResponse{Content: "mock"}, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) GetDefaultModel() string {
	return "mock"
}
