---
title: Clara Companion
sidebar_label: Clara Companion
sidebar_position: 10
---

# Clara Companion -- MCP Bridge

Connect your local tools and filesystem to ClaraVerse via the Clara Companion CLI. It bridges local MCP servers to your ClaraVerse instance over WebSocket.

![Clara Companion](/img/features/claracompanion.png)

## Quick Start

```bash
# Install via the claraverse CLI
claraverse companion

# Login (choose default localhost:3000 or enter your server URL)
clara_companion login

# Start the bridge
clara_companion
```

Or install manually from [GitHub Releases](https://github.com/claraverse-space/ClaraVerse/releases).

## Default MCP Servers

These servers are available out of the box (requires Node.js for npx):

| Server | Description | Package |
|--------|-------------|---------|
| `filesystem` | File operations (read, write, list) | `@modelcontextprotocol/server-filesystem` |
| `git` | Git operations (status, diff, commit) | `@modelcontextprotocol/server-git` |
| `memory` | Persistent knowledge graph | `@modelcontextprotocol/server-memory` |

## CLI Commands

| Command | Description |
|---------|-------------|
| `login` | Setup wizard (auth + servers + service + start) |
| `logout` | Log out and revoke this device |
| `start` | Start the bridge |
| `status` | Show connection status |
| `list` | List configured servers |
| `add <name>` | Add an MCP server |
| `remove <name>` | Remove an MCP server |
| `service install` | Install as background service |
| `service uninstall` | Remove background service |
| `service status` | Check service status |
| `service start/stop` | Start or stop service |
| `devices list` | List connected devices |
| `devices revoke <id>` | Revoke a device |
| `devices rename <id> <name>` | Rename a device |

Running `clara_companion` without arguments starts the bridge.

## Background Service

Install as a background service so the bridge auto-starts on login:

**macOS** (launchd):

```bash
clara_companion service install
# Location: ~/Library/LaunchAgents/com.claraverse.mcp-client.plist
```

**Linux** (systemd):

```bash
clara_companion service install
# Location: ~/.config/systemd/user/claraverse-mcp.service
```

**Windows:** Background service not yet supported. Use the Startup folder or Task Scheduler.

Logs: `~/.claraverse/logs/mcp-client.log`

## Config File

Location: `~/.claraverse/mcp-config.yaml`

```yaml
backend_url: wss://your-instance.com/mcp/connect
device:
  device_id: dev_abc123
  refresh_token: eyJ...
  user_email: user@example.com
mcp_servers:
  - name: filesystem
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/Users/name"]
    enabled: true
```

## Platform Support

| Platform | Auth | MCP Servers | Background Service |
|----------|------|-------------|-------------------|
| macOS | Yes | Yes | Yes (launchd) |
| Linux | Yes | Yes | Yes (systemd) |
| Windows | Yes | Yes | Manual start only |

## Troubleshooting

**Auth issues:**

```bash
clara_companion status   # Check token
clara_companion login    # Re-authenticate
```

**Connection issues:**
- Check your ClaraVerse backend is running
- Check firewall allows WebSocket connections
- Run `clara_companion status` for diagnostics

**Service issues:**

```bash
clara_companion service status
tail -f ~/.claraverse/logs/mcp-client.log
```
