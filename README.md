# ah-be-mcp

An [MCP](https://modelcontextprotocol.io) server for Albert Heijn Belgium (ah.be). It lets an AI assistant search products, check the bonus, and manage your winkelmandje. Nederlandse uitleg: [README.nl.md](README.nl.md).

> Unofficial. Not affiliated with Albert Heijn. It uses the private API of the AH app, which AH can change or block at any time, and it may go against AH's terms. Use at your own risk.

It wraps [appie-go](https://github.com/gwillem/appie-go) and borrows its tool set from [ah-mcp](https://github.com/mrserzhan/ah-mcp). AGPL-3.0, like both.

## Install

It runs on your own computer and works with Claude Desktop (macOS, Windows), Claude Code, Codex CLI, Cursor and other local MCP clients. It does not work with ChatGPT or on a Chromebook, because neither can start a program on your computer.

For Claude Desktop, download the `.mcpb` file for your computer from the [latest release](https://github.com/FreDeRoover/ah-be-mcp/releases/latest) (`macos-apple-silicon`, `macos-intel` or `windows`), double-click it and confirm.

For other clients, download the archive from the release or run `go install github.com/FreDeRoover/ah-be-mcp@latest` (Go 1.25+), then point your client at the binary:

```bash
claude mcp add ah-be -- /absolute/path/to/ah-be-mcp
```

```json
{ "mcpServers": { "ah-be": { "command": "/absolute/path/to/ah-be-mcp" } } }
```

## Log in

Searching products and the bonus works without an account. For your winkelmandje, orders, receipts and profile, ask your assistant to call `ah_login`. It opens ah.be in your browser and waits up to 5 minutes. You type your password on ah.be only, so neither this server nor your assistant sees it.

Tokens are saved in your user config directory as `ah-mcp-be/tokens.json`, readable by you only. Set `AH_BE_TOKENS_PATH` to use another file. `ah_logout` deletes them. Anyone who can read that file can act as you on ah.be.

Nothing leaves your computer except calls to AH, and what the tools return becomes part of your conversation with the assistant. `ah_set_cart_item` and `ah_clear_cart` change your real winkelmandje. Nothing here places or pays for an order.

## Tools

`ah_login`, `ah_logout`, `ah_get_member`, `ah_search_products`, `ah_get_product`, `ah_get_bonus`, `ah_get_bonus_group`, `ah_get_cart`, `ah_set_cart_item`, `ah_clear_cart`, `ah_get_orders`, `ah_get_order_details`, `ah_get_receipts`, `ah_get_receipt`.

On ah.be the winkelmandje (ah.be/mijnlijst) is the shopping list, not an online order, and the cart tools read and write that list.

## Troubleshooting

- macOS or Windows warns about an unknown developer: the binaries are not signed. On macOS allow it under System Settings > Privacy & Security, on Windows choose "More info" > "Run anyway".
- The login page doesn't open: the URL is in your client's MCP log. Open it by hand.
- "Not logged in": call `ah_login` again.

## Limitations

- There is no store search. The mobile API only knows Dutch stores, even for Belgian postal codes.
- Tested on macOS against a real account: search, bonus and every winkelmandje action. Not tested: free-text items, orders, receipts, and anything on Windows or Linux.
