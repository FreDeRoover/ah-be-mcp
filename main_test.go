package main

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	appie "github.com/gwillem/appie-go"
	"github.com/mark3labs/mcp-go/server"
)

// Real-shaped v2 list: a bonus product, a non-bonus product (price only in
// priceBeforeBonus), and a free-text item without productDetails.
const listFixture = `{"id":"x","items":[
 {"quantity":2,"type":"SHOPPABLE","originCode":"PRD","productDetails":{"product":{
   "webshopId":235930,"title":"AH Gehakt","salesUnitSize":"1 kg","mainCategory":"Vlees",
   "currentPrice":3.74,"priceBeforeBonus":7.49,"bonusMechanism":"50% korting"}}},
 {"quantity":1,"type":"SHOPPABLE","originCode":"PRD","productDetails":{"product":{
   "webshopId":425572,"title":"Melk 12-pack","priceBeforeBonus":19.08}}},
 {"quantity":1,"type":"SHOPPABLE","originCode":"TXT","description":"bier"}]}`

// fakeAH stands in for api.ah.be.
type fakeAH struct {
	mu      sync.Mutex
	list    string
	patches [][]map[string]any
	anon    int
	hits    int

	personalQuery string
}

func (f *fakeAH) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hits++
	switch {
	case r.URL.Path == "/mobile-auth/v1/auth/token/anonymous":
		f.anon++
		io.WriteString(w, `{"access_token":"anon","refresh_token":"r"}`)
	case r.URL.Path == "/mobile-services/shoppinglist/v2/items" && r.Method == "GET":
		io.WriteString(w, f.list)
	case r.URL.Path == "/mobile-services/shoppinglist/v2/items" && r.Method == "PATCH":
		var body struct {
			Items []map[string]any `json:"items"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		f.patches = append(f.patches, body.Items)
		io.WriteString(w, `{}`)
	case r.URL.Path == "/mobile-services/product/detail/v4/fir/999":
		http.NotFound(w, r)
	case strings.HasPrefix(r.URL.Path, "/mobile-services/product/detail/v4/fir/"):
		io.WriteString(w, `{"productCard":{"webshopId":54074,"title":"AH Komkommer"}}`)
	case r.URL.Path == "/graphql":
		io.WriteString(w, `{"data":{"product":{"id":54074,"tradeItem":{"nutritions":[{"nutrients":[{"type":"PROTEIN","name":"Eiwitten","value":"0.6 g"}]}]}}}}`)
	case r.URL.Path == "/mobile-services/bonuspage/v3/metadata":
		io.WriteString(w, `{"periods":[{"bonusStartDate":"2026-10-05","bonusEndDate":"2026-10-11"},{"bonusStartDate":"2026-10-12","bonusEndDate":"2026-10-18"}]}`)
	case r.URL.Path == "/mobile-services/bonuspage/v1/personal":
		f.personalQuery = r.URL.RawQuery
		io.WriteString(w, `{"bonusGroupOrProducts":[{"bonusGroup":{"id":"338409","segmentDescription":"Kattenvoer","discountDescription":"25% KORTING"}}]}`)
	case r.URL.Path == "/mobile-services/product/search/v2":
		io.WriteString(w, `{"products":[{"webshopId":54074,"title":"AH Komkommer","currentPrice":0.99,"priceBeforeBonus":1.29,"isBonus":true,"bonusMechanism":"25% KORTING","unitPriceDescription":"per stuk","images":[{"url":"https://x/img.png","width":800,"height":800}]}],"page":{"totalElements":1}}`)
	default:
		http.NotFound(w, r)
	}
}

// setup points the global client at a fake API and returns an MCP server with
// all tools registered. loggedIn=false leaves the client without tokens.
func setup(t *testing.T, loggedIn bool) (*server.MCPServer, *fakeAH) {
	t.Helper()
	f := &fakeAH{list: listFixture}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	opts := []appie.Option{appie.WithBaseURL(srv.URL)}
	if loggedIn {
		opts = append(opts, appie.WithTokens("access", "refresh"))
	}
	client, anonymous = appie.New(opts...), false
	s := server.NewMCPServer("test", "0")
	registerTools(s)
	return s, f
}

// call invokes a tool through the real MCP JSON-RPC path.
func call(t *testing.T, s *server.MCPServer, name string, args map[string]any) (text string, isErr bool) {
	t.Helper()
	req, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "tools/call",
		"params": map[string]any{"name": name, "arguments": args},
	})
	out, _ := json.Marshal(s.HandleMessage(context.Background(), req))
	var resp struct {
		Result struct {
			Content []struct{ Text string } `json:"content"`
			IsError bool                    `json:"isError"`
		} `json:"result"`
		Error *struct{ Message string } `json:"error"`
	}
	if err := json.Unmarshal(out, &resp); err != nil || resp.Error != nil || len(resp.Result.Content) == 0 {
		t.Fatalf("bad response for %s: %s", name, out)
	}
	return resp.Result.Content[0].Text, resp.Result.IsError
}

func TestToolsMatchManifest(t *testing.T) {
	s, _ := setup(t, true)
	out, _ := json.Marshal(s.HandleMessage(context.Background(),
		json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)))
	var list struct {
		Result struct{ Tools []struct{ Name string } } `json:"result"`
	}
	json.Unmarshal(out, &list)
	var got []string
	for _, tl := range list.Result.Tools {
		got = append(got, tl.Name)
	}

	raw, err := os.ReadFile("mcpb/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m struct{ Tools []struct{ Name string } }
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, tl := range m.Tools {
		want = append(want, tl.Name)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("registered tools and mcpb/manifest.json differ:\n got  %v\n want %v", got, want)
	}
}

func TestAuthToolsRequireLogin(t *testing.T) {
	s, f := setup(t, false)
	text, isErr := call(t, s, "ah_get_cart", nil)
	if !isErr || !strings.Contains(text, "ah_login") {
		t.Fatalf("want not-logged-in error, got %q (isErr=%v)", text, isErr)
	}
	if f.hits != 0 {
		t.Fatalf("API was hit %d times before login", f.hits)
	}
}

func TestBrowsingFetchesAnonymousTokenAndThenBlocksAuthTools(t *testing.T) {
	s, f := setup(t, false)
	text, isErr := call(t, s, "ah_search_products", map[string]any{"query": "komkommer"})
	if isErr || !strings.Contains(text, "AH Komkommer") || f.anon != 1 {
		t.Fatalf("anonymous search failed: %q isErr=%v anon=%d", text, isErr, f.anon)
	}
	// an anonymous token must not count as logged in
	if text, isErr := call(t, s, "ah_get_cart", nil); !isErr || !strings.Contains(text, "ah_login") {
		t.Fatalf("anonymous session reached cart: %q", text)
	}
}

func TestGetCart(t *testing.T) {
	s, _ := setup(t, true)
	text, isErr := call(t, s, "ah_get_cart", nil)
	if isErr {
		t.Fatal(text)
	}
	var cart struct {
		Items []struct {
			ProductID int     `json:"product_id"`
			Title     string  `json:"title"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
			Bonus     string  `json:"bonus"`
		}
		TotalQuantity  int     `json:"total_quantity"`
		EstimatedTotal float64 `json:"estimated_total"`
	}
	if err := json.Unmarshal([]byte(text), &cart); err != nil {
		t.Fatal(err)
	}
	if len(cart.Items) != 3 || cart.TotalQuantity != 4 {
		t.Fatalf("items=%d qty=%d", len(cart.Items), cart.TotalQuantity)
	}
	if g := cart.Items[0]; g.ProductID != 235930 || g.Price != 3.74 || g.Bonus != "50% korting" {
		t.Errorf("bonus item wrong: %+v", g)
	}
	if m := cart.Items[1]; m.Price != 19.08 { // falls back to priceBeforeBonus
		t.Errorf("non-bonus price fallback wrong: %+v", m)
	}
	if b := cart.Items[2]; b.Title != "bier" || b.ProductID != 0 {
		t.Errorf("free-text item wrong: %+v", b)
	}
	if want := 2*3.74 + 19.08; math.Abs(cart.EstimatedTotal-want) > 0.001 {
		t.Errorf("total %.2f, want %.2f", cart.EstimatedTotal, want)
	}
}

func TestSetCartItems(t *testing.T) {
	s, f := setup(t, true)

	// one call, one PATCH: product (title looked up), free text, and quantity 0
	text, isErr := call(t, s, "ah_set_cart_items", map[string]any{"items": []any{
		map[string]any{"product_id": 54074, "quantity": 3},
		map[string]any{"name": "wc-papier", "quantity": 2},
		map[string]any{"product_id": 4164, "quantity": 0},
	}})
	if isErr || len(f.patches) != 1 || len(f.patches[0]) != 3 {
		t.Fatalf("want one PATCH with 3 items: %q isErr=%v patches=%v", text, isErr, f.patches)
	}
	p, txt, zero := f.patches[0][0], f.patches[0][1], f.patches[0][2]
	if p["productId"] != float64(54074) || p["quantity"] != float64(3) || p["description"] != "AH Komkommer" ||
		p["originCode"] != "PRD" || p["strikeThrough"] != false {
		t.Errorf("bad product item %v", p)
	}
	if txt["originCode"] != "TXT" || txt["description"] != "wc-papier" || txt["quantity"] != float64(2) {
		t.Errorf("bad free-text item %v", txt)
	}
	if zero["quantity"] != float64(0) {
		t.Errorf("quantity 0 not passed through: %v", zero)
	}

	// invalid batches send nothing, even when only one item is bad
	for name, items := range map[string][]any{
		"empty":     {},
		"no id":     {map[string]any{"product_id": 54074}, map[string]any{"quantity": 1}},
		"duplicate": {map[string]any{"product_id": 54074}, map[string]any{"product_id": 54074}},
		"unknown":   {map[string]any{"product_id": 54074}, map[string]any{"product_id": 999}},
	} {
		n := len(f.patches)
		if _, isErr := call(t, s, "ah_set_cart_items", map[string]any{"items": items}); !isErr || len(f.patches) != n {
			t.Errorf("%s: expected an error and no PATCH", name)
		}
	}
}

func TestGetProducts(t *testing.T) {
	s, _ := setup(t, false)
	text, isErr := call(t, s, "ah_get_products", map[string]any{"product_ids": []any{54074, 999}, "nutrition": true})
	if isErr {
		t.Fatal(text)
	}
	var out []struct {
		ID        int                            `json:"product_id"`
		Title     string                         `json:"title"`
		Error     string                         `json:"error"`
		Nutrition []struct{ Name, Value string } `json:"nutrition_per_100g"`
	}
	if err := json.Unmarshal([]byte(text), &out); err != nil || len(out) != 2 {
		t.Fatalf("bad output %q: %v", text, err)
	}
	if out[0].ID != 54074 || out[0].Title != "AH Komkommer" || len(out[0].Nutrition) != 1 || out[0].Nutrition[0].Value != "0.6 g" {
		t.Errorf("first product wrong: %+v", out[0])
	}
	if out[1].ID != 999 || out[1].Error == "" { // one bad id must not sink the batch
		t.Errorf("second product should carry an error: %+v", out[1])
	}
	if _, isErr := call(t, s, "ah_get_products", map[string]any{"product_ids": []any{}}); !isErr {
		t.Error("empty product_ids should error")
	}
}

func TestBonusPeriodsAndPersonalBonus(t *testing.T) {
	s, f := setup(t, false)
	text, isErr := call(t, s, "ah_get_bonus_periods", nil)
	if isErr || !strings.Contains(text, `"start_date":"2026-10-12"`) {
		t.Fatalf("periods: %q isErr=%v", text, isErr)
	}
	// personal bonus is member-specific: anonymous sessions are refused
	if text, isErr := call(t, s, "ah_get_personal_bonus", nil); !isErr || !strings.Contains(text, "ah_login") {
		t.Fatalf("anonymous personal bonus: %q", text)
	}

	s, f = setup(t, true)
	text, isErr = call(t, s, "ah_get_personal_bonus", map[string]any{"start_date": "2026-10-12"})
	if isErr || f.personalQuery != "bonusStartDate=2026-10-12" ||
		!strings.Contains(text, `"bonus_segment_id":"338409"`) || !strings.Contains(text, "25% KORTING") {
		t.Fatalf("personal bonus: %q query=%q isErr=%v", text, f.personalQuery, isErr)
	}
}

func TestClearCart(t *testing.T) {
	s, f := setup(t, true)

	if _, isErr := call(t, s, "ah_clear_cart", map[string]any{"confirm": "no"}); !isErr || len(f.patches) != 0 {
		t.Fatal("clear without confirm=yes must error and not PATCH")
	}

	if text, isErr := call(t, s, "ah_clear_cart", map[string]any{"confirm": "yes"}); isErr {
		t.Fatal(text)
	}
	if len(f.patches) != 1 || len(f.patches[0]) != 3 {
		t.Fatalf("want one PATCH with 3 items, got %v", f.patches)
	}
	for _, it := range f.patches[0] {
		if it["quantity"] != float64(0) || it["description"] == "" || it["strikeThrough"] != false {
			t.Errorf("bad removal item %v", it)
		}
	}

	// already empty: no PATCH
	f.list, f.patches = `{"items":[]}`, nil
	if text, isErr := call(t, s, "ah_clear_cart", map[string]any{"confirm": "yes"}); isErr || !strings.Contains(text, "empty") || len(f.patches) != 0 {
		t.Fatalf("empty cart: %q isErr=%v patches=%v", text, isErr, f.patches)
	}
}

func TestAPIErrorBecomesToolError(t *testing.T) {
	s, _ := setup(t, true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"boom"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()
	client = appie.New(appie.WithBaseURL(srv.URL), appie.WithTokens("a", "r"))
	if text, isErr := call(t, s, "ah_get_cart", nil); !isErr || text == "" {
		t.Fatalf("want tool error, got %q isErr=%v", text, isErr)
	}
}

func TestSearchOutputIsCompact(t *testing.T) {
	s, _ := setup(t, false)
	text, isErr := call(t, s, "ah_search_products", map[string]any{"query": "komkommer"})
	if isErr {
		t.Fatal(text)
	}
	var out []map[string]any
	if err := json.Unmarshal([]byte(text), &out); err != nil || len(out) != 1 {
		t.Fatalf("bad output %q: %v", text, err)
	}
	p := out[0]
	if p["product_id"] != float64(54074) || p["price"] != 0.99 || p["price_before_bonus"] != 1.29 ||
		p["bonus"] != "25% KORTING" || p["unit_price"] != "per stuk" {
		t.Errorf("wrong fields: %v", p)
	}
	if strings.Contains(text, "img.png") || strings.Contains(text, "images") {
		t.Errorf("image data leaked into search output: %s", text)
	}
}

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"0.2.1", "0.2.0", true},
		{"0.10.0", "0.9.0", true}, // numeric, not string, comparison
		{"1.0.0", "0.99.99", true},
		{"0.2.0", "0.2.0", false},
		{"0.2.0", "0.2.1", false},
		{"0.2.1", "0.2", true},
		{"abc", "0.2.0", false},
		{"0.3.0-rc1", "0.2.0", false}, // pre-release tags are never offered
	} {
		if got := newer(c.a, c.b); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestCheckUpdate(t *testing.T) {
	status := http.StatusOK
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		io.WriteString(w, `{"tag_name":"v0.3.0","html_url":"https://example.test/release"}`)
	}))
	defer gh.Close()
	oldURL, oldVersion := releaseURL, version
	defer func() { releaseURL, version = oldURL, oldVersion }()
	releaseURL = gh.URL

	s, f := setup(t, false)
	for _, c := range []struct {
		current string
		update  bool
		msg     string
	}{
		{"0.2.1", true, "newer version"},
		{"0.3.0", false, "up to date"},
		{"0.4.0", false, "up to date"},
		{"dev", false, "development build"},
	} {
		version = c.current
		text, isErr := call(t, s, "ah_check_update", nil)
		var out struct {
			Current string `json:"current_version"`
			Latest  string `json:"latest_version"`
			Update  bool   `json:"update_available"`
			URL     string `json:"release_url"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(text), &out); isErr || err != nil {
			t.Fatalf("%s: %q isErr=%v err=%v", c.current, text, isErr, err)
		}
		if out.Update != c.update || out.Latest != "0.3.0" || out.Current != c.current ||
			out.URL != "https://example.test/release" || !strings.Contains(out.Message, c.msg) {
			t.Errorf("current %s: %+v", c.current, out)
		}
	}
	if f.hits != 0 {
		t.Errorf("update check must not call AH (hits=%d)", f.hits)
	}

	status = http.StatusInternalServerError
	if text, isErr := call(t, s, "ah_check_update", nil); !isErr || !strings.Contains(text, "GitHub") {
		t.Errorf("want a GitHub error, got %q isErr=%v", text, isErr)
	}
}
