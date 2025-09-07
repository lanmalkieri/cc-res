# cc-res - Claude Code Resume

A TUI (Terminal User Interface) wrapper for Claude Code that helps you easily browse and resume previous chat sessions.

## Features

- Lists all Claude Code sessions from `~/.claude/projects/`
- Shows sessions sorted by most recently modified
- **Fuzzy search filtering** - start typing to find sessions instantly
- Provides automatic summaries of each session
- Interactive selection using arrow keys
- Launches `claude --resume` with the correct session ID and working directory

## Installation

```bash
go install github.com/cstobie/cc-res@latest
```

Or build from source:

```bash
git clone https://github.com/cstobie/cc-res.git
cd cc-res
go build
```

## Usage

Simply run:

```bash
cc-res
```

### Controls

- **Just start typing**: Filter sessions in real-time
- **↑/↓** or **j/k**: Navigate through sessions
- **Enter**: Select and resume session
- **Escape**: Clear search filter
- **q** or **Ctrl+C**: Quit (disabled while filtering)

### AI-Powered Summaries

By default, cc-res generates basic summaries from the chat content. For better summaries using Claude AI:

```bash
USE_AI_SUMMARY=1 cc-res
```

This will use the Claude CLI to generate more intelligent summaries of your sessions (requires `claude` CLI to be installed and configured).

## How It Works

1. Scans `~/.claude/projects/` for all project folders
2. Reads JSONL files containing chat histories
3. Parses and summarizes each session
4. Presents an interactive list sorted by modification time
5. Launches `claude --resume <session-id> --cwd <original-path>` for the selected session

## Requirements

- Go 1.21+
- Claude Code CLI (`claude`) installed and in PATH
- Existing Claude Code sessions in `~/.claude/projects/`

## License

MIT

## Contributing

Pull requests welcome! Please feel free to submit issues and enhancement requests.

## Author

Created by cstobie