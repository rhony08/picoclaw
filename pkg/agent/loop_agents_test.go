package agent

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/stretchr/testify/assert"
)

// MockProvider implements providers.LLMProvider for testing
type MockProvider struct {
	name string
}

func (m *MockProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	return &providers.LLMResponse{
		Content: "mock response",
	}, nil
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) GetDefaultModel() string {
	return "mock"
}

func TestAgentLoop_getAgentProfile(t *testing.T) {
	// Create temp directory for workspace
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.Agents.Profiles = map[string]config.AgentProfile{
		"coder": {
			Model:       "claude-opus",
			Temperature: 0.2,
		},
	}
	cfg.Agents.Routing = []config.RoutingRule{
		{
			Channel: "telegram",
			UserID:  "123456",
			Agent:   "coder",
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &MockProvider{name: "mock"}
	al := NewAgentLoop(cfg, msgBus, provider)

	tests := []struct {
		name       string
		sessionKey string
		channel    string
		senderID   string
		wantModel  string
		wantTemp   float64
	}{
		{
			name:       "routing rule takes priority",
			sessionKey: "test:1",
			channel:    "telegram",
			senderID:   "123456",
			wantModel:  "claude-opus",
			wantTemp:   0.2,
		},
		{
			name:       "session preference when no routing",
			sessionKey: "test:2",
			channel:    "discord",
			senderID:   "123456",
			wantModel:  "claude-opus",
			wantTemp:   0.2,
		},
		{
			name:       "default when nothing set",
			sessionKey: "test:3",
			channel:    "",
			senderID:   "",
			wantModel:  "glm-4.7",
			wantTemp:   0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set session agent if needed
			if tt.sessionKey == "test:2" {
				al.sessions.SetCurrentAgent(tt.sessionKey, "coder")
			}

			profile := al.getAgentProfile(tt.sessionKey, tt.channel, tt.senderID)
			assert.Equal(t, tt.wantModel, profile.Model)
			assert.Equal(t, tt.wantTemp, profile.Temperature)
		})
	}
}

func TestAgentLoop_SwitchAgent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.Agents.Profiles = map[string]config.AgentProfile{
		"coder": {Model: "claude-opus"},
	}

	msgBus := bus.NewMessageBus()
	provider := &MockProvider{name: "mock"}
	al := NewAgentLoop(cfg, msgBus, provider)

	sessionKey := "test:session"

	// Initially no agent set
	agent := al.sessions.GetCurrentAgent(sessionKey)
	assert.Empty(t, agent)

	// Switch to existing agent
	err := al.SwitchAgent(sessionKey, "coder")
	assert.NoError(t, err)

	agent = al.sessions.GetCurrentAgent(sessionKey)
	assert.Equal(t, "coder", agent)

	// Try to switch to non-existent agent
	err = al.SwitchAgent(sessionKey, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestAgentLoop_ListAgents(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.Agents.Profiles = map[string]config.AgentProfile{
		"coder":      {Model: "claude-opus"},
		"researcher": {Model: "gpt-4"},
	}

	msgBus := bus.NewMessageBus()
	provider := &MockProvider{name: "mock"}
	al := NewAgentLoop(cfg, msgBus, provider)

	agents := al.ListAgents()

	assert.Len(t, agents, 3)
	assert.Contains(t, agents, "default")
	assert.Contains(t, agents, "coder")
	assert.Contains(t, agents, "researcher")
}

func TestAgentLoop_GetCurrentAgentInfo(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.Agents.Defaults.Model = "default-model"
	cfg.Agents.Defaults.Temperature = 0.5
	cfg.Agents.Profiles = map[string]config.AgentProfile{
		"coder": {
			Model:       "claude-opus",
			Temperature: 0.2,
			MaxTokens:   8192,
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &MockProvider{name: "mock"}
	al := NewAgentLoop(cfg, msgBus, provider)

	sessionKey := "test:session"

	// Set agent
	al.sessions.SetCurrentAgent(sessionKey, "coder")

	// Get info
	info := al.GetCurrentAgentInfo(sessionKey, "", "")

	assert.Equal(t, "coder", info["name"])
	assert.Equal(t, "claude-opus", info["model"])
	assert.Equal(t, 0.2, info["temperature"])
	assert.Equal(t, 8192, info["max_tokens"])
}

func TestAgentLoop_handleAgentCommand(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.Agents.Profiles = map[string]config.AgentProfile{
		"coder": {Model: "claude-opus"},
	}

	msgBus := bus.NewMessageBus()
	provider := &MockProvider{name: "mock"}
	al := NewAgentLoop(cfg, msgBus, provider)

	tests := []struct {
		name        string
		content     string
		wantHandled bool
		wantContain string
	}{
		{
			name:        "not an agent command",
			content:     "Hello world",
			wantHandled: false,
		},
		{
			name:        "agent list command",
			content:     "/agent list",
			wantHandled: true,
			wantContain: "Available agents",
		},
		{
			name:        "agent switch command",
			content:     "/agent switch coder",
			wantHandled: true,
			wantContain: "Switched to 'coder'",
		},
		{
			name:        "agent info command",
			content:     "/agent info",
			wantHandled: true,
			wantContain: "Current Agent",
		},
		{
			name:        "agent help command",
			content:     "/agent help",
			wantHandled: true,
			wantContain: "Agent Commands",
		},
		{
			name:        "unknown agent command",
			content:     "/agent unknown",
			wantHandled: true,
			wantContain: "Unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := bus.InboundMessage{
				Content:    tt.content,
				SessionKey: "test:session",
			}
			response, handled := al.handleAgentCommand(nil, msg)
			assert.Equal(t, tt.wantHandled, handled)
			if tt.wantHandled {
				assert.Contains(t, response, tt.wantContain)
			}
		})
	}
}
