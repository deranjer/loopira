---
sidebar_position: 5
---

# MCP integration

Loopira exposes an MCP ([Model Context Protocol](https://modelcontextprotocol.io/))
server at `/mcp` so agents like Claude, Cursor, or ChatGPT can read and
manage issues directly.

## 1. Create an API key

In the app, go to **Settings → API Keys → New API key**. Choose read-write
(can create/edit issues) or read-only. The plaintext key is shown once —
copy it before leaving the page.

## 2. Connect your MCP client

For Claude Code:

```sh
claude mcp add --transport http loopira http://localhost:8080/mcp \
  --header "Authorization: Bearer <your key>"
```

Other clients (Cursor, ChatGPT, etc.) take the same URL and header — check
their MCP configuration docs for the exact syntax.

## Reusing the key for the REST API

The same API key also works as a Bearer token against the regular REST API
(`/api/v1/...`), which is useful for scripts that don't need the full MCP
tool surface.
