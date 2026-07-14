package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestOpenAPIParity ensures every /api/v1 route mounted in Mount() is documented
// in docs/api/openapi.yaml and vice versa (path templates only; methods optional detail).
func TestOpenAPIParity(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	routerSrc, err := os.ReadFile(filepath.Join(root, "internal/transport/http/api/router.go"))
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	openapiSrc, err := os.ReadFile(filepath.Join(root, "docs/api/openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}

	routerPaths := extractRouterAPIPaths(string(routerSrc))
	openapiPaths := extractOpenAPIPaths(string(openapiSrc))

	var missing []string
	for p := range routerPaths {
		if _, ok := openapiPaths[p]; !ok {
			missing = append(missing, p)
		}
	}
	var extra []string
	for p := range openapiPaths {
		if _, ok := routerPaths[p]; !ok {
			extra = append(extra, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("OpenAPI ↔ router path mismatch:\n  in router, missing from OpenAPI: %v\n  in OpenAPI, missing from router: %v", missing, extra)
	}

	// Stronger check: HTTP methods per path (same sources, still regex-only).
	routerOps := extractRouterOperations(string(routerSrc))
	openapiOps := extractOpenAPIOperations(string(openapiSrc))
	var missingOps, extraOps []string
	for op := range routerOps {
		if !openapiOps[op] {
			missingOps = append(missingOps, op)
		}
	}
	for op := range openapiOps {
		if !routerOps[op] {
			extraOps = append(extraOps, op)
		}
	}
	sort.Strings(missingOps)
	sort.Strings(extraOps)
	if len(missingOps) > 0 || len(extraOps) > 0 {
		t.Fatalf("OpenAPI ↔ router operation mismatch:\n  in router, missing from OpenAPI: %v\n  in OpenAPI, missing from router: %v", missingOps, extraOps)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

var (
	mountFuncRE     = regexp.MustCompile(`(?s)func \(rt \*Router\) Mount\(r chi\.Router\) \{.*?\n\}`)
	routeCallRE     = regexp.MustCompile(`\br\.(Get|Post|Put|Patch|Delete)\("([^"]+)"`)
	apiRouteRE      = regexp.MustCompile(`\br\.Route\("(/api/v1)"`)
	openapiPathRE   = regexp.MustCompile(`(?m)^  (/api/v1/[^:\s]+):\s*$`)
	openapiMethodRE = regexp.MustCompile(`(?m)^    (get|post|put|patch|delete):\s*$`)
)

func extractMountBody(src string) string {
	m := mountFuncRE.FindString(src)
	return m
}

func extractRouterAPIPaths(src string) map[string]struct{} {
	ops := extractRouterOperations(src)
	out := make(map[string]struct{}, len(ops))
	for op := range ops {
		// "GET /api/v1/tasks" → "/api/v1/tasks"
		parts := strings.SplitN(op, " ", 2)
		if len(parts) != 2 {
			continue
		}
		out[parts[1]] = struct{}{}
	}
	return out
}

func extractRouterOperations(src string) map[string]bool {
	body := extractMountBody(src)
	if body == "" {
		return nil
	}
	prefix := "/api/v1"
	if m := apiRouteRE.FindStringSubmatch(body); len(m) == 2 {
		prefix = m[1]
	}
	out := make(map[string]bool)
	for _, m := range routeCallRE.FindAllStringSubmatch(body, -1) {
		method := strings.ToUpper(m[1])
		path := m[2]
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		out[fmt.Sprintf("%s %s%s", method, prefix, path)] = true
	}
	return out
}

func extractOpenAPIPaths(src string) map[string]struct{} {
	ops := extractOpenAPIOperations(src)
	out := make(map[string]struct{}, len(ops))
	for op := range ops {
		parts := strings.SplitN(op, " ", 2)
		if len(parts) != 2 {
			continue
		}
		out[parts[1]] = struct{}{}
	}
	return out
}

func extractOpenAPIOperations(src string) map[string]bool {
	lines := strings.Split(src, "\n")
	out := make(map[string]bool)
	var currentPath string
	inPaths := false
	for _, line := range lines {
		if strings.HasPrefix(line, "paths:") {
			inPaths = true
			continue
		}
		if !inPaths {
			continue
		}
		// Next top-level key ends paths section.
		if len(line) > 0 && line[0] != ' ' && line[0] != '#' && strings.HasSuffix(strings.TrimSpace(line), ":") {
			break
		}
		if m := openapiPathRE.FindStringSubmatch(line); len(m) == 2 {
			currentPath = m[1]
			continue
		}
		if currentPath == "" {
			continue
		}
		if m := openapiMethodRE.FindStringSubmatch(line); len(m) == 2 {
			out[fmt.Sprintf("%s %s", strings.ToUpper(m[1]), currentPath)] = true
		}
	}
	return out
}
