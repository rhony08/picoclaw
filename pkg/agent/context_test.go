package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/tools"
)

func TestContextBuilder_NewContextBuilder(t *testing.T) {
	// Create a temporary workspace for testing
	tempDir := t.TempDir()
	
	cb := NewContextBuilder(tempDir)
	if cb == nil {
		t.Fatal("NewContextBuilder returned nil")
	}
	
	if cb.workspace != tempDir {
		t.Errorf("Expected workspace %s, got %s", tempDir, cb.workspace)
	}
	
	if cb.skillsLoader == nil {
		t.Error("Skills loader should not be nil")
	}
	
	if cb.memory == nil {
		t.Error("Memory store should not be nil")
	}
}

func TestContextBuilder_SetToolsRegistry(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Create a mock tool registry
	registry := tools.NewToolRegistry()
	cb.SetToolsRegistry(registry)
	
	if cb.tools != registry {
		t.Error("Tools registry was not set correctly")
	}
}

func TestContextBuilder_getIdentity(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Set up a mock tools registry
	registry := tools.NewToolRegistry()
	cb.SetToolsRegistry(registry)
	
	identity := cb.getIdentity()
	
	if identity == "" {
		t.Error("Identity should not be empty")
	}
	
	// Check that identity contains expected components
	if !contains(identity, "picoclaw") {
		t.Error("Identity should contain 'picoclaw'")
	}
	
	if !contains(identity, "workspace") {
		t.Error("Identity should contain workspace information")
	}
	
	if !contains(identity, "Current Time") {
		t.Error("Identity should contain current time")
	}
}

func TestContextBuilder_buildToolsSection(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test with no tools registry
	result := cb.buildToolsSection()
	if result != "" {
		t.Error("Expected empty result when no tools registry is set")
	}
	
	// Test with empty tools registry
	registry := tools.NewToolRegistry()
	cb.SetToolsRegistry(registry)
	result = cb.buildToolsSection()
	if result != "" {
		t.Error("Expected empty result when no tools are available")
	}
}

func TestContextBuilder_BuildSystemPrompt(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	prompt := cb.BuildSystemPrompt()
	
	if prompt == "" {
		t.Error("System prompt should not be empty")
	}
	
	// Check that prompt contains expected sections
	expectedSections := []string{"picoclaw", "Current Time", "Runtime", "Workspace"}
	for _, section := range expectedSections {
		if !contains(prompt, section) {
			t.Errorf("System prompt should contain section: %s", section)
		}
	}
}

func TestContextBuilder_LoadBootstrapFiles(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test with no bootstrap files
	result := cb.LoadBootstrapFiles()
	if result != "" {
		t.Error("Expected empty result when no bootstrap files exist")
	}
	
	// Create a test bootstrap file
	bootstrapFile := filepath.Join(tempDir, "AGENTS.md")
	content := "Test agents content"
	err := os.WriteFile(bootstrapFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create bootstrap file: %v", err)
	}
	
	result = cb.LoadBootstrapFiles()
	if !contains(result, "AGENTS.md") {
		t.Error("Bootstrap file content should be loaded")
	}
}

func TestContextBuilder_BuildMessages(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test with basic parameters
	history := []providers.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	
	messages := cb.BuildMessages(history, "Previous summary", "New message", nil, "test-channel", "123")
	
	if len(messages) == 0 {
		t.Error("Messages should not be empty")
	}
	
	// Check that system message is first
	if messages[0].Role != "system" {
		t.Error("First message should be system message")
	}
	
	// Check that user message is last
	if messages[len(messages)-1].Role != "user" {
		t.Error("Last message should be user message")
	}
}

func TestContextBuilder_BuildMessagesWithContext(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	data := &ContextData{
		History:     []providers.Message{{Role: "user", Content: "Hello"}},
		Summary:     "Previous summary",
		UserMessage: "New message",
		Media:       []string{"test.jpg"},
		Channel:     "test-channel",
		ChatID:      "123",
		SystemPrompt: "Custom system prompt",
	}
	
	messages := cb.BuildMessagesWithContext(data)
	
	if len(messages) == 0 {
		t.Error("Messages should not be empty")
	}
	
	// Check that system prompt contains custom prompt
	systemPrompt := messages[0].Content
	if !contains(systemPrompt, "Custom system prompt") {
		t.Error("System prompt should contain custom prompt")
	}
	
	// Check that session info is included
	if !contains(systemPrompt, "Current Session") {
		t.Error("System prompt should contain session info")
	}
}

func TestContextBuilder_AddToolResult(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	messages := []providers.Message{
		{Role: "user", Content: "Hello"},
	}
	
	toolCallID := "test-call-id"
	toolName := "test-tool"
	result := "Tool execution result"
	
	updatedMessages := cb.AddToolResult(messages, toolCallID, toolName, result)
	
	if len(updatedMessages) != 2 {
		t.Error("Should have 2 messages after adding tool result")
	}
	
	if updatedMessages[1].Role != "tool" {
		t.Error("Added message should be a tool result")
	}
	
	if updatedMessages[1].Content != result {
		t.Errorf("Expected tool result '%s', got '%s'", result, updatedMessages[1].Content)
	}
	
	if updatedMessages[1].ToolCallID != toolCallID {
		t.Errorf("Expected tool call ID '%s', got '%s'", toolCallID, updatedMessages[1].ToolCallID)
	}
}

func TestContextBuilder_AddAssistantMessage(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	messages := []providers.Message{
		{Role: "user", Content: "Hello"},
	}
	
	content := "Assistant response"
	toolCalls := []map[string]interface{}{
		{"name": "test-tool", "arguments": "{}"},
	}
	
	updatedMessages := cb.AddAssistantMessage(messages, content, toolCalls)
	
	if len(updatedMessages) != 2 {
		t.Error("Should have 2 messages after adding assistant message")
	}
	
	if updatedMessages[1].Role != "assistant" {
		t.Error("Added message should be an assistant message")
	}
	
	if updatedMessages[1].Content != content {
		t.Errorf("Expected assistant content '%s', got '%s'", content, updatedMessages[1].Content)
	}
}

func TestContextBuilder_GetSkillsInfo(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	info := cb.GetSkillsInfo()
	
	if info == nil {
		t.Error("Skills info should not be nil")
	}
	
	if info["total"] == nil {
		t.Error("Skills info should contain total count")
	}
	
	if info["available"] == nil {
		t.Error("Skills info should contain available count")
	}
	
	if info["names"] == nil {
		t.Error("Skills info should contain names")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// Helper function to find index of substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Test for MemoryStore functionality
func TestMemoryStore(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test memory context retrieval
	memoryContext := cb.memory.GetMemoryContext()
	
	// Should be empty initially or contain default content
	if memoryContext == "" {
		// This is expected if no memory files exist
	} else {
		// If there is content, it should be properly formatted
		if !contains(memoryContext, "Memory") {
			t.Error("Memory context should be properly formatted")
		}
	}
}

// Test for SkillsLoader integration
func TestSkillsLoaderIntegration(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test skills summary building
	skillsSummary := cb.skillsLoader.BuildSkillsSummary()
	
	// Should be empty initially or contain valid skills content
	if skillsSummary != "" {
		if !contains(skillsSummary, "Skills") {
			t.Error("Skills summary should be properly formatted")
		}
	}
}

// Test error handling for invalid workspace
func TestContextBuilder_InvalidWorkspace(t *testing.T) {
	// Test with empty workspace
	cb := NewContextBuilder("")
	if cb == nil {
		t.Error("Context builder should handle empty workspace")
	}
	
	// Test with workspace that doesn't exist
	cb = NewContextBuilder("/non/existent/path")
	if cb == nil {
		t.Error("Context builder should handle non-existent workspace")
	}
}

// Test context building with various edge cases
func TestContextBuilder_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	cb := NewContextBuilder(tempDir)
	
	// Test with empty history
	messages := cb.BuildMessages([]providers.Message{}, "", "", nil, "", "")
	if len(messages) == 0 {
		t.Error("Should handle empty history gracefully")
	}
	
	// Test with very long user message
	longMessage := string(make([]byte, 10000)) // 10KB message
	messages = cb.BuildMessages([]providers.Message{}, "", longMessage, nil, "", "")
	if len(messages) == 0 {
		t.Error("Should handle long messages gracefully")
	}
	
	// Test with special characters in messages
	specialChars := "Hello 世界! 🚀"
	messages = cb.BuildMessages([]providers.Message{}, "", specialChars, nil, "", "")
	if len(messages) == 0 {
		t.Error("Should handle special characters gracefully")
	}
}