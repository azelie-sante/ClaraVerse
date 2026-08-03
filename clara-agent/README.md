# Clara Agent

Clara Agent is ClaraVerse's local coding agent CLI: a lightweight, terminal-native agent that talks directly to your configured LLM provider and runs tools (read, write, edit, bash) locally on your machine.

It is a fork of [Pi](https://github.com/earendil-works/pi) by Mario Zechner and Armin Ronacher, used under the MIT license (see [LICENSE](LICENSE)). Clara Agent adds a ClaraVerse account/provider integration on top so it shares your ClaraVerse model configuration and personal memory, and pushes finished sessions back into your ClaraVerse chat history.

## Packages

| Package | Description |
|---------|-------------|
| **[@claraverse/clara-agent](packages/coding-agent)** | Interactive coding agent CLI |
| **[@earendil-works/pi-agent-core](packages/agent)** | Agent runtime: tool calling and state management |
| **[@earendil-works/pi-ai](packages/ai)** | Unified multi-provider LLM API (OpenAI, Anthropic, Google, and more) |
| **[@earendil-works/pi-tui](packages/tui)** | Terminal UI library with differential rendering |

The internal package names under `@earendil-works` are unchanged from upstream Pi to keep the diff against upstream small; only the CLI itself (`@claraverse/clara-agent`) is renamed and rebranded.

## Log in with your ClaraVerse account

Start the agent, then log in from inside the session:

```bash
clara-agent
```

```
> /login claraverse
```

This opens the same device-authorization flow ClaraVerse's other connected-device features use: you get a short code, confirm it in your browser while logged into ClaraVerse, and Clara Agent picks up your account's configured model and provider automatically.

By default Clara Agent talks to `http://localhost:3000`. Point it at a different ClaraVerse instance with `CLARAVERSE_URL`:

```bash
CLARAVERSE_URL=https://your-instance.example.com clara-agent
```

## Permissions & Containerization

Clara Agent does not include a built-in permission system for restricting filesystem, process, network, or credential access. By default, it runs with the permissions of the user and process that launched it. Be deliberate about where you run it.
