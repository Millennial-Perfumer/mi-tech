package mcp

import (
	"strings"
	"testing"
)

// TestCatalogToolNamesUnique verifies tool names are unique.
func TestCatalogToolNamesUnique(t *testing.T) {
	seen := make(map[string]struct{}, len(DefaultCatalog))
	for _, spec := range DefaultCatalog {
		if spec.Name == "" {
			t.Fatalf("tool with empty name found at index %d", 0)
		}
		if _, dup := seen[spec.Name]; dup {
			t.Fatalf("duplicate tool name: %s", spec.Name)
		}
		seen[spec.Name] = struct{}{}
	}
}

// TestCatalogScopes verifies every tool carries a valid scope and has a
// non-empty description. Write-capable tools must use a write or destructive
// scope.
func TestCatalogScopesReadOnly(t *testing.T) {
	for _, spec := range DefaultCatalog {
		isWriteScope := strings.HasSuffix(spec.Scope, ":write") || strings.HasSuffix(spec.Scope, ":destructive") || spec.Scope == ScopeMarketingPublish
		if !strings.HasSuffix(spec.Scope, ":read") && !isWriteScope {
			t.Errorf("tool %s has non-read-only scope: %s", spec.Name, spec.Scope)
		}
		if isWriteScope {
			if !spec.Write {
				t.Errorf("tool %s has write scope but is not write-capable", spec.Name)
			}
		} else if spec.Write {
			t.Errorf("tool %s unexpectedly has write capability", spec.Name)
		}
		if spec.Scope == "" {
			t.Errorf("tool %s has empty scope", spec.Name)
		}
		if spec.Description == "" {
			t.Errorf("tool %s has empty description", spec.Name)
		}
	}
}

// TestCatalogArgSpecsValid verifies argument names are unique per tool and types are valid.
func TestCatalogArgSpecsValid(t *testing.T) {
	validTypes := map[ArgType]bool{ArgString: true, ArgInt: true, ArgNumber: true, ArgObject: true, ArgBoolean: true}
	for _, spec := range DefaultCatalog {
		seen := make(map[string]struct{}, len(spec.Args))
		for _, a := range spec.Args {
			if a.Name == "" {
				t.Errorf("tool %s has an arg with empty name", spec.Name)
			}
			if _, dup := seen[a.Name]; dup {
				t.Errorf("tool %s has duplicate arg name: %s", spec.Name, a.Name)
			}
			seen[a.Name] = struct{}{}
			if !validTypes[a.Type] {
				t.Errorf("tool %s arg %s has invalid type: %s", spec.Name, a.Name, a.Type)
			}
		}
		for _, name := range spec.PathArgs {
			if _, ok := seen[name]; !ok {
				t.Errorf("tool %s path arg %s is not declared in Args", spec.Name, name)
			}
		}
		for _, name := range spec.QueryArgs {
			if _, ok := seen[name]; !ok {
				t.Errorf("tool %s query arg %s is not declared in Args", spec.Name, name)
			}
		}
	}
}

// TestRouteMapComplete verifies every catalog tool maps to exactly one route
// with the same path and method declared by the catalog.
func TestRouteMapComplete(t *testing.T) {
	seen := make(map[string]struct{}, len(DefaultCatalog))
	for _, spec := range DefaultCatalog {
		b, ok := RouteFor(spec.Name)
		if !ok {
			t.Errorf("catalog tool %s has no route binding", spec.Name)
			continue
		}
		wantMethod := "GET"
		if spec.Write {
			wantMethod = spec.Method
			if wantMethod == "" {
				wantMethod = "POST"
			}
		}
		if b.Method != wantMethod {
			t.Errorf("catalog tool %s maps to %s method: %s", spec.Name, wantMethod, b.Method)
		}
		if b.Path == "" {
			t.Errorf("catalog tool %s maps to empty path", spec.Name)
		}
		if b.Path != spec.Route {
			t.Errorf("catalog tool %s maps to path %q, want %q", spec.Name, b.Path, spec.Route)
		}
		if _, dup := seen[spec.Name]; dup {
			t.Errorf("catalog tool %s appears more than once", spec.Name)
		}
		seen[spec.Name] = struct{}{}
	}

	// Ensure no extra route bindings exist outside the catalog.
	for name := range routeMap {
		if _, ok := DefaultCatalog.Lookup(name); !ok {
			t.Errorf("route binding %s has no catalog tool", name)
		}
	}
}

func TestWriteToolContracts(t *testing.T) {
	ordersUpdate, ok := DefaultCatalog.Lookup("orders_update")
	if !ok {
		t.Fatal("orders_update missing from catalog")
	}
	if len(ordersUpdate.Args) != 2 || ordersUpdate.Args[0].Name != "payload" || ordersUpdate.Args[1].Name != "id" {
		t.Fatalf("orders_update args = %#v, want payload and id", ordersUpdate.Args)
	}
	if len(ordersUpdate.QueryArgs) != 1 || ordersUpdate.QueryArgs[0] != "id" {
		t.Fatalf("orders_update query args = %#v, want [id]", ordersUpdate.QueryArgs)
	}
	if markDelivered, ok := DefaultCatalog.Lookup("orders_mark_delivered"); !ok {
		t.Fatal("orders_mark_delivered missing from catalog")
	} else if len(markDelivered.Args) != 1 || markDelivered.Args[0].Name != "id" {
		t.Errorf("orders_mark_delivered args = %#v, want only id", markDelivered.Args)
	}

	for _, name := range []string{"planner_sprint_update", "planner_sprint_delete"} {
		spec, ok := DefaultCatalog.Lookup(name)
		if !ok {
			t.Fatalf("%s missing from catalog", name)
		}
		if len(spec.QueryArgs) != 1 || spec.QueryArgs[0] != "id" {
			t.Errorf("%s query args = %#v, want [id]", name, spec.QueryArgs)
		}
	}

	for _, name := range []string{"whatsapp_template_sync_single", "whatsapp_template_fetch"} {
		spec, ok := DefaultCatalog.Lookup(name)
		if !ok {
			t.Fatalf("%s missing from catalog", name)
		}
		if len(spec.Args) != 1 || spec.Args[0].Name != "name" || !spec.Args[0].Required {
			t.Errorf("%s args = %#v, want one required name arg", name, spec.Args)
		}
		if len(spec.QueryArgs) != 1 || spec.QueryArgs[0] != "name" {
			t.Errorf("%s query args = %#v, want [name]", name, spec.QueryArgs)
		}
	}

	if _, ok := DefaultCatalog.Lookup("settings_update"); ok {
		t.Error("generic settings_update must not be exposed through MCP")
	}
	if _, ok := RouteFor("settings_update"); ok {
		t.Error("generic settings_update must not have a route binding")
	}

	for _, name := range []string{"inventory_clear", "shopify_reset_orders", "customers_delete", "ai_conversation_delete"} {
		spec, ok := DefaultCatalog.Lookup(name)
		if !ok {
			t.Fatalf("%s missing from catalog", name)
		}
		if !strings.HasSuffix(spec.Scope, ":destructive") {
			t.Errorf("%s scope = %q, want a destructive scope", name, spec.Scope)
		}
	}
}

// TestNoWriteRoutes verifies the read-only route surface never includes
// known mutating endpoints.
func TestNoWriteRoutes(t *testing.T) {
	writeSuffixes := []string{
		"/sync", "/reset", "/issue", "/cancel", "/accept", "/reject",
		"/convert", "/move", "/reveal", "/upload", "/send", "/post",
	}
	for _, path := range ReadOnlyPaths() {
		for _, suffix := range writeSuffixes {
			if strings.HasSuffix(path, suffix) {
				t.Errorf("read-only route surface includes mutating path: %s", path)
			}
		}
		if strings.Contains(path, "/bulk") {
			t.Errorf("read-only route surface includes bulk path: %s", path)
		}
	}
}

// TestReadOnlyPathsDistinct verifies the read-only path set is deduplicated.
func TestReadOnlyPathsDistinct(t *testing.T) {
	paths := ReadOnlyPaths()
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if _, dup := seen[p]; dup {
			t.Errorf("duplicate read-only path: %s", p)
		}
		seen[p] = struct{}{}
	}
}

// TestLookup verifies Catalog.Lookup behaves correctly.
func TestLookup(t *testing.T) {
	spec, ok := DefaultCatalog.Lookup("orders_list")
	if !ok {
		t.Fatal("expected orders_list in catalog")
	}
	if spec.Scope != ScopeOrders {
		t.Errorf("orders_list scope = %s, want %s", spec.Scope, ScopeOrders)
	}

	if _, ok := DefaultCatalog.Lookup("not_a_real_tool"); ok {
		t.Error("expected unknown tool lookup to fail")
	}
}

// TestScopes verifies the distinct scope list is derived correctly.
func TestScopes(t *testing.T) {
	scopes := DefaultCatalog.Scopes()
	if len(scopes) == 0 {
		t.Fatal("expected non-empty scope list")
	}
	seen := make(map[string]bool)
	for _, s := range scopes {
		if seen[s] {
			t.Errorf("duplicate scope in Scopes(): %s", s)
		}
		seen[s] = true
		if !strings.HasSuffix(s, ":read") && s != ScopeMarketingPublish && !strings.HasSuffix(s, ":write") && !strings.HasSuffix(s, ":destructive") {
			t.Errorf("non-read-only scope derived: %s", s)
		}
	}
}
