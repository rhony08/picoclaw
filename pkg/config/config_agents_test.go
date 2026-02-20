package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentsConfig_GetAgentProfile(t *testing.T) {
	cfg := DefaultConfig()

	// Add test profiles
	cfg.Agents.Profiles = map[string]AgentProfile{
		"coder": {
			Models: []ModelSpec{
				{Provider: "anthropic", Model: "claude-opus-4"},
			},
			Temperature:       0.2,
			MaxTokens:         8192,
			MaxToolIterations: 30,
			SystemPrompt:      "You are a coder",
		},
		"researcher": {
			Model:       "openai/gpt-4",
			Temperature: 0.8,
		},
	}

	cfg.PrepareAgentModels()

	tests := []struct {
		name          string
		profileName   string
		wantModel     string
		wantTemp      float64
		wantMaxTokens int
	}{
		{
			name:          "default profile",
			profileName:   "default",
			wantModel:     "glm-4.7", // From defaults
			wantTemp:      0.7,
			wantMaxTokens: 8192,
		},
		{
			name:          "empty profile name returns default",
			profileName:   "",
			wantModel:     "glm-4.7",
			wantTemp:      0.7,
			wantMaxTokens: 8192,
		},
		{
			name:          "coder profile",
			profileName:   "coder",
			wantModel:     "claude-opus-4",
			wantTemp:      0.2,
			wantMaxTokens: 8192,
		},
		{
			name:          "researcher profile with partial config",
			profileName:   "researcher",
			wantModel:     "openai/gpt-4",
			wantTemp:      0.8,
			wantMaxTokens: 8192, // From defaults
		},
		{
			name:          "non-existent profile returns default",
			profileName:   "nonexistent",
			wantModel:     "glm-4.7",
			wantTemp:      0.7,
			wantMaxTokens: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := cfg.GetAgentProfile(tt.profileName)
			assert.Equal(t, tt.wantModel, profile.Model)
			assert.Equal(t, tt.wantTemp, profile.Temperature)
			assert.Equal(t, tt.wantMaxTokens, profile.MaxTokens)
		})
	}
}

func TestAgentsConfig_GetRoutedAgent(t *testing.T) {
	cfg := DefaultConfig()

	// Add routing rules
	cfg.Agents.Routing = []RoutingRule{
		{
			Channel: "telegram",
			Agent:   "coder",
		},
		{
			Channel: "discord",
			UserIDs: []string{"789012"},
			Agent:   "researcher",
		},
		{
			Channel: "line",
			UserIDs: []string{"*", "111"},
			Agent:   "creative",
		},
	}

	tests := []struct {
		name      string
		channel   string
		userID    string
		wantAgent string
	}{
		{
			name:      "matching telegram rule",
			channel:   "telegram",
			userID:    "123456",
			wantAgent: "coder",
		},
		{
			name:      "matching discord rule",
			channel:   "discord",
			userID:    "789012",
			wantAgent: "researcher",
		},
		{
			name:      "discord rule but different user",
			channel:   "discord",
			userID:    "000000",
			wantAgent: "",
		},
		{
			name:      "line wildcard rule",
			channel:   "line",
			userID:    "555",
			wantAgent: "creative",
		},
		{
			name:      "line specific user",
			channel:   "line",
			userID:    "111",
			wantAgent: "creative",
		},
		{
			name:      "no match - different channel",
			channel:   "slack",
			userID:    "123456",
			wantAgent: "",
		},
		{
			name:      "empty channel",
			channel:   "",
			userID:    "123456",
			wantAgent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetRoutedAgent(tt.channel, tt.userID)
			assert.Equal(t, tt.wantAgent, got)
		})
	}
}

func TestAgentsConfig_ListAgentProfiles(t *testing.T) {
	cfg := DefaultConfig()

	// Initially should only have default
	profiles := cfg.ListAgentProfiles()
	assert.Contains(t, profiles, "default")
	assert.Len(t, profiles, 1)

	// Add profiles
	cfg.Agents.Profiles = map[string]AgentProfile{
		"coder":      {Model: "claude-opus-4"},
		"researcher": {Model: "gpt-4"},
	}

	profiles = cfg.ListAgentProfiles()
	assert.Len(t, profiles, 3)
	assert.Contains(t, profiles, "default")
	assert.Contains(t, profiles, "coder")
	assert.Contains(t, profiles, "researcher")
}

func TestAgentsConfig_ProfileExists(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Agents.Profiles = map[string]AgentProfile{
		"coder": {Model: "claude-opus-4"},
	}

	tests := []struct {
		name      string
		profile   string
		wantExist bool
	}{
		{"default exists", "default", true},
		{"empty string is default", "", true},
		{"existing profile", "coder", true},
		{"non-existent profile", "researcher", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.ProfileExists(tt.profile)
			assert.Equal(t, tt.wantExist, got)
		})
	}
}

func TestAgentProfile_MergeWithDefaults(t *testing.T) {
	cfg := DefaultConfig()

	// Set non-default defaults
	cfg.Agents.Defaults.Model = "default-model"
	cfg.Agents.Defaults.Temperature = 0.5
	cfg.Agents.Defaults.MaxTokens = 4096
	cfg.Agents.Defaults.MaxToolIterations = 10

	// Create profile that only overrides some fields
	cfg.Agents.Profiles = map[string]AgentProfile{
		"partial": {
			Model:       "custom-model",
			Temperature: 0.9,
			// MaxTokens and MaxToolIterations not set
		},
	}

	profile := cfg.GetAgentProfile("partial")

	// Overridden fields
	assert.Equal(t, "custom-model", profile.Model)
	assert.Equal(t, 0.9, profile.Temperature)

	// Inherited from defaults
	assert.Equal(t, 4096, profile.MaxTokens)
	assert.Equal(t, 10, profile.MaxToolIterations)
}


