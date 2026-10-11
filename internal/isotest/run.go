package isotest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// Run builds upstream suites as executables so //go:isolate entry metadata is
// emitted. Ordinary go test binaries currently omit that compiler metadata.
func Run(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	source, _ := filepath.Glob("*.go")
	suites, _ := filepath.Glob("testdata/upstream/*.go")
	if len(suites) == 0 {
		t.Fatal("no sample suites")
	}
	temp, err := os.MkdirTemp(dir, ".isotest-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(temp) })
	var paths, tests, entries []string
	for _, path := range source {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, raw, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Doc == nil || fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
				continue
			}
			first, ok := fn.Type.Params.List[0].Type.(*ast.SelectorExpr)
			if !ok || first.Sel.Name != "Context" {
				continue
			}
			for _, c := range fn.Doc.List {
				if c.Text == "//go:isolate" {
					entries = append(entries, fn.Name.Name)
				}
			}
		}
	}
	moduleRoot := dir
	for {
		if _, err := os.Stat(filepath.Join(moduleRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(moduleRoot)
		if parent == moduleRoot {
			t.Fatal("no module")
		}
		moduleRoot = parent
	}
	mod, err := os.ReadFile(filepath.Join(moduleRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module := strings.Fields(string(mod))[1]
	relative, err := filepath.Rel(moduleRoot, dir)
	if err != nil {
		t.Fatal(err)
	}
	importPath := module
	if relative != "." {
		importPath += "/" + filepath.ToSlash(relative)
	}
	for _, path := range suites {
		if strings.HasSuffix(path, "_test.go") && !strings.HasPrefix(path, "testdata/") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		text = regexp.MustCompile(`(?m)^package \w+`).ReplaceAllString(text, "package main")
		if !strings.Contains(text, strconv.Quote(importPath)) {
			text = strings.Replace(text, "package main", "package main\nimport . "+strconv.Quote(importPath), 1)
		}
		text = strings.ReplaceAll(text, `"go.temporal.io/sdk/testsuite"`, `testsuite "github.com/mfateev/samples-go-poc/internal/isotest"`)
		f, err := parser.ParseFile(token.NewFileSet(), path, text, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "Test") && strings.HasPrefix(path, "testdata/") {
				tests = append(tests, fmt.Sprintf("{%q,func(t *testing.T){t.Cleanup(isotest.CloseEnvironments);%s(t)}}", fn.Name.Name, fn.Name.Name))
			}
			if fn.Doc != nil {
				for _, c := range fn.Doc.List {
					if c.Text == "//go:isolate" {
						entries = append(entries, fn.Name.Name)
					}
				}
			}
		}
		target := filepath.Join(temp, strings.TrimSuffix(filepath.Base(path), "_test.go")+"_fixture.go")
		if err := os.WriteFile(target, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, target)
	}
	_, harnessFile, _, _ := runtime.Caller(0)
	catalog, err := os.ReadFile(filepath.Join(filepath.Dir(harnessFile), "testdata/catalog.txt"))
	if err != nil {
		t.Fatal(err)
	}
	imports := strings.ReplaceAll(strings.TrimPrefix(string(catalog), "package isotest\n"), `_ "`+importPath+`"`, "")
	main := fmt.Sprintf("package main\nimport . %q\nimport(\"testing\";\"regexp\";\"github.com/mfateev/samples-go-poc/internal/isotest\")\nfunc main(){isotest.RegisterWorkflows(%s);testing.Main(regexp.MatchString,[]testing.InternalTest{%s},nil,nil)}\n", importPath, strings.Join(entries, ","), strings.Join(tests, ","))
	if len(entries) == 0 {
		main = strings.Replace(main, "import . "+strconv.Quote(importPath)+"\n", "", 1)
	}
	main = strings.Replace(main, "package main\n", "package main\n"+imports, 1)
	entry := filepath.Join(temp, "main.go")
	if err := os.WriteFile(entry, []byte(main), 0600); err != nil {
		t.Fatal(err)
	}
	_, file, _, _ := runtime.Caller(0)
	compiler := filepath.Join(filepath.Dir(file), "../../../golang-go/bin/go")
	binary := filepath.Join(temp, "sample")
	build := exec.Command(compiler, append([]string{"build", "-o", binary}, append(paths, entry)...)...)
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build sample: %v\n%s", err, out)
	}
	cmd := exec.Command(binary, "-test.v", "-test.timeout=90s")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sample suite: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}
