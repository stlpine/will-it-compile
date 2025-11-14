# TUI Quick Start

## 1. Start the TUI

```bash
# Easy way (automatic setup)
./scripts/demo-tui.sh

# Or manual way
make run          # Terminal 1: Start API server
make run-tui      # Terminal 2: Start TUI
```

## 2. Essential Keys

| Key | What It Does |
|-----|--------------|
| `?` | Show help |
| `f` | Load a file |
| `l` | Change language |
| `Enter` | Compile code |
| `Tab` | View history |
| `Esc` | Go back |
| `q` | Quit |

## 3. Quick Workflow

```
1. Press 'f' → navigate to tests/samples/hello.cpp → press Enter
2. Press Enter again to compile
3. View results
4. Press Tab to see history
5. Press q to quit
```

## 4. Troubleshooting

**"API server unreachable"**
```bash
# Start the API server
make run
```

**Need help?**
- Press `?` in the TUI
- Read [TUI_GUIDE.md](./TUI_GUIDE.md)
- Check [README.md](./README.md)

## Sample Test File

Created for you at `tests/samples/hello.cpp`:
```cpp
#include <iostream>

int main() {
    std::cout << "Hello from will-it-compile TUI!" << std::endl;
    return 0;
}
```
