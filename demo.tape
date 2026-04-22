# vhs script — renders a demo GIF for ccopy.
# Prerequisites:
#   - `ccopy` installed into $GOPATH/bin (run `make install` first)
#   - a sample directory at /tmp/ccopy-demo (see README for setup)
# Record with: vhs demo.tape

Set FontSize 14
Set Width 1000
Set Height 640
Set Theme "Dracula"
Set Padding 12

Output demo.gif

Env PATH "$HOME/go/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"

Type "cd /tmp/ccopy-demo && ccopy"
Sleep 400ms
Enter
Sleep 1.5s

# Pick a couple of files
Down
Space
Sleep 400ms
Down
Space
Sleep 400ms
Right
Sleep 400ms
Down
Space
Sleep 800ms

# Filter
Type "/"
Type "main"
Sleep 800ms
Escape
Sleep 400ms

# Help overlay
Type "?"
Sleep 1.2s
Escape
Sleep 400ms

# Commit selection
Enter
Sleep 1.5s
