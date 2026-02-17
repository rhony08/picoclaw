package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/stretchr/testify/assert"
)

func TestSessionManager_GetCurrentAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	sessionKey := "test:session:1"

	// Initially should return empty string
	agent := sm.GetCurrentAgent(sessionKey)
	assert.Empty(t, agent)

	// Create session and set agent
	session := sm.GetOrCreate(sessionKey)
	session.CurrentAgent = "coder"

	// Should now return the agent
	agent = sm.GetCurrentAgent(sessionKey)
	assert.Equal(t, "coder", agent)
}

func TestSessionManager_SetCurrentAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	sessionKey := "test:session:1"

	// Create session
	sm.GetOrCreate(sessionKey)

	// Set agent
	sm.SetCurrentAgent(sessionKey, "researcher")

	// Verify
	agent := sm.GetCurrentAgent(sessionKey)
	assert.Equal(t, "researcher", agent)

	// Update agent
	sm.SetCurrentAgent(sessionKey, "coder")

	// Verify update
	agent = sm.GetCurrentAgent(sessionKey)
	assert.Equal(t, "coder", agent)
}

func TestSessionManager_SetCurrentAgent_NonExistentSession(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Try to set agent for non-existent session - should not panic
	sm.SetCurrentAgent("non:existent", "coder")

	// Session still doesn't exist
	agent := sm.GetCurrentAgent("non:existent")
	assert.Equal(t, "coder", agent)
}

func TestSessionManager_PersistCurrentAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	sessionKey := "telegram:123456"

	// Create session and set agent
	sm.GetOrCreate(sessionKey)
	sm.SetCurrentAgent(sessionKey, "coder")

	// Save session
	err := sm.Save(sessionKey)
	assert.NoError(t, err)

	// Verify file exists
	sessionFile := filepath.Join(tmpDir, "telegram_123456.json")
	_, err = os.Stat(sessionFile)
	assert.NoError(t, err)

	// Create new session manager (simulating restart)
	sm2 := NewSessionManager(tmpDir)

	// Agent should be loaded from file
	agent := sm2.GetCurrentAgent(sessionKey)
	assert.Equal(t, "coder", agent)
}

func TestSessionManager_CurrentAgentInSessionStruct(t *testing.T) {
	session := &Session{
		Key:          "test:1",
		CurrentAgent: "researcher",
		Messages:     []providers.Message{},
	}

	assert.Equal(t, "researcher", session.CurrentAgent)

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(session)
	assert.NoError(t, err)

	var restored Session
	err = json.Unmarshal(data, &restored)
	assert.NoError(t, err)
	assert.Equal(t, "researcher", restored.CurrentAgent)
}
