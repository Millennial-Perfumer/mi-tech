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
// non-empty description. The explicitly scoped social queue publisher is the
// only write-capable tool.
func TestCatalogScopesReadOnly(t *testing.T) {
	for _, spec := range DefaultCatalog {
		if !strings.HasSuffix(spec.Scope, ":read") && spec.Scope != ScopeMarketingPublish {
			t.Errorf("tool %s has non-read-only scope: %s", spec.Name, spec.Scope)
		}
		if spec.Write != (spec.Name == "smm_queue_create") {
			t.Errorf("tool %s has unexpected write capability: %v", spec.Name, spec.Write)
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
	validTypes := map[ArgType]bool{ArgString: true, ArgInt: true, ArgNumber: true}
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
	}
}

// TestRouteMapComplete verifies every catalog tool maps to exactly one GET route.
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
			wantMethod = "POST"
		}
		if b.Method != wantMethod {
			t.Errorf("catalog tool %s maps to %s method: %s", spec.Name, wantMethod, b.Method)
		}
		if b.Path == "" {
			t.Errorf("catalog tool %s maps to empty path", spec.Name)
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
		if !strings.HasSuffix(s, ":read") && s != ScopeMarketingPublish {
			t.Errorf("non-read-only scope derived: %s", s)
		}
	}
}
