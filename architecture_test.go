// Package main_test provides repository-wide architectural boundary tests.
package main_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHexagonalArchitectureEnforcement verifies that all Go components strictly adhere
// to Hexagonal Architecture (Ports & Adapters) dependency boundaries.
//
// Rules enforced:
// 1. Domain / Core (`domain`, `core`):
//   - Zero dependencies on `app`, `adapters`, or `ports`.
//   - Zero dependencies on external infrastructure (MongoDB, NATS, K8s, net/http, database/sql).
//
// 2. Application (`app`):
//   - Zero dependencies on `adapters`.
//
// 3. Ports (`ports`):
//   - Zero dependencies on `adapters` or `app`.
func TestHexagonalArchitectureEnforcement(t *testing.T) {
	componentDirs := []string{
		"catalog",
		"control-plane",
		"data-dashboard",
		"data-plane",
		"identity-hub",
		"internal",
	}

	for _, compDir := range componentDirs {
		if _, err := os.Stat(compDir); os.IsNotExist(err) {
			continue
		}

		t.Run(compDir, func(t *testing.T) {
			checkDomainDependencies(t, compDir)
			checkAppDependencies(t, compDir)
			checkPortsDependencies(t, compDir)
		})
	}
}

// checkDomainDependencies inspects domain/core files to guarantee no infrastructure or outer-layer imports exist.
func checkDomainDependencies(t *testing.T, baseDir string) {
	subDirs := []string{"domain", "core"}

	for _, sub := range subDirs {
		domainDir := filepath.Join(baseDir, sub)
		if _, err := os.Stat(domainDir); os.IsNotExist(err) {
			continue
		}

		forbiddenSubstrings := []string{
			"/adapters",
			"/ports",
			"go.mongodb.org",
			"github.com/nats-io",
			"k8s.io",
			"database/sql",
		}

		inspectImports(t, domainDir, baseDir+"/"+sub, func(importPath string, filePath string) {
			for _, forbidden := range forbiddenSubstrings {
				if strings.Contains(importPath, forbidden) {
					t.Errorf("[HEXAGONAL ARCHITECTURE VIOLATION] Domain/Core layer in %s imports forbidden package '%s' in file %s",
						domainDir, importPath, filePath)
				}
			}
		})
	}
}

// checkAppDependencies inspects app files to guarantee no direct adapter imports exist.
func checkAppDependencies(t *testing.T, baseDir string) {
	appDir := filepath.Join(baseDir, "app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return
	}

	inspectImports(t, appDir, "app", func(importPath string, filePath string) {
		if strings.Contains(importPath, "/adapters") {
			t.Errorf("[HEXAGONAL ARCHITECTURE VIOLATION] App usecase layer in %s imports adapter package '%s' in file %s. Adapters must be injected via Ports!",
				appDir, importPath, filePath)
		}
	})
}

// checkPortsDependencies inspects ports files to ensure ports do not import adapters or app implementation details.
func checkPortsDependencies(t *testing.T, baseDir string) {
	portsDir := filepath.Join(baseDir, "ports")
	if _, err := os.Stat(portsDir); os.IsNotExist(err) {
		return
	}

	inspectImports(t, portsDir, "ports", func(importPath string, filePath string) {
		if strings.Contains(importPath, "/adapters") {
			t.Errorf("[HEXAGONAL ARCHITECTURE VIOLATION] Ports layer in %s imports forbidden package '%s' in file %s",
				portsDir, importPath, filePath)
		}
	})
}

// inspectImports parses all non-test Go source files in dirPath and invokes validateImport for each import path.
func inspectImports(t *testing.T, dirPath string, layerName string, validateImport func(importPath string, filePath string)) {
	fset := token.NewFileSet()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("failed to parse Go file %s: %v", path, parseErr)
			return parseErr
		}

		for _, imp := range node.Imports {
			cleanImport := strings.Trim(imp.Path.Value, `"`)
			validateImport(cleanImport, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("error inspecting imports for %s: %v", dirPath, err)
	}
}
