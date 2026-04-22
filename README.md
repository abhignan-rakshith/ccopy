# ccopy

**Interactive TUI file picker that concatenates selected files and copies them to the system clipboard — built for pasting codebase context into LLM chats.**

![demo placeholder](demo.gif)

## Install

```sh
go install github.com/abhignan-rakshith/ccopy/cmd/ccopy@latest
```

Or from source:

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
ccopy --format markdown      # fenced code blocks with lang hints
ccopy --no-gitignore         # don't respect .gitignore
ccopy --max-size 5MB         # raise the large-file threshold
ccopy --dry-run              # print to stdout instead of clipboard
```

## Keybindings

| Key              | Action                                        |
| ---------------- | --------------------------------------------- |
| `↑`/`↓` or `j/k` | move cursor                                   |
| `←`/`→` or `h/l` | collapse / expand directory                   |
| `space`          | toggle selection (dir = select all under it)  |
| `a`              | select all                                    |
| `n`              | clear selection                               |
| `i`              | invert selection                              |
| `/`              | filter tree (type to narrow, `esc` to cancel) |
| `enter`          | copy selected files to clipboard and exit     |
| `q` / `esc`      | quit without copying                          |
| `?`              | toggle help overlay                           |

## Output formats

- **`tail`** (default) — `==> path <==` header before each file's contents. Same shape as `tail -n +1 file1 file2 ...`.
- **`markdown`** — `## path` heading + fenced code block with a language hint inferred from the extension.
- **`xml`** — `<file path="...">...</file>` — recommended when pasting into Claude.

## Default exclusions

Skipped during walk: `.git`, `node_modules`, `.venv`, `venv`, `__pycache__`, `.DS_Store`, `dist`, `build`, `target`, plus anything matched by a `.gitignore` in the walked tree.

Binary files (detected by scanning the first 512 bytes for a null byte) appear dimmed in the tree and cannot be selected. Files over the size limit (1 MB by default) appear highlighted and prompt for confirmation when selected.

## Configuration

Put overrides in `~/.config/ccopy/config.toml`:

```toml
format = "xml"
max_file_size = 2097152          # 2 MiB
respect_gitignore = true
exclude_patterns = [".git", "node_modules", "target", "vendor"]
```

All keys are optional — unspecified keys keep the built-in defaults.

## Clipboard support

- **macOS**: works out of the box (uses `pbcopy`)
- **Linux**: requires `xclip` or `xsel` (X11) or `wl-copy` (Wayland)
- **Windows**: works out of the box

If the clipboard is unavailable, `ccopy` falls back to writing the output to stdout.

## Record a demo

```sh
# Install vhs: brew install charmbracelet/tap/vhs
vhs demo.tape
```

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

## License

MIT
