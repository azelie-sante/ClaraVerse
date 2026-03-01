---
title: "How We Built Clara Companion: A Self-Hosted MCP Bridge for Remote AI Tool Access"
slug: how-we-built-mcp-bridge
authors: [badboysm890, aruntemme]
tags: [mcp, technical, golang, open-source]
description: "Learn how Clara Companion bridges local MCP servers to the cloud over WebSocket, giving your self-hosted AI tools real-time access to filesystem, git, and more."
keywords:
  - MCP bridge
  - Model Context Protocol
  - self-hosted AI tools
  - Clara Companion
  - ClaraVerse
  - WebSocket bridge
  - remote tool execution
  - MCP server
  - Go CLI
  - open source AI
image: /img/features/claracompanion.png
---

# How We Built Clara Companion: A Self-Hosted MCP Bridge for Remote AI Tool Access

MCP (Model Context Protocol) is powerful -- but only when it can reach your machine. We built Clara Companion to solve that exact gap: a Go CLI that bridges local MCP servers to ClaraVerse over WebSocket so your AI assistant can access filesystem, git, and memory tools from anywhere. Here is the story of why we built it, how it works under the hood, and how you can set it up in under two minutes.

<!-- truncate -->

## Why We Needed a Model Context Protocol Bridge

When Anthropic released MCP in late 2024, we immediately saw the potential. MCP gives AI models structured access to external tools -- file systems, databases, version control, and more. But there was a fundamental problem: **MCP servers run locally, and cloud-hosted AI lives remotely.**

We had a strong conviction: MCP should be connected where the machine is, but AI should still be able to build things and expand its abilities even during a chat session. The user's local environment is where the real work happens -- files, repositories, running services. The AI needs to reach all of that in real time.

We started building Clara Companion on **December 15, 2024** and shipped the first release on **January 1, 2025**. We wanted to beat others to market with this bridge concept, and while the broader MCP ecosystem moved fast, the core idea has held up: a lightweight, persistent tunnel between your local tools and a cloud AI platform.

## How the WebSocket Bridge Architecture Works

Clara Companion is a single Go binary that does three things:

1. **Starts local MCP servers** (filesystem, git, memory, or any custom server)
2. **Opens a persistent WebSocket connection** to the ClaraVerse backend
3. **Registers available tools** and proxies execution requests in real time

The architecture is intentionally simple. When you chat with Clara in the browser, and it decides to read a file or run a git command, the request flows like this:

```
Browser Chat --> ClaraVerse Backend --> WebSocket --> Clara Companion --> MCP Server --> Your Machine
```

The bridge handles all the complexity: tool registration, heartbeats, token refresh, exponential backoff on disconnects, and graceful reconnection with automatic re-registration.

Here is the core of the WebSocket bridge from our Go source:

```go
// Bridge manages the WebSocket connection to the backend
type Bridge struct {
    backendURL     string
    authToken      string
    conn           *websocket.Conn
    writeChan      chan Message
    onToolCall     func(ToolCall)
    onReconnect    func()
}

// Messages flow as typed JSON over the socket
type Message struct {
    Type    string                 `json:"type"`
    Payload map[string]interface{} `json:"payload"`
}
```

Every tool call from the backend arrives as a `tool_call` message. The companion executes it against the appropriate MCP server and sends back a `tool_result`. The write channel is buffered (100 messages), and both application-level heartbeats (30s) and WebSocket-level pings (45s) keep the connection alive through proxies and firewalls.

## Why We Chose Go for Cross-Platform MCP Tooling

Go was the obvious choice for a CLI that needs to run everywhere. A single `go build` produces a statically linked binary with no runtime dependencies. We ship prebuilt binaries for:

- **macOS** (Intel and Apple Silicon)
- **Linux** (amd64)
- **Windows** (amd64)

No Python virtual environments, no Node.js version managers, no Docker required on the client side. Download, authenticate, run. The binary is typically under 15 MB and starts in milliseconds.

Go's goroutine model also maps cleanly to the bridge's concurrency needs: one goroutine reads from the WebSocket, one writes, and tool calls execute concurrently without blocking the connection.

## Installing and Setting Up Clara Companion

Getting started takes three commands:

```bash
# 1. Install the binary (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/ClaraVerse/ClaraVerse/main/backend/mcp-bridge/scripts/install.sh | sh

# 2. Authenticate with your ClaraVerse account
clara_companion login

# 3. Launch the interactive dashboard
clara_companion
```

On Windows, use PowerShell:

```powershell
irm https://raw.githubusercontent.com/ClaraVerse/ClaraVerse/main/backend/mcp-bridge/scripts/install.ps1 | iex
```

The `login` command uses an **OAuth 2.0 Device Authorization Grant** (RFC 8628). It displays a short code in your terminal, opens your browser, and waits for you to authorize. No passwords are pasted into the CLI, and it works with any auth provider -- email, Google, GitHub.

```
To authenticate, open this URL in your browser:
  https://claraverse.app/device

Then enter this code:

   ┌────────────────┐
   │   ABCD-1234    │
   └────────────────┘

Waiting for authorization... (expires in 15:00)
```

Once authorized, the companion saves credentials to `~/.claraverse/mcp-config.yaml` and launches the TUI dashboard.

## Adding and Managing Self-Hosted MCP Servers

Clara Companion ships with three default MCP servers that cover the most common use cases:

| Server       | What It Does                      | Package                                        |
|-------------|-----------------------------------|-------------------------------------------------|
| `filesystem` | Read, write, and list files       | `@modelcontextprotocol/server-filesystem`       |
| `git`        | Status, diff, commit, log         | `@modelcontextprotocol/server-git`              |
| `memory`     | Persistent knowledge graph        | `@modelcontextprotocol/server-memory`           |

You can add any MCP-compatible server with a single command:

```bash
# Add a browser automation server
clara_companion add browser --command npx --args @browsermcp/mcp@latest

# Add a custom server by path
clara_companion add mytools --command /usr/local/bin/my-mcp-server

# List all configured servers
clara_companion list

# Remove a server
clara_companion remove browser
```

The configuration lives in a single YAML file:

```yaml
backend_url: wss://claraverse.app/mcp/connect
device:
  device_id: dev_abc123
  refresh_token: eyJ...
  user_email: user@example.com
mcp_servers:
  - name: filesystem
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
    enabled: true
  - name: git
    command: npx
    args: ["-y", "@modelcontextprotocol/server-git"]
    enabled: true
```

## Running as a Background Service on macOS and Linux

For always-on access, Clara Companion can install itself as a background service:

```bash
clara_companion service install
```

On **macOS**, this creates a `launchd` plist at `~/Library/LaunchAgents/com.claraverse.clara-companion.plist` that starts the bridge automatically on login and restarts it if it crashes.

On **Linux**, it creates a `systemd` user service at `~/.config/systemd/user/claraverse-mcp.service` with `Restart=always` and a 10-second restart delay.

```bash
# Check if the service is running
clara_companion service status

# View logs
tail -f ~/.claraverse/logs/clara_companion.log

# Stop or restart
clara_companion service stop
clara_companion service start
```

The bridge handles network interruptions gracefully. If the WebSocket drops, it reconnects with exponential backoff (starting at 1 second, capping at 60 seconds) and re-registers all tools automatically. Token refresh happens preemptively -- five minutes before expiry -- so sessions do not interrupt mid-conversation.

## What We Learned Building an Early MCP Bridge

Building Clara Companion before the MCP ecosystem matured taught us a few things:

**Ship the simplest thing that works.** Our first version was just WebSocket + tool proxy. No TUI, no service management, no device auth. That core loop -- connect, register, proxy, respond -- has barely changed since December 2024.

**Go is the right tool for CLI distribution.** We considered Node.js (ecosystem alignment) and Rust (performance), but Go's build simplicity and goroutine model won out. Cross-compiling for six targets is a single shell script.

**Device auth is worth the effort.** Pasting tokens into a terminal is a security risk and a poor UX. The OAuth 2.0 Device Grant flow adds a few hundred lines of code but makes onboarding feel polished.

The MCP ecosystem has grown rapidly since we started. But the fundamental problem -- bridging local tools to remote AI -- remains relevant whether you are self-hosting or using a managed platform. Clara Companion gives you that bridge with a single binary and zero infrastructure.

---

Ready to connect your local tools to ClaraVerse? Check out the [full Companion documentation](/docs/features/companion) or grab the source on [GitHub](https://github.com/ClaraVerse/ClaraVerse).
