package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatLine struct {
	ParentUUID  *string   `json:"parentUuid"`
	IsSidechain bool      `json:"isSidechain"`
	UserType    string    `json:"userType"`
	CWD         string    `json:"cwd"`
	SessionID   string    `json:"sessionId"`
	Version     string    `json:"version"`
	GitBranch   string    `json:"gitBranch"`
	Type        string    `json:"type"`
	Message     Message   `json:"message"`
	UUID        string    `json:"uuid"`
	Timestamp   time.Time `json:"timestamp"`
}

type Session struct {
	ProjectPath string
	SessionID   string
	FilePath    string
	ModTime     time.Time
	Messages    []ChatLine
	Summary     string
	LastActive  time.Time
}

type model struct {
	list     list.Model
	sessions []Session
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		
		case "q":
			if !m.list.SettingFilter() && !m.list.IsFiltered() {
				m.quitting = true
				return m, tea.Quit
			}

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.sessionID
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return "Cancelled.\n"
	}
	return "\n" + m.list.View()
}

type item struct {
	title       string
	description string
	sessionID   string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title + " " + i.description }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 3 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n  %s", i.title, i.description)

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("170")).
				Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func readJSONL(filePath string) ([]ChatLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []ChatLine
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024*1024), 64*1024*1024) // 64MB buffer

	for scanner.Scan() {
		var msg ChatLine
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, scanner.Err()
}

func summarizeChat(messages []ChatLine) string {
	if len(messages) == 0 {
		return "Empty session"
	}

	var userMessages []string
	var assistantActions []string
	projectDir := ""

	for _, msg := range messages {
		if msg.CWD != "" && projectDir == "" {
			projectDir = filepath.Base(msg.CWD)
		}

		switch msg.Type {
		case "user":
			content := strings.TrimSpace(msg.Message.Content)
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			userMessages = append(userMessages, content)
		case "assistant":
			if strings.Contains(msg.Message.Content, "tool_calls") {
				assistantActions = append(assistantActions, "executed commands")
			}
		}
	}

	summary := fmt.Sprintf("[%s] ", projectDir)
	if len(userMessages) > 0 {
		summary += userMessages[0]
		if len(userMessages) > 1 {
			summary += fmt.Sprintf(" (+%d more messages)", len(userMessages)-1)
		}
	}

	lastMsg := messages[len(messages)-1]
	duration := time.Since(lastMsg.Timestamp)
	if duration < 24*time.Hour {
		summary += fmt.Sprintf(" (%s ago)", duration.Round(time.Minute))
	} else {
		summary += fmt.Sprintf(" (%s)", lastMsg.Timestamp.Format("Jan 2"))
	}

	return summary
}

func summarizeWithClaude(messages []ChatLine) (string, error) {
	if len(messages) == 0 {
		return "Empty session", nil
	}

	var conversation strings.Builder
	for i, msg := range messages {
		if i > 20 {
			break
		}
		if msg.Type == "user" || msg.Type == "assistant" {
			role := msg.Type
			content := msg.Message.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			conversation.WriteString(fmt.Sprintf("%s: %s\n", role, content))
		}
	}

	prompt := fmt.Sprintf(`Summarize this conversation in one concise line (max 100 chars). Focus on the main task or problem being addressed:

%s

Summary:`, conversation.String())

	cmd := exec.Command("claude", "--no-conversation", prompt)
	output, err := cmd.Output()
	if err != nil {
		return summarizeChat(messages), nil
	}

	summary := strings.TrimSpace(string(output))
	if len(summary) > 100 {
		summary = summary[:100]
	}

	return summary, nil
}

func loadSessions() ([]Session, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var sessions []Session

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())
		files, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".jsonl") {
				continue
			}

			filePath := filepath.Join(projectPath, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			messages, err := readJSONL(filePath)
			if err != nil {
				continue
			}

			sessionID := strings.TrimSuffix(file.Name(), ".jsonl")
			
			// Get the timestamp of the last message for better sorting
			lastActive := info.ModTime()
			if len(messages) > 0 {
				lastActive = messages[len(messages)-1].Timestamp
			}
			
			session := Session{
				ProjectPath: entry.Name(),
				SessionID:   sessionID,
				FilePath:    filePath,
				ModTime:     info.ModTime(),
				Messages:    messages,
				LastActive:  lastActive,
			}

			if os.Getenv("USE_AI_SUMMARY") == "1" {
				session.Summary, _ = summarizeWithClaude(messages)
			} else {
				session.Summary = summarizeChat(messages)
			}

			sessions = append(sessions, session)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive.After(sessions[j].LastActive)
	})

	return sessions, nil
}

func main() {
	sessions, err := loadSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No Claude sessions found in ~/.claude/projects/")
		os.Exit(0)
	}

	items := make([]list.Item, len(sessions))
	sessionMap := make(map[string]*Session)

	for i, session := range sessions {
		desc := fmt.Sprintf("Session: %s | %s",
			session.SessionID,
			session.ModTime.Format("2006-01-02 15:04"),
		)
		items[i] = item{
			title:       session.Summary,
			description: desc,
			sessionID:   session.SessionID,
		}
		sessionMap[session.SessionID] = &sessions[i]
	}

	const defaultWidth = 120
	const listHeight = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Claude Code Sessions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true) // Always show filter
	l.FilterInput.Placeholder = "Type to filter..."
	l.FilterInput.CharLimit = 100
	l.FilterInput.Prompt = "🔍 Filter: "
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("99"))
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(4)
	l.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{list: l, sessions: sessions}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m := finalModel.(model); m.choice != "" {
		if session, ok := sessionMap[m.choice]; ok {
			fmt.Printf("Resuming session: %s\n", session.SessionID)
			
			// Find the working directory from the session messages
			var cwd string
			for _, msg := range session.Messages {
				if msg.CWD != "" {
					cwd = msg.CWD
					// Verify the directory still exists
					if _, err := os.Stat(cwd); err == nil {
						break
					}
					// Directory doesn't exist, keep looking for another one
					cwd = ""
				}
			}
			
			args := []string{"--resume", session.SessionID}
			
			cmd := exec.Command("claude", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			// Set the working directory if we found one
			if cwd != "" {
				cmd.Dir = cwd
			}
			
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error launching claude: %v\n", err)
				os.Exit(1)
			}
		}
	}
}