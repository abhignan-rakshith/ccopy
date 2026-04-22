# ccopy

Interactive TUI file picker that concatenates selected files and copies them to the clipboard — built for pasting codebase context into LLM chats.

![ccopy demo](demo.gif)

## Install

```sh
go install github.com/abhignan-rakshith/ccopy/cmd/ccopy@latest
```

From source:

```sh
git clone https://github.com/abhignan-rakshith/ccopy
cd ccopy
make install
```

## Usage

```sh
ccopy                        # pick files in the current directory
ccopy ~/projects/api         # pick in a specific directory
ccopy --format xml           # XML output (nice for Claude)
ccopy --format markdown      # fenced code blocks with language hints
ccopy --no-gitignore         # don't respect .gitignore
ccopy --max-size 5MB         # raise the large-file threshold
ccopy --dry-run              # print to stdout instead of clipboard
```

### Keybindings

| Key              | Action                                        |
| ---------------- | --------------------------------------------- |
| `↑`/`↓` or `j/k` | move cursor                                   |
| `←`/`→` or `h/l` | collapse / expand directory                   |
| `space`          | toggle selection (dir = select all under it)  |
| `a` / `n` / `i`  | select all / clear / invert                   |
| `/`              | filter tree (type to narrow, `esc` to cancel) |
| `enter`          | copy selected files to clipboard and exit     |
| `q` / `esc`      | quit without copying                          |
| `?`              | toggle help overlay                           |

### Output formats

- **`tail`** *(default)* — `==> path <==` header per file, matching `tail -n +1 file1 file2 ...`
- **`markdown`** — `## path` heading + fenced code block with a language hint from the extension
- **`xml`** — `<file path="...">...</file>` — recommended for Claude

### Exclusions

Skipped by default: `.git`, `node_modules`, `.venv`, `venv`, `__pycache__`, `.DS_Store`, `dist`, `build`, `target`, plus anything matched by a `.gitignore` in the walked tree.

Binary files (detected via null byte in the first 512 bytes) appear dimmed and cannot be selected. Files over 1 MB appear highlighted and prompt for confirmation.

## Configuration

Overrides live in `~/.config/ccopy/config.toml` — all keys optional:

```toml
format = "xml"
max_file_size = 2097152          # 2 MiB
respect_gitignore = true
exclude_patterns = [".git", "node_modules", "target", "vendor"]
```

## Clipboard support

- **macOS** — works out of the box (`pbcopy`)
- **Linux** — requires `xclip`, `xsel`, or `wl-copy`
- **Windows** — works out of the box

Falls back to stdout if the clipboard is unavailable.

## Development

```sh
make test    # run all unit tests
make build   # build into ./bin/ccopy
make vet     # run go vet
make fmt     # format with gofmt
```

Layout:

```
cmd/ccopy/          entrypoint
internal/tree/      fs walker + gitignore + binary detection
internal/selection/ selection set + tri-state dir logic
internal/formatter/ tail / markdown / xml output
internal/tui/       Bubble Tea model/update/view
internal/config/    TOML loader with defaults
```

## Record a demo

```sh
brew install charmbracelet/tap/vhs
vhs demo.tape
```

## License

MIT
