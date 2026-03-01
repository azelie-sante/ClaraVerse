MCP Workspace Directory - Python Execution Environment
======================================================

Welcome to your isolated MCP (Model Context Protocol) workspace!

OVERVIEW:
This workspace provides a completely isolated Python environment where you can:
- Execute Python code safely without affecting your system
- Install packages that won't interfere with system Python
- Save and load files in a dedicated workspace
- Run shell commands with automatic Python/pip routing

WORKSPACE FEATURES:
✓ Isolated Python Virtual Environment (.venv/)
✓ Clean workspace for file operations
✓ Cross-platform shell command support
✓ Automatic dependency management
✓ Safe package installation

AVAILABLE TOOLS:

1. py(code="...")
   - Execute Python code in isolated environment
   - Auto-prints last line expressions
   - Supports multi-line code, imports, functions
   - Examples:
     py(code="import math; math.sqrt(16)")
     py(code="[x**2 for x in range(5)]")
     py(code="def greet(name): return f'Hello {name}!'")

2. Shell: Windows PowerShell
   - powershell(cmd="...") on Windows / sh(cmd="...") on Unix
   - Execute system commands
   - Auto-routes python/pip to virtual environment
   - Examples:
     powershell(cmd="Get-Process python")
     sh(cmd="ps aux | grep python")

3. pip(pkg="...")
   - Install Python packages safely
   - Only affects this workspace environment
   - Examples:
     pip(pkg="requests")
     pip(pkg="numpy pandas matplotlib")
     pip(pkg="beautifulsoup4==4.9.3")

4. save(name="...", text="...")
   - Save content to workspace files
   - Persistent across MCP session
   - Examples:
     save(name="script.py", text="print('Hello World')")
     save(name="data.json", text='{"key": "value"}')

5. load(name="...")
   - Read file content from workspace
   - Access previously saved files
   - Examples:
     load(name="script.py")
     load(name="data.json")

6. ls()
   - List all workspace files and directories
   - Shows file sizes and types
   - Excludes .venv for clarity

7. open()
   - Open workspace in system file manager
   - Direct access to workspace folder
   - Platform-specific file manager

GETTING STARTED:
1. Check Python version: py(code="import sys; print(sys.version)")
2. Install a package: pip(pkg="requests")
3. Test the package: py(code="import requests; print('Requests installed!')")
4. Save a script: save(name="test.py", text="print('Hello from saved file')")
5. List files: ls()
6. Load and run: py(code=load(name="test.py"))

WORKSPACE STRUCTURE:
├── README.txt          (this file)
├── .venv/             (Python virtual environment - hidden from ls())
├── your_files.py      (files you save)
├── data_files.json    (data you create)
└── any_other_files    (content you work with)

TIPS FOR AI MODELS:
- Use py() for Python calculations, data processing, API calls
- Use pip() to install libraries before using them in py()
- Use save()/load() to persist code and data between operations
- Use ls() to see what files are available
- All operations are isolated and safe to experiment with
- Files persist within the same MCP session

VIRTUAL ENVIRONMENT DETAILS:
- Location: .venv/
- Python: Isolated Python 3.x installation
- Packages: Separated from system Python
- Activation: Automatic for all py() and python commands

This workspace is your sandbox - experiment freely!
