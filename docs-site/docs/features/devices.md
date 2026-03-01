---
title: Devices
sidebar_label: Devices
sidebar_position: 9
---

# Devices

Connect multiple machines to your ClaraVerse instance. Each device running [Clara Companion](./companion.md) registers itself and exposes its local MCP servers to Clara.

## How It Works

1. Install [Clara Companion](./companion.md) on any machine
2. Run `mcp-client login` to authenticate with your ClaraVerse instance
3. The device appears in your **Settings > Devices** tab
4. Clara can now use MCP tools from that device remotely

## Managing Devices

Open **Settings** and navigate to the **Devices** tab. From here you can:

- **View connected devices** -- See device name, platform, version, last active time, and location
- **Rename a device** -- Give it a friendly name for easy identification
- **Revoke a device** -- Disconnect and remove a device from your account

You can also manage devices from the CLI:

```bash
mcp-client devices list
mcp-client devices rename <device-id> "My Laptop"
mcp-client devices revoke <device-id>
```

## Use Cases

- **Work + Home** -- Connect both machines so Clara has access to files on either
- **Server access** -- Run Clara Companion on a remote server to give Clara access to server-side tools
- **Team setups** -- Each team member connects their own device with their own MCP servers

## Tips

- Each device maintains its own set of MCP servers configured in `~/.claraverse/mcp-config.yaml`
- Devices reconnect automatically if the connection drops
- Revoking a device immediately disconnects it and removes its tools from Clara
- For production deployments, use a reverse proxy with TLS so device communication is encrypted over WSS
