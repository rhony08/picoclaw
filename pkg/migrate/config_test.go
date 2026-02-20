package migrate

import (
	"testing"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestIsNewFormat(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected bool
	}{
		{
			name: "New format with models array",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Models: []config.ModelSpec{
							{Model: "glm-4.7", Provider: "z.ai"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "New format with multiple models",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Models: []config.ModelSpec{
							{Model: "glm-4.7", Provider: "z.ai"},
							{Model: "claude-3", Provider: "anthropic"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Old format with model only",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model: "glm-4.7",
					},
				},
			},
			expected: false,
		},
		{
			name: "Old format with model and provider",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model:    "glm-4.7",
						Provider: "z.ai",
					},
				},
			},
			expected: false,
		},
		{
			name: "Empty config",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewFormat(tt.cfg)
			if result != tt.expected {
				t.Errorf("IsNewFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNeedsMigration(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected bool
	}{
		{
			name: "New format - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Models: []config.ModelSpec{
							{Model: "glm-4.7", Provider: "z.ai"},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Old format with both model and provider - needs migration",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model:    "glm-4.7",
						Provider: "z.ai",
					},
				},
			},
			expected: true,
		},
		{
			name: "Old format with model only - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model: "glm-4.7",
					},
				},
			},
			expected: false,
		},
		{
			name: "Old format with provider only - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Provider: "z.ai",
					},
				},
			},
			expected: false,
		},
		{
			name: "Empty config - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsMigration(tt.cfg)
			if result != tt.expected {
				t.Errorf("NeedsMigration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMigrateToNewFormat(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		expectModel string
		expectProv  string
		modelsCount int
	}{
		{
			name: "Successful migration - old format to new",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model:    "glm-4.7",
						Provider: "z.ai",
					},
				},
			},
			expectError: false,
			expectModel: "",
			expectProv:  "",
			modelsCount: 1,
		},
		{
			name: "Already new format - no migration",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Models: []config.ModelSpec{
							{Model: "glm-4.7", Provider: "z.ai"},
						},
					},
				},
			},
			expectError: false,
			expectModel: "",
			expectProv:  "",
			modelsCount: 1,
		},
		{
			name: "Old format with model only - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model: "glm-4.7",
					},
				},
			},
			expectError: false,
			expectModel: "",
			expectProv:  "",
			modelsCount: 0,
		},
		{
			name: "Empty config - no migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{},
				},
			},
			expectError: false,
			expectModel: "",
			expectProv:  "",
			modelsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MigrateToNewFormat(tt.cfg)
			
			if tt.expectError && err == nil {
				t.Errorf("MigrateToNewFormat() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("MigrateToNewFormat() unexpected error: %v", err)
			}
			
			// Check that models array is correctly populated
			if len(tt.cfg.Agents.Defaults.Models) != tt.modelsCount {
				t.Errorf("Models count = %d, want %d", len(tt.cfg.Agents.Defaults.Models), tt.modelsCount)
			}
			
			// Check that old fields are cleared for migrated configs
			if tt.modelsCount > 0 {
				if tt.cfg.Agents.Defaults.Model != "" {
					t.Errorf("Old model field should be empty, got: %s", tt.cfg.Agents.Defaults.Model)
				}
				if tt.cfg.Agents.Defaults.Provider != "" {
					t.Errorf("Old provider field should be empty, got: %s", tt.cfg.Agents.Defaults.Provider)
				}
			}
			
			// Check migrated model spec
			if tt.modelsCount == 1 {
				migratedModel := tt.cfg.Agents.Defaults.Models[0]
				if migratedModel.Model != "glm-4.7" {
					t.Errorf("Migrated model = %s, want glm-4.7", migratedModel.Model)
				}
				if migratedModel.Provider != "z.ai" {
					t.Errorf("Migrated provider = %s, want z.ai", migratedModel.Provider)
				}
			}
		})
	}
}

func TestConvertModelSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected config.ModelSpec
		expectError bool
	}{
		{
			name: "Complete model spec",
			input: map[string]interface{}{
				"model":    "glm-4.7",
				"provider": "z.ai",
			},
			expected: config.ModelSpec{
				Model:    "glm-4.7",
				Provider: "z.ai",
			},
			expectError: false,
		},
		{
			name: "Model only",
			input: map[string]interface{}{
				"model": "glm-4.7",
			},
			expected: config.ModelSpec{
				Model:    "glm-4.7",
				Provider: "",
			},
			expectError: false,
		},
		{
			name: "Provider only",
			input: map[string]interface{}{
				"provider": "z.ai",
			},
			expected: config.ModelSpec{
				Model:    "",
				Provider: "z.ai",
			},
			expectError: false,
		},
		{
			name: "Empty map",
			input: map[string]interface{}{},
			expected: config.ModelSpec{
				Model:    "",
				Provider: "",
			},
			expectError: false,
		},
		{
			name: "Invalid types",
			input: map[string]interface{}{
				"model":    123,
				"provider": true,
			},
			expected: config.ModelSpec{
				Model:    "",
				Provider: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertModelSpec(tt.input)
			
			if tt.expectError && err == nil {
				t.Errorf("convertModelSpec() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("convertModelSpec() unexpected error: %v", err)
			}
			
			if result.Model != tt.expected.Model {
				t.Errorf("Model = %s, want %s", result.Model, tt.expected.Model)
			}
			if result.Provider != tt.expected.Provider {
				t.Errorf("Provider = %s, want %s", result.Provider, tt.expected.Provider)
			}
		})
	}
}