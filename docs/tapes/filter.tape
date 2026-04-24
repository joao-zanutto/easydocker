# to be used with https://github.com/charmbracelet/vhs

Output filter.gif

Set Width 1920
Set Height 800
Set Shell zsh
Set PlaybackSpeed 1.0

# Launch the TUI
Type "easydocker"
Sleep 1s
Enter
Sleep 4s

# Show container filtering
Type "/"
Sleep 500ms
Type "postgres"
Sleep 1500ms
Escape
Sleep 300ms

# Move down to specific container
Down
Sleep 300ms
Enter
Sleep 3s

# Filter logs
Type "/"
Sleep 500ms
Type "warning"
Sleep 2s
