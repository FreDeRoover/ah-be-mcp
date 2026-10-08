// ah-mcp-be: stdio MCP server for Albert Heijn Belgium (ah.be), built on
// github.com/gwillem/appie-go. Local use only: login opens your browser.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	appie "github.com/gwillem/appie-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// version is set at build time: -ldflags "-X main.version=1.2.3"
var version = "dev"

var (
	client     *appie.Client
	tokensPath string
	anonymous  bool // true when only an anonymous (browse-only) token is held
)

// handler returns any JSON-serialisable value; errors become tool errors.
type handler func(ctx context.Context, req mcp.CallToolRequest) (any, error)

// add registers a tool. auth=true means the user must be logged in; otherwise
// an anonymous token is fetched on demand so browsing works without login.
func add(s *server.MCPServer, auth bool, h handler, name, desc string, opts ...mcp.ToolOption) {
	opts = append([]mcp.ToolOption{mcp.WithDescription(desc)}, opts...)
	s.AddTool(mcp.NewTool(name, opts...), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if auth && (anonymous || !client.IsAuthenticated()) {
			return mcp.NewToolResultError("Not logged in. Call ah_login first."), nil
		}
		if !auth && !client.IsAuthenticated() {
			if err := client.GetAnonymousToken(ctx); err != nil {
				return mcp.NewToolResultError("anonymous token: " + err.Error()), nil
			}
			anonymous = true
		}
		out, err := h(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, err := json.Marshal(out)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})
}

func main() {
	// appie-go prints the login URL to stdout, which would corrupt the stdio
	// protocol. Keep the real stdout for MCP and send anything fmt prints to
	// stderr instead. appie-go only opens the browser itself on macOS/Linux, so
	// on Windows we open the printed URL ourselves.
	mcpOut := os.Stdout
	if r, w, err := os.Pipe(); err == nil {
		os.Stdout = w
		go func() {
			sc := bufio.NewScanner(r)
			for sc.Scan() {
				line := sc.Text()
				fmt.Fprintln(os.Stderr, line)
				if runtime.GOOS == "windows" && strings.HasPrefix(line, "http") {
					_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", line).Start()
				}
			}
		}()
	} else {
		os.Stdout = os.Stderr
	}

	cfg, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	tokensPath = os.Getenv("AH_BE_TOKENS_PATH")
	if tokensPath == "" {
		tokensPath = filepath.Join(cfg, "ah-mcp-be", "tokens.json")
	}
	if err := os.MkdirAll(filepath.Dir(tokensPath), 0o700); err != nil {
		fatal(err)
	}
	client, err = appie.NewWithConfig(tokensPath, appie.WithSite("be"))
	if err != nil {
		fatal(err)
	}

	s := server.NewMCPServer("Albert Heijn BE", version)
	registerTools(s)

	if err := server.NewStdioServer(s).Listen(context.Background(), os.Stdin, mcpOut); err != nil {
		fatal(err)
	}
}

const listPath = "/mobile-services/shoppinglist/v2/items"

// v2 list item as returned by GET; also the shape PATCH wants back (quantity 0 deletes).
type listItem struct {
	Quantity    int    `json:"quantity"`
	Type        string `json:"type"`
	OriginCode  string `json:"originCode"`
	Description string `json:"description,omitempty"`
	Product     struct {
		ID       int     `json:"webshopId"`
		Title    string  `json:"title"`
		Size     string  `json:"salesUnitSize"`
		Price    float64 `json:"currentPrice"`
		Was      float64 `json:"priceBeforeBonus"`
		Bonus    string  `json:"bonusMechanism"`
		Category string  `json:"mainCategory"`
	} `json:"product"`
}

func (i listItem) title() string {
	if i.Product.Title != "" {
		return i.Product.Title
	}
	return i.Description
}

func getCart(ctx context.Context) ([]listItem, error) {
	var raw struct {
		Items []struct {
			listItem
			Details struct {
				Product json.RawMessage `json:"product"`
			} `json:"productDetails"`
		} `json:"items"`
	}
	if err := client.DoRequest(ctx, "GET", listPath, nil, &raw); err != nil {
		return nil, err
	}
	items := make([]listItem, 0, len(raw.Items))
	for _, r := range raw.Items {
		it := r.listItem
		if len(r.Details.Product) > 0 {
			_ = json.Unmarshal(r.Details.Product, &it.Product)
		}
		items = append(items, it)
	}
	return items, nil
}

func patchCart(ctx context.Context, items []map[string]any) error {
	return client.DoRequest(ctx, "PATCH", listPath, map[string]any{"items": items}, nil)
}

// cartItem turns {product_id | name, quantity} into the item shape PATCH wants.
func cartItem(ctx context.Context, spec map[string]any) (map[string]any, error) {
	qty := 1
	if v, ok := spec["quantity"].(float64); ok {
		qty = int(v)
	}
	it := map[string]any{"quantity": qty, "type": "SHOPPABLE", "originCode": "PRD", "strikeThrough": false}
	if id, _ := spec["product_id"].(float64); id > 0 {
		prod, err := client.GetProduct(ctx, int(id)) // the API wants the title as description
		if err != nil {
			return nil, fmt.Errorf("product %d: %w", int(id), err)
		}
		it["productId"], it["description"], it["searchTerm"] = int(id), prod.Title, prod.Title
	} else if name, _ := spec["name"].(string); name != "" {
		it["originCode"], it["description"] = "TXT", name
	} else {
		return nil, fmt.Errorf("each item needs a product_id or a name")
	}
	return it, nil
}

// productLine is a compact product view (no image URLs) for multi-product tools.
type productLine struct {
	ID         int                     `json:"product_id"`
	Title      string                  `json:"title"`
	Size       string                  `json:"size,omitempty"`
	Price      float64                 `json:"price"`
	Was        float64                 `json:"price_before_bonus,omitempty"`
	Bonus      string                  `json:"bonus,omitempty"`
	Segment    string                  `json:"bonus_segment_id,omitempty"` // expand with ah_get_bonus_group
	NutriScore string                  `json:"nutriscore,omitempty"`
	Nutrition  []appie.NutritionalInfo `json:"nutrition_per_100g,omitempty"`
	Error      string                  `json:"error,omitempty"`
}

func lineOf(p appie.Product) productLine {
	l := productLine{ID: p.ID, Title: p.Title, Size: p.UnitSize, Price: p.Price.Now, Segment: p.BonusSegmentID, NutriScore: p.NutriScore, Nutrition: p.NutritionalInfo}
	if p.IsBonus {
		l.Was, l.Bonus = p.Price.Was, p.BonusMechanism
	}
	return l
}

func linesOf(ps []appie.Product) []productLine {
	out := make([]productLine, 0, len(ps))
	for _, p := range ps {
		out = append(out, lineOf(p))
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ah-mcp-be:", err)
	os.Exit(1)
}

func registerTools(s *server.MCPServer) {
	str := func(n, d string, req bool) mcp.ToolOption {
		if req {
			return mcp.WithString(n, mcp.Required(), mcp.Description(d))
		}
		return mcp.WithString(n, mcp.Description(d))
	}
	num := func(n, d string, req bool) mcp.ToolOption {
		if req {
			return mcp.WithNumber(n, mcp.Required(), mcp.Description(d))
		}
		return mcp.WithNumber(n, mcp.Description(d))
	}

	// --- auth ---
	add(s, false, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		anonymous = false
		client.Logout() // drop any anonymous token so Login starts clean
		if err := client.Login(ctx); err != nil {
			return nil, err
		}
		return client.GetMember(ctx)
	}, "ah_login", "Log in to ah.be. Opens a browser window on this machine and blocks until login completes (max 5 min).")

	add(s, false, func(context.Context, mcp.CallToolRequest) (any, error) {
		client.Logout()
		anonymous = false
		if err := os.Remove(tokensPath); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return "logged out", nil
	}, "ah_logout", "Forget stored tokens.")

	add(s, true, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		return client.GetMember(ctx)
	}, "ah_get_member", "Member profile and bonus card.")

	// --- products (work anonymously) ---
	add(s, false, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		return client.SearchProductsFiltered(ctx, appie.SearchOptions{
			Query: r.GetString("query", ""),
			Limit: r.GetInt("limit", 10),
			Bonus: r.GetBool("bonus_only", false),
		})
	}, "ah_search_products", "Search ah.be products (Dutch terms work best). Returns id, title, price, bonus info.",
		str("query", "Search term, e.g. 'melk'", true),
		num("limit", "Max results (default 10)", false),
		mcp.WithBoolean("bonus_only", mcp.Description("Only products currently in bonus")))

	add(s, false, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		id, err := r.RequireInt("product_id")
		if err != nil {
			return nil, err
		}
		if r.GetBool("nutrition", false) {
			return client.GetProductFull(ctx, id)
		}
		return client.GetProduct(ctx, id)
	}, "ah_get_product", "Details for one product.",
		num("product_id", "Product id from ah_search_products", true),
		mcp.WithBoolean("nutrition", mcp.Description("Include nutritional info")))

	add(s, false, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		return client.GetBonusProducts(ctx)
	}, "ah_get_bonus", "Current bonus offers. Group entries carry a bonusSegmentId; expand with ah_get_bonus_group.")

	add(s, false, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		return client.GetBonusGroupProducts(ctx, r.GetString("segment_id", ""))
	}, "ah_get_bonus_group", "Products in a bonus group.",
		str("segment_id", "bonusSegmentId from ah_get_bonus", true))

	add(s, false, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		ids := r.GetIntSlice("product_ids", nil)
		if len(ids) == 0 || len(ids) > 20 {
			return nil, fmt.Errorf("give 1 to 20 product_ids")
		}
		out := make([]productLine, 0, len(ids))
		for _, id := range ids { // ponytail: sequential, parallelise if 20 products ever feels slow
			var p *appie.Product
			var err error
			if r.GetBool("nutrition", false) {
				p, err = client.GetProductFull(ctx, id)
			} else {
				p, err = client.GetProduct(ctx, id)
			}
			if err != nil {
				out = append(out, productLine{ID: id, Error: err.Error()})
				continue
			}
			l := lineOf(*p)
			l.ID = id
			out = append(out, l)
		}
		return out, nil
	}, "ah_get_products", "Details for up to 20 products at once. With nutrition=true each product includes its nutritional values per 100 g, handy for calculating calories and protein of a recipe.",
		mcp.WithArray("product_ids", mcp.Required(), mcp.Description("Product ids"), mcp.WithNumberItems()),
		mcp.WithBoolean("nutrition", mcp.Description("Include nutritional info")))

	add(s, false, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		periods, err := client.GetBonusPeriods(ctx)
		out := make([]map[string]string, 0, len(periods))
		for _, p := range periods {
			out = append(out, map[string]string{"start_date": p.StartDate, "end_date": p.EndDate})
		}
		return out, err
	}, "ah_get_bonus_periods", "Bonus weeks: the current one and, a few days ahead, next week (start_date and end_date). Use a start date with ah_get_personal_bonus.")

	add(s, true, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		ps, err := client.GetPersonalBonus(ctx, r.GetString("start_date", ""))
		if err != nil {
			return nil, err
		}
		return linesOf(ps), nil
	}, "ah_get_personal_bonus", "Your personal bonus offers (Bonus Box) for the current week, or for the week starting at start_date.",
		str("start_date", "Week start from ah_get_bonus_periods, e.g. 2026-10-12. Empty means the current week.", false))

	// --- cart: on ah.be the "winkelmandje" (/mijnlijst) is the v2 shopping list ---
	add(s, true, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		items, err := getCart(ctx)
		if err != nil {
			return nil, err
		}
		type line struct {
			ProductID int     `json:"product_id,omitempty"`
			Title     string  `json:"title"`
			Quantity  int     `json:"quantity"`
			Size      string  `json:"size,omitempty"`
			Price     float64 `json:"price,omitempty"`
			Was       float64 `json:"price_before_bonus,omitempty"`
			Bonus     string  `json:"bonus,omitempty"`
			Category  string  `json:"category,omitempty"`
		}
		lines, qty, total := []line{}, 0, 0.0
		for _, i := range items {
			p := i.Product
			if p.Price == 0 { // non-bonus items only carry priceBeforeBonus
				p.Price = p.Was
			}
			lines = append(lines, line{p.ID, i.title(), i.Quantity, p.Size, p.Price, p.Was, p.Bonus, p.Category})
			qty += i.Quantity
			total += p.Price * float64(i.Quantity)
		}
		// ponytail: price sum ignores stacked/multi-buy bonus rules, so it's an estimate
		return map[string]any{"items": lines, "total_quantity": qty, "estimated_total": total}, nil
	}, "ah_get_cart", "Your ah.be winkelmandje (ah.be/mijnlijst): products, quantities, prices and bonus.")

	add(s, true, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		raw, _ := r.GetArguments()["items"].([]any)
		if len(raw) == 0 {
			return nil, fmt.Errorf("items is empty")
		}
		items, seen := make([]map[string]any, 0, len(raw)), map[any]bool{}
		for _, x := range raw {
			spec, _ := x.(map[string]any)
			it, err := cartItem(ctx, spec) // build everything first so a bad item sends nothing
			if err != nil {
				return nil, err
			}
			if id, ok := it["productId"]; ok {
				if seen[id] {
					return nil, fmt.Errorf("product %v appears twice", id)
				}
				seen[id] = true
			}
			items = append(items, it)
		}
		return fmt.Sprintf("ok, %d items", len(items)), patchCart(ctx, items)
	}, "ah_set_cart_items", "Add products (product_id) or free-text items (name) to the winkelmandje, or change quantities, in one call. Quantity 0 removes an item.",
		mcp.WithArray("items", mcp.Required(), mcp.Description("Items to add or change"), mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"product_id": map[string]any{"type": "number", "description": "Product id from ah_search_products"},
				"name":       map[string]any{"type": "string", "description": "Free-text item when there is no product_id"},
				"quantity":   map[string]any{"type": "number", "description": "Quantity, 0 removes (default 1)"},
			},
		})))

	add(s, true, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		if r.GetString("confirm", "") != "yes" {
			return nil, fmt.Errorf("pass confirm=yes to empty the winkelmandje")
		}
		items, err := getCart(ctx)
		if err != nil || len(items) == 0 {
			return "already empty", err
		}
		zeros := make([]map[string]any, 0, len(items))
		for _, i := range items {
			z := map[string]any{"quantity": 0, "type": i.Type, "originCode": i.OriginCode, "strikeThrough": false}
			if i.Product.ID > 0 {
				z["productId"] = i.Product.ID
			}
			z["description"] = i.title() // PATCH rejects items without a description
			zeros = append(zeros, z)
		}
		return "ok", patchCart(ctx, zeros)
	}, "ah_clear_cart", "Remove everything from the winkelmandje.",
		str("confirm", "Must be 'yes'", true))

	// --- orders & receipts ---
	add(s, true, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		return client.GetFulfillments(ctx)
	}, "ah_get_orders", "Scheduled deliveries / pickups.")

	add(s, true, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		id, err := r.RequireInt("order_id")
		if err != nil {
			return nil, err
		}
		return client.GetOrderDetails(ctx, id)
	}, "ah_get_order_details", "Items of a specific order.",
		num("order_id", "Order id from ah_get_orders", true))

	add(s, true, func(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
		return client.GetReceipts(ctx)
	}, "ah_get_receipts", "Recent in-store receipts.")

	add(s, true, func(ctx context.Context, r mcp.CallToolRequest) (any, error) {
		return client.GetReceipt(ctx, r.GetString("receipt_id", ""))
	}, "ah_get_receipt", "Full detail of one receipt.",
		str("receipt_id", "Receipt id from ah_get_receipts", true))
}
