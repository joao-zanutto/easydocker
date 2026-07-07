# to be used with https://github.com/charmbracelet/vhs

Output example.webm

Set FontSize 18
Set Width 1920
Set Height 800
Set Shell zsh
Set PlaybackSpeed 1.3

# Launch the TUI
Type "easydocker"
Sleep 1s
Enter
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 300ms

# Navigate images
Right
Sleep 1500ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 1500ms

# Navigate networks
Right
Sleep 2000ms
# Navigate volumes
Right
Sleep 2000ms

# Go back to containers
Left
Sleep 1000ms
Left
Sleep 1000ms
Left
Sleep 2000ms

# Filter and navigate container logs
Type "/"
Sleep 300ms
Type "airflow"
Enter
Sleep 300ms
Enter
Sleep 1000ms
Down
Sleep 500ms
Enter
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up 10
Sleep 300ms
Up 10
Sleep 300ms

# Filter
Type "/"
Sleep 300ms
Type "warning"
Enter
Sleep 1000ms
Escape
Sleep 300ms

# Show help
Escape
Sleep 300ms
Enter
Sleep 300ms
Escape
Sleep 300ms
Escape
Sleep 300ms

# Shell
Type "s"
Sleep 300ms
Type "ls"
Enter
Sleep 300ms
Type "pwd"
Enter

Escape
Sleep 2s