# ah-be-mcp

Stdio [MCP](https://modelcontextprotocol.io) server for Albert Heijn Belgium (ah.be). Search products, check bonus offers, manage your winkelmandje, and look up orders and receipts from any MCP client.

Nederlandse uitleg: [README.nl.md](README.nl.md).

> **Unofficial.** Not affiliated with, endorsed by, or supported by Albert Heijn. It uses the same private API as the AH app, which is not meant for third parties. AH can change or block it at any time, and using it may go against AH's terms. Use at your own risk.

A thin wrapper around [appie-go](https://github.com/gwillem/appie-go) (`WithSite("be")`), with a tool set modelled on [ah-mcp](https://github.com/mrserzhan/ah-mcp). Licensed AGPL-3.0, like both.

## Install

### Claude Desktop (one click)

1. Download the `.mcpb` file for your computer from the [latest release](https://github.com/FreDeRoover/ah-be-mcp/releases/latest):
   `macos-apple-silicon` (M1 or newer), `macos-intel`, or `windows`.
2. Double-click it, or drag it into Claude Desktop, and confirm the install.
3. Ask Claude to log in to AH (see [Log in](#log-in)).

### Other MCP clients (Codex, Cursor, Claude Code, ...)

Download the archive for your platform from the release, or use Go 1.25+:

```bash
go install github.com/FreDeRoover/ah-be-mcp@latest
```

Then point your client at the absolute path of the binary, e.g. Claude Code:

```bash
claude mcp add ah-be -- /absolute/path/to/ah-be-mcp
```

or in a JSON config:

```json
{ "mcpServers": { "ah-be": { "command": "/absolute/path/to/ah-be-mcp" } } }
```

### From source

```bash
git clone https://github.com/FreDeRoover/ah-be-mcp.git
cd ah-be-mcp
go build -o ah-be-mcp .
```

## Log in

Product search and bonus work anonymously. Everything personal (winkelmandje, orders, receipts, profile) needs a login: ask your assistant to call `ah_login`. It opens the ah.be login page in your browser on the machine running the server and waits (max 5 minutes) until you are done. You only enter your password on ah.be itself; this server and your AI assistant never see it.

## Privacy and security

- Everything runs on your own computer. There is no server of ours and no telemetry.
- Login tokens are stored in your user config directory (`ah-mcp-be/tokens.json`, e.g. `~/Library/Application Support/` on macOS, `%AppData%` on Windows, `~/.config/` on Linux), readable by you only (mode 0600), and refresh automatically. Set `AH_BE_TOKENS_PATH` to use another file. `ah_logout` deletes them.
- Anyone who can read that file can act as you on ah.be, so don't share it.
- Your AI assistant sees what the tools return (products, your winkelmandje, receipts, profile) as part of the conversation, and is subject to your AI provider's privacy terms.
- `ah_clear_cart` and `ah_set_cart_item` change your real winkelmandje. Nothing here places or pays for an order.

## Tools

| Tool | What it does |
|---|---|
| `ah_login`, `ah_logout` | Log in via the browser, or forget the stored tokens |
| `ah_get_member` | Profile and bonus card |
| `ah_search_products` | Search products (Dutch terms work best), optionally bonus only |
| `ah_get_product` | Product details, optionally with nutritional info |
| `ah_get_bonus`, `ah_get_bonus_group` | Current bonus offers, and the products in a bonus group |
| `ah_get_cart` | Your winkelmandje (ah.be/mijnlijst) with prices, bonus and an estimated total |
| `ah_set_cart_item` | Add a product or free-text item, or change its quantity (0 removes) |
| `ah_clear_cart` | Empty the winkelmandje (needs `confirm=yes`) |
| `ah_get_orders`, `ah_get_order_details` | Scheduled deliveries/pickups and their items |
| `ah_get_receipts`, `ah_get_receipt` | In-store receipts (kassabonnen) |

On ah.be the winkelmandje is the shopping list, not an online order, so the cart tools read and write that list.

## Troubleshooting

- **macOS says the app can't be opened / is from an unidentified developer.** The binaries are not signed or notarized. In Terminal, run `xattr -dr com.apple.quarantine "<folder of the extension or binary>"`, or allow it under System Settings > Privacy & Security.
- **Windows SmartScreen warns.** Same reason (unsigned). Choose "More info" > "Run anyway".
- **Login page doesn't open.** The URL is also written to the server log (stderr). Check your client's MCP logs and open it manually.
- **"Not logged in".** Call `ah_login` again; tokens may have expired or been revoked.

## Limitations

- Local use only (stdio, browser login). No SSE/HTTP transport, so ChatGPT's web connectors can't use it.
- No store search: the mobile API only knows Dutch stores, even for Belgian postal codes.
- Tested against a real account on macOS: search, bonus, winkelmandje read/add/change/remove/clear. **Not yet tested:** free-text items, orders and receipts, and everything on Windows and Linux (including opening the login page on Windows).
- Not included compared to ah-mcp: vandaag-af bargains, reopening or editing submitted orders, favourite-list editing.

## Releasing

Tag a version and push it; GitHub Actions builds everything with `scripts/release.sh` and publishes the release.

```bash
git tag v0.1.0 && git push origin v0.1.0
```
