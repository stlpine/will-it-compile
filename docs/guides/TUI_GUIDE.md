# TUI (Terminal User Interface) Guide

## Overview

The will-it-compile TUI provides an interactive terminal interface for submitting code to the API server and viewing compilation results in real-time.

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea), the TUI offers:
- Interactive code editor
- File loading from disk
- Live job monitoring
- Job history browser
- Keyboard-driven navigation

## Prerequisites

- API server running (typically at `http://localhost:8080`)
- Terminal with support for ANSI colors and escape codes
- Go 1.25+ (for building)

## Quick Start

### Option 1: Using the Demo Script

```bash
./scripts/demo-tui.sh
```

This script will:
1. Check if the API server is running
2. Start it if needed
3. Build the TUI if not already built
4. Launch the TUI with helpful tips

### Option 2: Manual Start

```bash
# Terminal 1: Start API server
make run

# Terminal 2: Start TUI
make run-tui
```

### Option 3: Connect to Remote API

```bash
API_URL=http://your-server:8080 ./bin/will-it-compile-tui
```

## User Interface

### Main Views

The TUI has 5 main views:

1. **Editor View** (default): Write or paste code
2. **History View**: Browse previous compilation jobs
3. **Job Detail View**: See detailed compilation results
4. **File Picker**: Select files to load
5. **Help Screen**: Keyboard shortcuts reference

### Status Bar

Located at the bottom, shows:
- **Left**: API server URL
- **Right**: Current status (Ready, Compiling, Errors)
- Color coding:
  - Blue: Normal status
  - Green: Success
  - Red: Errors
  - Yellow: Processing

## Keyboard Shortcuts

### Global Shortcuts

| Key | Action |
|-----|--------|
| `?` | Show help screen |
| `Esc` | Return to editor |
| `q` / `Ctrl+C` | Quit application |
| `Tab` | Toggle between editor and history |

### Editor View

| Key | Action |
|-----|--------|
| `Enter` | Submit code for compilation |
| `f` | Open file picker |
| `l` | Cycle through languages (C++ → C → Go → Rust → C++) |
| `Ctrl+L` | Clear editor |
| Normal typing | Edit code |
| Arrow keys | Navigate text |

### History View

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up in list |
| `↓` / `j` | Move down in list |
| `Enter` | View job details |
| `d` | Delete job from history |
| `c` | Clear all history |

### Job Detail View

| Key | Action |
|-----|--------|
| `r` | Refresh job status (manual poll) |
| `Backspace` | Return to history |

### File Picker

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate files/directories |
| `Enter` | Select file or enter directory |
| `Esc` | Cancel and return to editor |

## Workflows

### Basic Compilation

1. **Start TUI**: `./bin/will-it-compile-tui`
2. **Write Code**: Type directly in the editor
3. **Select Language**: Press `l` to cycle to desired language
4. **Compile**: Press `Enter`
5. **View Results**: TUI automatically switches to job detail view
6. **Return**: Press `Esc` to go back to editor

### Loading from File

1. **Open Picker**: Press `f` in editor
2. **Navigate**: Use arrow keys to find file
3. **Select**: Press `Enter` on desired file
4. **Compile**: Press `Enter` to submit

### Reviewing History

1. **Open History**: Press `Tab` from editor
2. **Browse**: Use `↑`/`↓` to navigate jobs
3. **View Details**: Press `Enter` on a job
4. **Return**: Press `Esc` to go back

## Features

### Language Support

Currently supported languages:
- **C++** (.cpp, .cc, .cxx, .c++)
- **C** (.c)
- **Go** (.go)
- **Rust** (.rs)

Cycle through languages with the `l` key.

### Job Status Indicators

In history view, jobs show status icons:
- `✓` Green: Compilation succeeded
- `✗` Red: Compilation failed
- `●` Yellow: Currently processing
- `○` Gray: Queued

### Live Monitoring

When you submit a job:
1. TUI shows "Compiling..." status
2. Automatically polls API every 500ms
3. Updates view when job completes
4. Shows spinner animation while waiting

### File Type Detection

File picker only shows supported file types:
- C++: .cpp, .cc, .cxx, .c++
- C: .c
- Go: .go
- Rust: .rs

Attempting to select other files shows an error.

### Output Display

Job detail view shows:
- **Job Info**: ID, status, language, timestamps
- **Result Summary**: Success/failure indicator
- **Exit Code**: Process exit code
- **Duration**: Compilation time
- **Stdout**: Standard output (first 500 chars)
- **Stderr**: Standard error (first 500 chars)

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | `http://localhost:8080` | API server base URL |

### Terminal Requirements

For best experience, use a terminal with:
- 256 color support
- UTF-8 encoding
- Minimum 80x24 size
- Mouse support (optional but enhanced)

## Troubleshooting

### "API server unreachable"

**Problem**: TUI shows connection error in status bar

**Solutions**:
1. Check API server is running: `curl http://localhost:8080/health`
2. Start API server: `make run` in another terminal
3. Check firewall settings
4. Verify `API_URL` environment variable

### "File type not supported"

**Problem**: Can't load certain files

**Solution**: Only files with extensions .cpp, .cc, .cxx, .c++, .c, .go, .rs are supported

### TUI doesn't display correctly

**Problem**: Garbled text or broken layout

**Solutions**:
1. Use a modern terminal (iTerm2, Alacritty, Wezterm)
2. Ensure terminal supports ANSI escape codes
3. Check terminal size is at least 80x24
4. Try different `TERM` value: `TERM=xterm-256color ./bin/will-it-compile-tui`

### Compilation stuck "Processing"

**Problem**: Job stays in processing state

**Solutions**:
1. Check API server logs
2. Verify Docker is running
3. Check Docker images exist: `make docker-build`
4. Press `r` in job detail view to manually refresh

### Can't quit TUI

**Problem**: Normal quit keys don't work

**Solution**:
- Try `Ctrl+C` multiple times
- Force quit terminal session
- Kill process: `pkill will-it-compile-tui`

## Tips & Tricks

### Productivity Tips

1. **Use File Loading**: Press `f` to quickly load test files
2. **Keep History**: Don't clear history (`c`) to compare results
3. **Language Switching**: Use `l` to test same code in different languages
4. **Keyboard Navigation**: Master `Tab`, `Esc`, `?` for fast navigation

### Workflow Optimization

1. **Prepare Test Files**: Keep .cpp files in `tests/samples/`
2. **Multiple Terminals**: Run TUI in one, edit code in another
3. **API Monitoring**: Watch API logs in separate terminal
4. **Script Integration**: Use demo script for quick setup

### Visual Customization

The TUI uses lipgloss for styling. Colors are defined in `cmd/tui/ui/styles.go`:
- Primary: Blue (#39)
- Success: Green (#42)
- Error: Red (#196)
- Warning: Yellow (#220)

## Advanced Usage

### Batch Testing

While TUI is interactive, you can prepare files for quick testing:

```bash
# Create test files
mkdir -p tests/my-tests
echo 'int main() { return 0; }' > tests/my-tests/valid.cpp
echo 'int main() { invalid }' > tests/my-tests/invalid.cpp

# Use TUI to test each file
./bin/will-it-compile-tui
# Press 'f', navigate to tests/my-tests, select file, compile
```

### Integration with Editors

Some editors can launch external tools:

**Vim/Neovim**:
```vim
:!./bin/will-it-compile-tui
```

**VS Code**:
Add to tasks.json:
```json
{
  "label": "Test with TUI",
  "type": "shell",
  "command": "./bin/will-it-compile-tui"
}
```

### Custom API Endpoints

For development with different API instances:

```bash
# Local development
API_URL=http://localhost:3000 ./bin/will-it-compile-tui

# Remote server
API_URL=https://compile.example.com ./bin/will-it-compile-tui

# Docker container
API_URL=http://172.17.0.2:8080 ./bin/will-it-compile-tui
```

## Comparison: CLI vs TUI

| Feature | CLI | TUI |
|---------|-----|-----|
| Mode | Batch processing | Interactive |
| File input | Command args | File picker or paste |
| Results | One-time output | Persistent history |
| Monitoring | Synchronous wait | Live polling |
| Multiple jobs | Separate commands | Tab navigation |
| Best for | Scripts, CI/CD | Development, testing |

## Architecture

### Component Structure

```
cmd/tui/
├── main.go              # Entry point
├── client/
│   └── client.go        # HTTP API client
└── ui/
    ├── model.go         # Bubbletea model
    ├── views.go         # View rendering
    ├── handlers.go      # Keyboard handlers
    └── styles.go        # Visual styles
```

### State Management

The TUI uses Bubbletea's Elm architecture:
- **Model**: Application state (job history, current view, etc.)
- **Update**: State transitions based on messages
- **View**: Render UI from current state

### API Communication

All API calls are asynchronous commands that return messages:
- `healthCheckMsg`: Server health status
- `environmentsMsg`: Available compilation environments
- `compileResultMsg`: New job submitted
- `jobUpdateMsg`: Job status updated

## Contributing

To extend the TUI:

1. **Add new view**: Add to `ViewState` enum and implement in `views.go`
2. **Add keyboard shortcut**: Update handlers in `handlers.go`
3. **Change styling**: Modify `styles.go`
4. **Add feature**: Update model in `model.go`

See [CLAUDE.md](./CLAUDE.md) for project architecture details.

## Resources

- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [API Documentation](./README.md#api-documentation)

## Support

For issues or questions:
1. Check this guide and [README.md](./README.md)
2. Review [CLAUDE.md](./CLAUDE.md) for architecture
3. Check API server logs
4. File an issue on GitHub

---

**Last Updated**: 2025-11-10
**TUI Version**: 1.0 (MVP)
