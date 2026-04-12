package tui

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const rootTUIImportPath = "easydocker/internal/tui"

func TestTUISubpackagesDoNotImportRootPackage(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve caller path")
	}

	tuiDir := filepath.Dir(thisFile)
	subpackageDirs := []string{"browse", "chrome", "logs", "mode", "state", "tables", "util"}

	for _, rel := range subpackageDirs {
		pkgDir := filepath.Join(tuiDir, rel)
		err := filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			fset := token.NewFileSet()
			file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if parseErr != nil {
				return parseErr
			}

			for _, imp := range file.Imports {
				importPath, unquoteErr := strconv.Unquote(imp.Path.Value)
				if unquoteErr != nil {
					return unquoteErr
				}
				if importPath == rootTUIImportPath {
					relPath, relErr := filepath.Rel(tuiDir, path)
					if relErr != nil {
						relPath = path
					}
					t.Errorf("subpackage file %s imports forbidden root package %q", filepath.ToSlash(relPath), rootTUIImportPath)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("failed walking %s: %v", rel, err)
		}
	}
}
