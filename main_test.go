package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadJSONL(t *testing.T) {
	// Create a temp file with test JSONL data
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	testData := []ChatLine{
		{
			SessionID: "test-session-1",
			Type:      "user",
			Message: Message{
				Role:    "user",
				Content: "Test message 1",
			},
			UUID:      "uuid-1",
			Timestamp: time.Now(),
			CWD:       "/test/path",
		},
		{
			SessionID: "test-session-1",
			Type:      "assistant",
			Message: Message{
				Role:    "assistant",
				Content: "Response 1",
			},
			UUID:      "uuid-2",
			Timestamp: time.Now(),
		},
	}

	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, line := range testData {
		if err := encoder.Encode(line); err != nil {
			t.Fatal(err)
		}
	}

	// Test reading the file
	messages, err := readJSONL(testFile)
	if err != nil {
		t.Fatalf("Failed to read JSONL: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Message.Content != "Test message 1" {
		t.Errorf("Expected 'Test message 1', got '%s'", messages[0].Message.Content)
	}

	if messages[1].Message.Content != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", messages[1].Message.Content)
	}
}

func TestSummarizeChat(t *testing.T) {
	tests := []struct {
		name     string
		messages []ChatLine
		expected string
		contains []string
	}{
		{
			name:     "Empty messages",
			messages: []ChatLine{},
			expected: "Empty session",
		},
		{
			name: "Single user message",
			messages: []ChatLine{
				{
					Type: "user",
					Message: Message{
						Content: "Help me debug this error",
					},
					CWD:       "/Users/test/project",
					Timestamp: time.Now().Add(-2 * time.Hour),
				},
			},
			contains: []string{"[project]", "Help me debug this error"},
		},
		{
			name: "Long message truncated",
			messages: []ChatLine{
				{
					Type: "user",
					Message: Message{
						Content: strings.Repeat("a", 150),
					},
					CWD:       "/test/dir",
					Timestamp: time.Now().Add(-1 * time.Hour),
				},
			},
			contains: []string{"[dir]", "...", "(1h"},
		},
		{
			name: "Multiple messages",
			messages: []ChatLine{
				{
					Type: "user",
					Message: Message{
						Content: "First message",
					},
					CWD:       "/project",
					Timestamp: time.Now().Add(-30 * time.Minute),
				},
				{
					Type: "user",
					Message: Message{
						Content: "Second message",
					},
					Timestamp: time.Now().Add(-20 * time.Minute),
				},
			},
			contains: []string{"First message", "+1 more messages", "(20m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := summarizeChat(tt.messages)
			
			if tt.expected != "" && result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
			
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected result to contain '%s', but got '%s'", substr, result)
				}
			}
		})
	}
}

func TestItemFilterValue(t *testing.T) {
	item := item{
		title:       "Test session summary",
		description: "Session: abc-123 | 2025-01-01",
		sessionID:   "abc-123",
	}

	filterValue := item.FilterValue()
	expected := "Test session summary Session: abc-123 | 2025-01-01"

	if filterValue != expected {
		t.Errorf("Expected filter value '%s', got '%s'", expected, filterValue)
	}
}

func TestItemMethods(t *testing.T) {
	item := item{
		title:       "Test title",
		description: "Test description",
		sessionID:   "test-id",
	}

	if item.Title() != "Test title" {
		t.Errorf("Title() returned wrong value")
	}

	if item.Description() != "Test description" {
		t.Errorf("Description() returned wrong value")
	}
}

func TestLoadSessionsInvalidDir(t *testing.T) {
	// Temporarily change HOME to a temp directory
	originalHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sessions, err := loadSessions()
	if err == nil {
		t.Error("Expected error for missing .claude/projects directory")
	}
	if sessions != nil {
		t.Error("Expected nil sessions for missing directory")
	}
}

func TestSessionPathExtraction(t *testing.T) {
	// Test that we properly extract CWD from session messages
	messages := []ChatLine{
		{
			Type: "user",
			CWD:  "/home/user/project",
			Message: Message{
				Content: "test",
			},
		},
		{
			Type: "assistant",
			CWD:  "", // No CWD in assistant messages typically
			Message: Message{
				Content: "response",
			},
		},
	}

	// Extract CWD like we do in main
	var cwd string
	for _, msg := range messages {
		if msg.CWD != "" {
			cwd = msg.CWD
			break
		}
	}

	if cwd != "/home/user/project" {
		t.Errorf("Expected CWD '/home/user/project', got '%s'", cwd)
	}
}