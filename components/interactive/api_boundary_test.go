package interactive_test

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

const forbiddenAPIPackagePrefix = "github.com/go-echarts/"

func TestPublicAPIIdentifiersAreRendererNeutral(t *testing.T) {
	t.Parallel()

	packages, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}
	var leaks []string
	for _, parsed := range packages {
		for _, file := range parsed.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				identifier, ok := node.(*ast.Ident)
				if ok && token.IsExported(identifier.Name) && strings.Contains(strings.ToLower(identifier.Name), "echarts") {
					leaks = append(leaks, identifier.Name)
				}
				return true
			})
		}
	}
	if len(leaks) != 0 {
		t.Fatalf("public API names expose renderer implementation: %s", strings.Join(leaks, ", "))
	}
}

func TestPublicAPIDocumentationIsRendererNeutral(t *testing.T) {
	t.Parallel()

	packages, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse package comments: %v", err)
	}
	for _, parsed := range packages {
		for _, file := range parsed.Files {
			for _, declaration := range file.Decls {
				switch current := declaration.(type) {
				case *ast.FuncDecl:
					if current.Name.IsExported() {
						rejectRendererDoc(t, current.Name.Name, current.Doc)
					}
				case *ast.GenDecl:
					for _, spec := range current.Specs {
						switch value := spec.(type) {
						case *ast.TypeSpec:
							if value.Name.IsExported() {
								rejectRendererDoc(t, value.Name.Name, value.Doc, current.Doc)
							}
						case *ast.ValueSpec:
							for _, name := range value.Names {
								if name.IsExported() {
									rejectRendererDoc(t, name.Name, value.Doc, current.Doc)
								}
							}
						}
					}
				}
			}
		}
	}
}

func rejectRendererDoc(t *testing.T, name string, groups ...*ast.CommentGroup) {
	t.Helper()
	for _, group := range groups {
		if group == nil {
			continue
		}
		text := strings.ToLower(group.Text())
		if strings.Contains(text, "echarts") || strings.Contains(text, "go-echarts") {
			t.Errorf("public API documentation for %s exposes renderer implementation", name)
		}
	}
}

func TestPublicAPIDoesNotExposeGoEChartsTypes(t *testing.T) {
	t.Parallel()

	packagePath, exports := packageExports(t)
	lookup := func(path string) (io.ReadCloser, error) {
		exportPath, ok := exports[path]
		if !ok || exportPath == "" {
			return nil, fmt.Errorf("no export data for %s", path)
		}
		return os.Open(exportPath)
	}

	loaded, err := importer.ForCompiler(token.NewFileSet(), "gc", lookup).Import(packagePath)
	if err != nil {
		t.Fatalf("import %s export data: %v", packagePath, err)
	}

	var leaks []string
	scope := loaded.Scope()
	for _, name := range scope.Names() {
		if !token.IsExported(name) {
			continue
		}
		object := scope.Lookup(name)
		findForbiddenTypes(name, object.Type(), packagePath, nil, &leaks)
	}

	sort.Strings(leaks)
	if len(leaks) != 0 {
		t.Fatalf("public API exposes go-echarts types:\n  %s", strings.Join(leaks, "\n  "))
	}
}

func packageExports(t *testing.T) (string, map[string]string) {
	t.Helper()
	command := exec.Command("go", "list", "-json", "-export", "-deps", ".")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go list package exports: %v\n%s", err, output)
	}

	type listedPackage struct {
		ImportPath string
		Export     string
		DepOnly    bool
	}
	exports := make(map[string]string)
	packagePath := ""
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed listedPackage
		if err := decoder.Decode(&listed); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode go list package exports: %v", err)
		}
		exports[listed.ImportPath] = listed.Export
		if !listed.DepOnly {
			packagePath = listed.ImportPath
		}
	}
	if packagePath == "" {
		t.Fatal("go list did not identify package under test")
	}
	return packagePath, exports
}

func findForbiddenTypes(path string, typ types.Type, localPackage string, seen map[types.Type]bool, leaks *[]string) {
	if typ == nil {
		return
	}
	if seen == nil {
		seen = make(map[types.Type]bool)
	}
	if seen[typ] {
		return
	}
	seen[typ] = true
	defer delete(seen, typ)

	switch current := typ.(type) {
	case *types.Alias:
		findForbiddenTypes(path, types.Unalias(current), localPackage, seen, leaks)
	case *types.Named:
		object := current.Obj()
		if object.Pkg() != nil {
			packagePath := object.Pkg().Path()
			if strings.HasPrefix(packagePath, forbiddenAPIPackagePrefix) {
				*leaks = append(*leaks, fmt.Sprintf("%s: %s", path, types.TypeString(current, packageQualifier)))
				return
			}
			if packagePath != localPackage {
				return
			}
		}
		findForbiddenTypes(path, current.Underlying(), localPackage, seen, leaks)
		for index := 0; index < current.TypeParams().Len(); index++ {
			findForbiddenTypes(path, current.TypeParams().At(index).Constraint(), localPackage, seen, leaks)
		}
		for index := 0; index < current.NumMethods(); index++ {
			method := current.Method(index)
			if method.Exported() {
				findForbiddenTypes(path+"."+method.Name(), method.Type(), localPackage, seen, leaks)
			}
		}
	case *types.Pointer:
		findForbiddenTypes(path, current.Elem(), localPackage, seen, leaks)
	case *types.Array:
		findForbiddenTypes(path, current.Elem(), localPackage, seen, leaks)
	case *types.Slice:
		findForbiddenTypes(path, current.Elem(), localPackage, seen, leaks)
	case *types.Map:
		findForbiddenTypes(path+" key", current.Key(), localPackage, seen, leaks)
		findForbiddenTypes(path+" value", current.Elem(), localPackage, seen, leaks)
	case *types.Chan:
		findForbiddenTypes(path, current.Elem(), localPackage, seen, leaks)
	case *types.Struct:
		for index := 0; index < current.NumFields(); index++ {
			field := current.Field(index)
			if field.Exported() {
				findForbiddenTypes(path+"."+field.Name(), field.Type(), localPackage, seen, leaks)
			}
		}
	case *types.Signature:
		findTupleForbiddenTypes(path+" parameter", current.Params(), localPackage, seen, leaks)
		findTupleForbiddenTypes(path+" result", current.Results(), localPackage, seen, leaks)
		for index := 0; index < current.TypeParams().Len(); index++ {
			findForbiddenTypes(path, current.TypeParams().At(index).Constraint(), localPackage, seen, leaks)
		}
	case *types.Interface:
		for index := 0; index < current.NumExplicitMethods(); index++ {
			method := current.ExplicitMethod(index)
			if method.Exported() {
				findForbiddenTypes(path+"."+method.Name(), method.Type(), localPackage, seen, leaks)
			}
		}
		for index := 0; index < current.NumEmbeddeds(); index++ {
			findForbiddenTypes(path, current.EmbeddedType(index), localPackage, seen, leaks)
		}
	case *types.TypeParam:
		findForbiddenTypes(path, current.Constraint(), localPackage, seen, leaks)
	case *types.Union:
		for index := 0; index < current.Len(); index++ {
			findForbiddenTypes(path, current.Term(index).Type(), localPackage, seen, leaks)
		}
	}
}

func findTupleForbiddenTypes(path string, tuple *types.Tuple, localPackage string, seen map[types.Type]bool, leaks *[]string) {
	for index := 0; index < tuple.Len(); index++ {
		findForbiddenTypes(fmt.Sprintf("%s %d", path, index), tuple.At(index).Type(), localPackage, seen, leaks)
	}
}

func packageQualifier(pkg *types.Package) string {
	return pkg.Path()
}
