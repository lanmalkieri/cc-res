# cc-res - Claude Code Resume

A fast, interactive TUI (Terminal User Interface) for browsing and resuming Claude Code sessions with real-time fuzzy search.

## Features

- 📂 **Auto-discovery** - Scans all Claude Code sessions from `~/.claude/projects/`
- 🔍 **Instant fuzzy search** - Filter input always visible, just start typing
- ⏱️ **Smart sorting** - Sessions ordered by most recently modified
- 📝 **Intelligent summaries** - Shows project context and last activity
- 🚀 **Quick resume** - Launches `claude --resume` with proper working directory
- 🎨 **Clean TUI** - Built with Bubble Tea for smooth interactions

## Installation

### Install from source

```bash
git clone https://github.com/cstobie/cc-res.git
cd cc-res
go build
# Optional: move to PATH
mv cc-res ~/bin/  # or /usr/local/bin
```

### Install with go install

```bash
go install github.com/cstobie/cc-res@latest
```

## Usage

```bash
cc-res
```

The TUI will immediately show all your Claude sessions with a filter input ready for typing.

### Controls

| Key | Action |
|-----|--------|
| **Type anything** | Filter sessions in real-time |
| **↑/↓** or **j/k** | Navigate through sessions |
| **Enter** | Resume selected session |
| **Escape** | Clear current filter |
| **q** | Quit (disabled while filtering) |
| **Ctrl+C** | Force quit anytime |

### Session Display

Each session shows:
- **Summary line** - Project name and first user message or task
- **Session ID** - The unique identifier for the conversation
- **Timestamp** - When the session was last active

Example:
```
> [helmfile] I am seeing this error when trying to start the pod...
  Session: 498fc50a-78b2-454c-bf5e-0dc30caf2b60 | 2025-09-04 09:39
```

### AI-Enhanced Summaries (Optional)

For more intelligent session summaries using Claude:

```bash
USE_AI_SUMMARY=1 cc-res
```

This uses the Claude CLI to generate concise summaries of each session's purpose (requires `claude` CLI installed).

## How It Works

1. **Discovery** - Reads all JSONL files from `~/.claude/projects/*/`
2. **Parsing** - Extracts messages, timestamps, and working directories
3. **Summarization** - Creates concise descriptions of each session
4. **Filtering** - Uses fuzzy matching on summaries and session IDs
5. **Launching** - Runs `claude --resume <session-id>` with the original working directory

## Requirements

- **Go 1.21+** (for building from source)
- **Claude CLI** (`claude`) installed and in PATH
- **Active sessions** in `~/.claude/projects/`
- **Terminal** with UTF-8 support (for emoji icons)

## Project Structure

```
cc-res/
├── main.go         # Main application logic
├── main_test.go    # Test suite
├── go.mod          # Go module definition
├── go.sum          # Dependency checksums
├── README.md       # This file
└── .gitignore      # Git ignore rules
```

## Testing

Run the test suite:

```bash
go test -v
```

Tests cover:
- JSONL parsing
- Session summarization
- Filter functionality
- Path extraction
- Cross-platform compatibility

## Troubleshooting

### No sessions showing up
- Check that `~/.claude/projects/` exists and contains `.jsonl` files
- Verify Claude CLI is creating session files

### Filter not working
- Ensure your terminal supports UTF-8
- Try clearing filter with Escape first

### Session won't resume
- Verify `claude` CLI is installed: `which claude`
- Check that the session ID exists: `ls ~/.claude/projects/*/*.jsonl`

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Author

Created by [cstobie](https://github.com/cstobie)

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Inspired by the need for better Claude Code session management