package main

import (
	"testing"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/migrate"
)

func TestNeedsMigrationDetection(t *testing.T) {
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
			result := migrate.NeedsMigration(tt.cfg)
			if result != tt.expected {
				t.Errorf("NeedsMigration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		expectModels int
	}{
		{
			name: "Successful migration",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model:    "glm-4.7",
						Provider: "z.ai",
					},
				},
			},
			expectError: false,
			expectModels: 1,
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
			expectModels: 1,
		},
		{
			name: "No migration needed",
			cfg: &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model: "glm-4.7",
					},
				},
			},
			expectError: false,
			expectModels: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := migrate.MigrateToNewFormat(tt.cfg)

			if tt.expectError && err == nil {
				t.Errorf("MigrateToNewFormat() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("MigrateToNewFormat() unexpected error: %v", err)
			}

			if len(tt.cfg.Agents.Defaults.Models) != tt.expectModels {
				t.Errorf("Models count = %d, want %d", len(tt.cfg.Agents.Defaults.Models), tt.expectModels)
			}
		})
	}
}

func TestCreateProviderWithFallback(t *testing.T) {
	// This test would normally require actual provider creation
	// For now, we'll test the logic without creating actual providers
	t.Run("Multiple model candidates", func(t *testing.T) {
		// Mock config with multiple model candidates
		cfg := &config.Config{
			Agents: config.AgentsConfig{
				Defaults: config.AgentDefaults{
					Models: []config.ModelSpec{
						{Model: "glm-4.7", Provider: "z.ai"},
						{Model: "claude-3", Provider: "anthropic"},
					},
				},
			},
			Providers: config.ProvidersConfig{
				Zhipu:     config.ProviderConfig{APIKey: "test-key"},
				Anthropic: config.ProviderConfig{APIKey: "test-key"},
			},
		}
		// Must call PrepareAgentModels to populate ResolvedModels
		cfg.PrepareAgentModels()

		candidates := cfg.Agents.Defaults.ModelCandidates()
		if len(candidates) != 2 {
			t.Errorf("Expected 2 candidates, got %d", len(candidates))
		}

		expected := []string{"glm-4.7", "claude-3"}
		for i, candidate := range candidates {
			if candidate != expected[i] {
				t.Errorf("Candidate %d = %s, want %s", i, candidate, expected[i])
			}
		}
	})

	t.Run("Single model candidate (old format)", func(t *testing.T) {
		singleCfg := &config.Config{
			Agents: config.AgentsConfig{
				Defaults: config.AgentDefaults{
					Model:    "glm-4.7",
					Provider: "z.ai",
				},
			},
			Providers: config.ProvidersConfig{
				Zhipu: config.ProviderConfig{APIKey: "test-key"},
			},
		}

		candidates := singleCfg.Agents.Defaults.ModelCandidates()
		if len(candidates) != 1 {
			t.Errorf("Expected 1 candidate, got %d", len(candidates))
		}

		// Old format without prepareModels returns the raw model value
		if candidates[0] != "glm-4.7" {
			t.Errorf("Candidate = %s, want glm-4.7", candidates[0])
		}
	})

	t.Run("No model configured", func(t *testing.T) {
		emptyCfg := &config.Config{
			Agents: config.AgentsConfig{
				Defaults: config.AgentDefaults{},
			},
		}

		candidates := emptyCfg.Agents.Defaults.ModelCandidates()
		if len(candidates) != 0 {
			t.Errorf("Expected 0 candidates, got %d", len(candidates))
		}
	})
}

func TestModelDisplayLogic(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *config.Config
		expectedOutput string
	}{
		{
			name: "Single model",
			setup: func() *config.Config {
				return &config.Config{
					Agents: config.AgentsConfig{
						Defaults: config.AgentDefaults{
							Model: "glm-4.7",
						},
					},
				}
			},
			expectedOutput: "Model: glm-4.7\n",
		},
		{
			name: "Multiple models",
			setup: func() *config.Config {
				cfg := &config.Config{
					Agents: config.AgentsConfig{
						Defaults: config.AgentDefaults{
							Models: []config.ModelSpec{
								{Model: "glm-4.7", Provider: "z.ai"},
								{Model: "claude-3", Provider: "anthropic"},
							},
						},
					},
				}
				// Must call PrepareAgentModels to populate ResolvedModels
				cfg.PrepareAgentModels()
				return cfg
			},
			expectedOutput: "Models: glm-4.7, claude-3 (first: glm-4.7)\n",
		},
		{
			name: "No model",
			setup: func() *config.Config {
				return &config.Config{
					Agents: config.AgentsConfig{
						Defaults: config.AgentDefaults{},
					},
				}
			},
			expectedOutput: "Model: \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup()
			// Capture output (this is a simplified test)
			candidates := cfg.Agents.Defaults.ModelCandidates()
			var output string

			if len(candidates) > 0 {
				if len(candidates) == 1 {
					output = "Model: " + candidates[0] + "\n"
				} else {
					output = "Models: " + joinCandidates(candidates) + "\n"
				}
			} else {
				output = "Model: " + cfg.Agents.Defaults.Model + "\n"
			}

			if output != tt.expectedOutput {
				t.Errorf("Output = %q, want %q", output, tt.expectedOutput)
			}
		})
	}
}

// Helper function to join candidates for testing
func joinCandidates(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	
	result := candidates[0]
	for i := 1; i < len(candidates); i++ {
		result += ", " + candidates[i]
	}
	result += " (first: " + candidates[0] + ")"
	return result
}