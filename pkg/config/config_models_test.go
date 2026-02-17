package config

import (
	"testing"
)

func TestModelSpecResolvedModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		provider string
		expected string
	}{
		{
			name:     "Both model and provider",
			model:    "glm-4.7",
			provider: "z.ai",
			expected: "z.ai/glm-4.7",
		},
		{
			name:     "Model only",
			model:    "glm-4.7",
			provider: "",
			expected: "glm-4.7",
		},
		{
			name:     "Provider only",
			model:    "",
			provider: "z.ai",
			expected: "",
		},
		{
			name:     "Both empty",
			model:    "",
			provider: "",
			expected: "",
		},
		{
			name:     "Whitespace trimming",
			model:    "  glm-4.7  ",
			provider: "  z.ai  ",
			expected: "z.ai/glm-4.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ModelSpec{
				Model:    tt.model,
				Provider: tt.provider,
			}
			result := spec.ResolvedModel()
			if result != tt.expected {
				t.Errorf("ResolvedModel() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildResolvedModelList(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		specs    []ModelSpec
		expected []string
	}{
		{
			name:     "Empty specs with base",
			base:     "glm-4.7",
			specs:    []ModelSpec{},
			expected: []string{"glm-4.7"},
		},
		{
			name:     "Multiple specs",
			base:     "",
			specs: []ModelSpec{
				{Model: "glm-4.7", Provider: "z.ai"},
				{Model: "claude-3", Provider: "anthropic"},
			},
			expected: []string{"z.ai/glm-4.7", "anthropic/claude-3"},
		},
		{
			name:     "Mixed valid and invalid specs",
			base:     "fallback-model",
			specs: []ModelSpec{
				{Model: "glm-4.7", Provider: "z.ai"},
				{Model: "", Provider: "anthropic"},
				{Model: "claude-3", Provider: ""},
			},
			expected: []string{"z.ai/glm-4.7", "claude-3"},
		},
		{
			name:     "All empty specs with base",
			base:     "fallback-model",
			specs: []ModelSpec{
				{Model: "", Provider: ""},
				{Model: "", Provider: ""},
			},
			expected: []string{"fallback-model"},
		},
		{
			name:     "No base, all empty specs",
			base:     "",
			specs: []ModelSpec{
				{Model: "", Provider: ""},
				{Model: "", Provider: ""},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildResolvedModelList(tt.base, tt.specs)
			if len(result) != len(tt.expected) {
				t.Errorf("Result length = %d, want %d", len(result), len(tt.expected))
			}
			
			for i, res := range result {
				if i >= len(tt.expected) {
					t.Errorf("Result has extra element at index %d: %s", i, res)
					break
				}
				if res != tt.expected[i] {
					t.Errorf("Result[%d] = %s, want %s", i, res, tt.expected[i])
				}
			}
		})
	}
}

func TestAgentProfilePrepareModels(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		specs          []ModelSpec
		expectedModel  string
		expectedModels []string
	}{
		{
			name:          "Model with specs",
			model:         "old-model",
			specs:         []ModelSpec{{Model: "new-model", Provider: "z.ai"}},
			expectedModel: "z.ai/new-model",
			expectedModels: []string{"z.ai/new-model"},
		},
		{
			name:          "Model without specs",
			model:         "glm-4.7",
			specs:         []ModelSpec{},
			expectedModel: "glm-4.7",
			expectedModels: []string{"glm-4.7"},
		},
		{
			name:          "Empty model with specs",
			model:         "",
			specs:         []ModelSpec{{Model: "glm-4.7", Provider: "z.ai"}},
			expectedModel: "z.ai/glm-4.7",
			expectedModels: []string{"z.ai/glm-4.7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := AgentProfile{
				Model:  tt.model,
				Models: tt.specs,
			}
			
			profile.prepareModels()
			
			if profile.Model != tt.expectedModel {
				t.Errorf("Model = %s, want %s", profile.Model, tt.expectedModel)
			}
			
			if len(profile.ResolvedModels) != len(tt.expectedModels) {
				t.Errorf("ResolvedModels length = %d, want %d", len(profile.ResolvedModels), len(tt.expectedModels))
			}
			
			for i, model := range profile.ResolvedModels {
				if i >= len(tt.expectedModels) {
					t.Errorf("ResolvedModels has extra element at index %d: %s", i, model)
					break
				}
				if model != tt.expectedModels[i] {
					t.Errorf("ResolvedModels[%d] = %s, want %s", i, model, tt.expectedModels[i])
				}
			}
		})
	}
}

func TestAgentProfileModelCandidates(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		specs     []ModelSpec
		expected  []string
	}{
		{
			name:     "With resolved models",
			model:    "old-model",
			specs:    []ModelSpec{{Model: "new-model", Provider: "z.ai"}},
			expected: []string{"z.ai/new-model"},
		},
		{
			name:     "With model only",
			model:    "glm-4.7",
			specs:    []ModelSpec{},
			expected: []string{"glm-4.7"},
		},
		{
			name:     "Empty model and specs",
			model:    "",
			specs:    []ModelSpec{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := AgentProfile{
				Model:          tt.model,
				Models:         tt.specs,
				ResolvedModels: buildResolvedModelList(tt.model, tt.specs),
			}
			
			result := profile.ModelCandidates()
			if len(result) != len(tt.expected) {
				t.Errorf("ModelCandidates length = %d, want %d", len(result), len(tt.expected))
			}
			
			for i, candidate := range result {
				if i >= len(tt.expected) {
					t.Errorf("ModelCandidates has extra element at index %d: %s", i, candidate)
					break
				}
				if candidate != tt.expected[i] {
					t.Errorf("ModelCandidates[%d] = %s, want %s", i, candidate, tt.expected[i])
				}
			}
		})
	}
}

func TestAgentProfileModelCandidatesNilCase(t *testing.T) {
	// Test the nil return case - when both ResolvedModels is nil and Model is empty
	profile := AgentProfile{
		Model:          "",
		Models:         nil,
		ResolvedModels: nil,
	}
	
	result := profile.ModelCandidates()
	if result != nil {
		t.Errorf("ModelCandidates() = %v, want nil", result)
	}
}

func TestAgentProfileModelCandidatesWithEmptyResolvedModels(t *testing.T) {
	// Test the case where ResolvedModels is empty but Model is set
	// This tests the second return branch: return []string{ap.Model}
	profile := AgentProfile{
		Model:          "direct-model",
		Models:         nil,
		ResolvedModels: []string{}, // Empty but initialized slice
	}
	
	result := profile.ModelCandidates()
	if len(result) != 1 || result[0] != "direct-model" {
		t.Errorf("ModelCandidates() = %v, want [direct-model]", result)
	}
}

func TestAgentDefaultsPrepareModels(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		specs          []ModelSpec
		expectedModel  string
		expectedModels []string
	}{
		{
			name:          "Model with specs",
			model:         "old-model",
			specs:         []ModelSpec{{Model: "new-model", Provider: "z.ai"}},
			expectedModel: "z.ai/new-model",
			expectedModels: []string{"z.ai/new-model"},
		},
		{
			name:          "Model without specs",
			model:         "glm-4.7",
			specs:         []ModelSpec{},
			expectedModel: "glm-4.7",
			expectedModels: []string{"glm-4.7"},
		},
		{
			name:          "Empty model with specs",
			model:         "",
			specs:         []ModelSpec{{Model: "glm-4.7", Provider: "z.ai"}},
			expectedModel: "z.ai/glm-4.7",
			expectedModels: []string{"z.ai/glm-4.7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := AgentDefaults{
				Model:  tt.model,
				Models: tt.specs,
			}
			
			defaults.prepareModels()
			
			if defaults.Model != tt.expectedModel {
				t.Errorf("Model = %s, want %s", defaults.Model, tt.expectedModel)
			}
			
			if len(defaults.ResolvedModels) != len(tt.expectedModels) {
				t.Errorf("ResolvedModels length = %d, want %d", len(defaults.ResolvedModels), len(tt.expectedModels))
			}
			
			for i, model := range defaults.ResolvedModels {
				if i >= len(tt.expectedModels) {
					t.Errorf("ResolvedModels has extra element at index %d: %s", i, model)
					break
				}
				if model != tt.expectedModels[i] {
					t.Errorf("ResolvedModels[%d] = %s, want %s", i, model, tt.expectedModels[i])
				}
			}
		})
	}
}

func TestAgentDefaultsModelCandidates(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		specs     []ModelSpec
		expected  []string
	}{
		{
			name:     "With resolved models",
			model:    "old-model",
			specs:    []ModelSpec{{Model: "new-model", Provider: "z.ai"}},
			expected: []string{"z.ai/new-model"},
		},
		{
			name:     "With model only",
			model:    "glm-4.7",
			specs:    []ModelSpec{},
			expected: []string{"glm-4.7"},
		},
		{
			name:     "Empty model and specs",
			model:    "",
			specs:    []ModelSpec{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := AgentDefaults{
				Model:          tt.model,
				Models:         tt.specs,
				ResolvedModels: buildResolvedModelList(tt.model, tt.specs),
			}
			
			result := defaults.ModelCandidates()
			if len(result) != len(tt.expected) {
				t.Errorf("ModelCandidates length = %d, want %d", len(result), len(tt.expected))
			}
			
			for i, candidate := range result {
				if i >= len(tt.expected) {
					t.Errorf("ModelCandidates has extra element at index %d: %s", i, candidate)
					break
				}
				if candidate != tt.expected[i] {
					t.Errorf("ModelCandidates[%d] = %s, want %s", i, candidate, tt.expected[i])
				}
			}
		})
	}
}

func TestAgentDefaultsModelCandidatesNilCase(t *testing.T) {
	// Test the nil return case - when both ResolvedModels is nil and Model is empty
	defaults := AgentDefaults{
		Model:          "",
		Models:         nil,
		ResolvedModels: nil,
	}

	result := defaults.ModelCandidates()
	if result != nil {
		t.Errorf("ModelCandidates() = %v, want nil", result)
	}
}

func TestAgentDefaultsModelCandidatesWithEmptyResolvedModels(t *testing.T) {
	// Test the case where ResolvedModels is empty but Model is set
	// This tests the second return branch: return []string{d.Model}
	defaults := AgentDefaults{
		Model:          "direct-model",
		Models:         nil,
		ResolvedModels: []string{}, // Empty but initialized slice
	}

	result := defaults.ModelCandidates()
	if len(result) != 1 || result[0] != "direct-model" {
		t.Errorf("ModelCandidates() = %v, want [direct-model]", result)
	}
}

func TestConfigPrepareAgentModels(t *testing.T) {
	cfg := &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Model:    "old-model",
				Provider: "old-provider",
				Models: []ModelSpec{
					{Model: "new-model", Provider: "z.ai"},
				},
			},
			Profiles: map[string]AgentProfile{
				"test": {
					Model:  "profile-model",
					Models: []ModelSpec{{Model: "profile-new-model", Provider: "anthropic"}},
				},
			},
		},
	}
	
	cfg.PrepareAgentModels()
	
	// Check defaults
	if cfg.Agents.Defaults.Model != "z.ai/new-model" {
		t.Errorf("Defaults.Model = %s, want z.ai/new-model", cfg.Agents.Defaults.Model)
	}
	if len(cfg.Agents.Defaults.ResolvedModels) != 1 {
		t.Errorf("Defaults.ResolvedModels length = %d, want 1", len(cfg.Agents.Defaults.ResolvedModels))
	}
	if cfg.Agents.Defaults.ResolvedModels[0] != "z.ai/new-model" {
		t.Errorf("Defaults.ResolvedModels[0] = %s, want z.ai/new-model", cfg.Agents.Defaults.ResolvedModels[0])
	}
	
	// Check profiles
	profile := cfg.Agents.Profiles["test"]
	if profile.Model != "anthropic/profile-new-model" {
		t.Errorf("Profile.Model = %s, want anthropic/profile-new-model", profile.Model)
	}
	if len(profile.ResolvedModels) != 1 {
		t.Errorf("Profile.ResolvedModels length = %d, want 1", len(profile.ResolvedModels))
	}
	if profile.ResolvedModels[0] != "anthropic/profile-new-model" {
		t.Errorf("Profile.ResolvedModels[0] = %s, want anthropic/profile-new-model", profile.ResolvedModels[0])
	}
}