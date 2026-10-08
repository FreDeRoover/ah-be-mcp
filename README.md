# ah-mcp-be

Stdio MCP server for Albert Heijn Belgium (ah.be). Thin wrapper around
[appie-go](https://github.com/gwillem/appie-go) (`WithSite("be")`), tool set modelled on
[ah-mcp](https://github.com/mrserzhan/ah-mcp). Licensed AGPL-3.0, like both.

```bash
go build -o ah-mcp-be .
```

Claude Desktop / Claude Code config:

```json
{ "mcpServers": { "ah-be": { "command": "/Users/frederik/Projects/ah-mcp-be/ah-mcp-be" } } }
```

Product search and bonus work anonymously. Everything personal (cart, lists, orders,
receipts, profile) needs `ah_login`: it opens ah.be login in your browser and blocks
until done. Tokens live in `~/Library/Application Support/ah-mcp-be/tokens.json`
(override with `AH_BE_TOKENS_PATH`) and refresh automatically.

Not included vs. ah-mcp: SSE/HTTP transports, remote OAuth proxy, store search and
vandaag-af bargains, reopen/edit submitted orders, favourite-list editing.
