# VHS demo for riff
# https://github.com/charmbracelet/vhs

Output docs/demo.gif

Set Shell zsh
Set FontSize 16
Set Width 900
Set Height 500
Set Padding 20
Set Theme {"name":"riff","black":"#45475a","red":"#f38ba8","green":"#a6e3a1","yellow":"#f9e2af","blue":"#89b4fa","magenta":"#f5c2e7","cyan":"#94e2d5","white":"#cdd6f4","brightBlack":"#585b70","brightRed":"#f38ba8","brightGreen":"#a6e3a1","brightYellow":"#f9e2af","brightBlue":"#89b4fa","brightMagenta":"#f5c2e7","brightCyan":"#94e2d5","brightWhite":"#a6adc8","background":"#252530","foreground":"#cdd6f4","selection":"#585b70","cursor":"#f5e0dc"}
Set TypingSpeed 100ms
Set CursorBlink false

# Load the riff shell function for auto-cd
Hide
Type `eval "$(command riff init)"`
Enter
Sleep 1s
Show
Ctrl+L
Sleep 300ms

# --- Scene 1: Show help ---
Type "riff --help"
Sleep 0.5s
Enter
Sleep 2.5s
Ctrl+L

# --- Scene 2: List existing projects ---
Type "riff list"
Sleep 300ms
Enter
Sleep 3s
Ctrl+L

# --- Scene 3: Create a new Bun project ---
Type "riff new bun"
Sleep 300ms
Enter
Sleep 5s

# --- Scene 4: Show we're in the project directory ---
Type "ls"
Sleep 300ms
Enter
Sleep 2.5s
Ctrl+L

# --- Scene 5: List again — new project at the top ---
Type "riff list"
Sleep 300ms
Enter
Sleep 3s
Ctrl+L

# --- Scene 6: Archive one project ---
Type "riff clean"
Sleep 300ms
Enter
Sleep 1.5s

# Skip past "All projects", land on the first individual project
Down
Sleep 300ms
# Select it
Type " "
Sleep 500ms
Enter
Sleep 1s

# Confirm archival
Type "y"
Sleep 500ms
Enter
Sleep 2s
Ctrl+L

# --- Scene 7: Show archived projects ---
Type "riff archive"
Sleep 300ms
Enter
Sleep 3s
