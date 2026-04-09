Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme { "name": "stacklit", "black": "#0d1117", "red": "#f85149", "green": "#56d364", "yellow": "#e3b341", "blue": "#58a6ff", "magenta": "#bc8cff", "cyan": "#39d2c0", "white": "#c9d1d9", "brightBlack": "#484f58", "brightRed": "#f85149", "brightGreen": "#56d364", "brightYellow": "#e3b341", "brightBlue": "#58a6ff", "brightMagenta": "#bc8cff", "brightCyan": "#39d2c0", "brightWhite": "#e6edf3", "background": "#0d1117", "foreground": "#c9d1d9", "selectionBackground": "#264f78", "cursorColor": "#c9d1d9" }
Set Padding 20

Type "# Stacklit: your codebase, in 1,500 tokens"
Enter
Sleep 1s

Type "stacklit init"
Enter
Sleep 3s

Type ""
Sleep 1s

Type "# stacklit.json is now committed to git"
Enter
Sleep 500ms

Type "cat stacklit.json | head -25"
Enter
Sleep 3s

Type ""
Sleep 1s

Type "# Start MCP server for AI agents"
Enter
Sleep 500ms

Type "# stacklit serve"
Enter
Sleep 2s
