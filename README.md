# ah-be-mcp

Stdio [MCP](https://modelcontextprotocol.io) server for Albert Heijn Belgium (ah.be). Search products, check bonus offers, manage your winkelmandje, and look up orders and receipts from any MCP client.

A thin wrapper around [appie-go](https://github.com/gwillem/appie-go) (`WithSite("be")`), with a tool set modelled on [ah-mcp](https://github.com/mrserzhan/ah-mcp). Licensed AGPL-3.0, like both.

## Install

Requires Go 1.25+.

```bash
go install github.com/FreDeRoover/ah-be-mcp@latest
```

The binary lands in `$(go env GOPATH)/bin/ah-be-mcp`. Or build from source:

```bash
git clone https://github.com/FreDeRoover/ah-be-mcp.git
cd ah-be-mcp
go build -o ah-be-mcp .
```

## Configure

Add it to your MCP client, using the absolute path to the binary.

Claude Desktop (`claude_desktop_config.json`):

```json
{ "mcpServers": { "ah-be": { "command": "/absolute/path/to/ah-be-mcp" } } }
```

Claude Code:

```bash
claude mcp add ah-be -- /absolute/path/to/ah-be-mcp
```

## Log in

Product search and bonus work anonymously. Everything personal (winkelmandje, orders, receipts, profile) needs a login: ask your assistant to call `ah_login`. It opens the ah.be login page in your browser on the machine running the server and waits (max 5 minutes) until you are done. You only enter your password on ah.be itself.

Tokens are stored in your user config directory (`ah-mcp-be/tokens.json`, e.g. `~/Library/Application Support/` on macOS or `~/.config/` on Linux), mode 0600, and refresh automatically. Set `AH_BE_TOKENS_PATH` to use another file. `ah_logout` deletes them.

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

## Limitations

- Unofficial: it uses the same private API as the AH app and can break whenever AH changes it. Not affiliated with Albert Heijn.
- Local use only (stdio, browser login). No SSE/HTTP transport or remote login.
- No store search: the mobile API only knows Dutch stores, even for Belgian postal codes.
- Vrije-tekstitems (`name` in `ah_set_cart_item`) and the orders/receipts tools are untested on ah.be; the rest was checked against a real account.
- Not included compared to ah-mcp: vandaag-af bargains, reopening or editing submitted orders, favourite-list editing.
