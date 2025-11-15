package gen_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bellwood4486/tmpltype/internal/gen"
)

func parseCode(t *testing.T, code string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "gen.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v\ncode:\n%s", err, code)
	}
	return f
}

func hasImport(f *ast.File, path string, name string) bool {
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, "\"")
		var n string
		if imp.Name != nil {
			n = imp.Name.Name
		}
		if p == path {
			if name == "" && n == "" {
				return true
			}
			if name == n {
				return true
			}
		}
	}
	return false
}

func findType(f *ast.File, name string) *ast.StructType {
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, s := range gd.Specs {
			ts := s.(*ast.TypeSpec)
			if ts.Name.Name == name {
				st, _ := ts.Type.(*ast.StructType)
				return st
			}
		}
	}
	return nil
}

func findFunc(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			return fd
		}
	}
	return nil
}

func TestEmit_BasicScaffoldAndTypes(t *testing.T) {
	u := gen.Unit{
		TemplateName: "tpl",
		Pkg:          "x",
		FilePath:    "tpl.tmpl",
		Source:       "{{ .User.Name }}\n{{ .Message }}\n",
	}

	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Check MainCode
	code := result.MainCode
	if !strings.Contains(code, "Option(\"missingkey=error\")") {
		t.Fatalf("missing Template Option missingkey=error\n%s", code)
	}

	// Check SourcesCode contains the template variable
	if !strings.Contains(result.SourcesCode, "var tplTplSource = `") {
		t.Fatalf("missing template source variable in SourcesCode\n%s", result.SourcesCode)
	}

	// AST checks on MainCode
	f := parseCode(t, code)
	if f.Name.Name != u.Pkg {
		t.Fatalf("package name = %s; want %s", f.Name.Name, u.Pkg)
	}
	// embedはもう使わないので確認しない
	if !hasImport(f, "io", "") || !hasImport(f, "text/template", "") {
		t.Fatalf("imports io or text/template not found")
	}

	// type TplUser struct{ Name string } (新しいフォーマット)
	user := findType(f, "TplUser")
	if user == nil || user.Fields == nil || len(user.Fields.List) == 0 {
		t.Fatalf("type TplUser struct not found or empty")
	}
	if len(user.Fields.List) != 1 || len(user.Fields.List[0].Names) != 1 || user.Fields.List[0].Names[0].Name != "Name" {
		t.Fatalf("TplUser fields unexpected")
	}
	if id, ok := user.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "string" {
		t.Fatalf("TplUser.Name type != string")
	}

	// type Tpl { Message string; User TplUser } with sorted order (新しいフォーマット)
	params := findType(f, "Tpl")
	if params == nil || params.Fields == nil || len(params.Fields.List) != 2 {
		t.Fatalf("Tpl fields unexpected")
	}
	if params.Fields.List[0].Names[0].Name != "Message" {
		t.Fatalf("Tpl first field = %s; want Message", params.Fields.List[0].Names[0].Name)
	}
	if id, ok := params.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "string" {
		t.Fatalf("Tpl.Message type != string")
	}
	if params.Fields.List[1].Names[0].Name != "User" {
		t.Fatalf("Tpl second field = %s; want User", params.Fields.List[1].Names[0].Name)
	}
	if id, ok := params.Fields.List[1].Type.(*ast.Ident); !ok || id.Name != "TplUser" {
		t.Fatalf("Tpl.User type != TplUser")
	}

	// RenderTpl and Render signatures (新しいフォーマット)
	renderTpl := findFunc(f, "RenderTpl")
	if renderTpl == nil || renderTpl.Type == nil || renderTpl.Type.Params == nil || renderTpl.Type.Results == nil {
		t.Fatalf("RenderTpl signature not found")
	}
	if len(renderTpl.Type.Params.List) != 2 || len(renderTpl.Type.Results.List) != 1 {
		t.Fatalf("RenderTpl parameters/results unexpected")
	}
	// w io.Writer
	if se, ok := renderTpl.Type.Params.List[0].Type.(*ast.SelectorExpr); !ok || se.Sel.Name != "Writer" {
		t.Fatalf("RenderTpl first param not io.Writer")
	}
	if id, ok := renderTpl.Type.Params.List[1].Type.(*ast.Ident); !ok || id.Name != "Tpl" {
		t.Fatalf("RenderTpl second param not Tpl")
	}
	if id, ok := renderTpl.Type.Results.List[0].Type.(*ast.Ident); !ok || id.Name != "error" {
		t.Fatalf("RenderTpl result not error")
	}

	// 汎用Render関数: Render(w io.Writer, name TemplateName, data any) error
	render := findFunc(f, "Render")
	if render == nil || len(render.Type.Params.List) != 3 {
		t.Fatalf("Render signature not found")
	}
	if id, ok := render.Type.Params.List[1].Type.(*ast.Ident); !ok || id.Name != "TemplateName" {
		t.Fatalf("Render second param not TemplateName")
	}
	if id, ok := render.Type.Params.List[2].Type.(*ast.Ident); !ok || id.Name != "any" {
		t.Fatalf("Render third param not any")
	}
}

func TestEmit_RangeAndIndex_TypesAndOrder(t *testing.T) {
	u := gen.Unit{
		TemplateName: "email",
		Pkg:          "x",
		FilePath:    "email.tmpl",
		Source:       "{{ range .Items }}{{ .Title }}{{ .ID }}{{ end }}\n{{ index .Meta \"env\" }}\n",
	}
	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}
	f := parseCode(t, result.MainCode)

	// type EmailItemsItem with fields Title, ID (order sorted) - 新しいフォーマット
	it := findType(f, "EmailItemsItem")
	if it == nil || it.Fields == nil || len(it.Fields.List) != 2 {
		t.Fatalf("EmailItemsItem struct unexpected")
	}
	if it.Fields.List[0].Names[0].Name != "ID" || it.Fields.List[1].Names[0].Name != "Title" {
		t.Fatalf("EmailItemsItem fields not sorted as expected: got %s, %s", it.Fields.List[0].Names[0].Name, it.Fields.List[1].Names[0].Name)
	}

	params := findType(f, "Email")  // 新しいフォーマット
	if params == nil || len(params.Fields.List) != 2 {
		t.Fatalf("Email unexpected")
	}
	if params.Fields.List[0].Names[0].Name != "Items" {
		t.Fatalf("Email first field = %s; want Items", params.Fields.List[0].Names[0].Name)
	}
	if at, ok := params.Fields.List[0].Type.(*ast.ArrayType); !ok {
		t.Fatalf("Email.Items not a slice")
	} else {
		if id, ok := at.Elt.(*ast.Ident); !ok || id.Name != "EmailItemsItem" {
			t.Fatalf("Email.Items element not EmailItemsItem")
		}
	}
	if params.Fields.List[1].Names[0].Name != "Meta" {
		t.Fatalf("Email second field = %s; want Meta", params.Fields.List[1].Names[0].Name)
	}
	if mt, ok := params.Fields.List[1].Type.(*ast.MapType); !ok {
		t.Fatalf("Email.Meta not a map")
	} else {
		if k, ok := mt.Key.(*ast.Ident); !ok || k.Name != "string" {
			t.Fatalf("Email.Meta key not string")
		}
		if v, ok := mt.Value.(*ast.Ident); !ok || v.Name != "string" {
			t.Fatalf("Email.Meta value not string")
		}
	}
}

func TestEmit_Golden_Simple(t *testing.T) {
	u := gen.Unit{TemplateName: "tpl", Pkg: "x", FilePath: "tpl.tmpl", Source: "{{ .User.Name }}\n{{ .Message }}\n"}
	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Check MainCode golden
	goldenPath := filepath.Join("testdata", "simple.golden")
	b, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden: %v", err)
	}
	want := string(b)
	if result.MainCode != want {
		// On mismatch, it helps to see a unified-ish diff. Keep it short.
		t.Fatalf("golden mismatch\n--- want\n%s\n--- got\n%s", want, result.MainCode)
	}

	// Check SourcesCode golden
	sourcesGoldenPath := filepath.Join("testdata", "simple_sources.golden")
	sourcesB, err := os.ReadFile(sourcesGoldenPath)
	if err != nil {
		t.Fatalf("failed to read sources golden: %v", err)
	}
	sourcesWant := string(sourcesB)
	if result.SourcesCode != sourcesWant {
		t.Fatalf("sources golden mismatch\n--- want\n%s\n--- got\n%s", sourcesWant, result.SourcesCode)
	}
}

func TestEmit_CompilesInTempModule(t *testing.T) {
	if runtime.GOOS == "js" || runtime.GOOS == "wasip1" {
		t.Skip("skip on restricted platforms")
	}

	u := gen.Unit{TemplateName: "tpl", Pkg: "x", FilePath: "tpl.tmpl", Source: "Hello {{ .Message }}"}
	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	dir := t.TempDir()
	// Create module
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmpmod\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write generated code (2 files now)
	if err := os.WriteFile(filepath.Join(dir, "gen.go"), []byte(result.MainCode), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "gen_sources_gen.go"), []byte(result.SourcesCode), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	// Ensure build cache is writable within sandbox
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(dir, ".gocache"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
}

func TestEmit_WithParamOverride_BasicTypes(t *testing.T) {
	src := `
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{ .User.Name }} is {{ .User.Age }} years old.
{{ if .User.Email }}Email: {{ .User.Email }}{{ end }}
`
	u := gen.Unit{
		TemplateName: "tpl",
		Pkg:          "x",
		FilePath:    "tpl.tmpl",
		Source:       src,
	}

	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	f := parseCode(t, result.MainCode)

	// Check TplUser struct has Age int and Email *string (新しいフォーマット)
	user := findType(f, "TplUser")
	if user == nil {
		t.Fatal("TplUser type not found")
	}

	foundAge := false
	foundEmail := false
	for _, field := range user.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		if name == "Age" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "int" {
				foundAge = true
			}
		}
		if name == "Email" {
			if st, ok := field.Type.(*ast.StarExpr); ok {
				if id, ok := st.X.(*ast.Ident); ok && id.Name == "string" {
					foundEmail = true
				}
			}
		}
	}

	if !foundAge {
		t.Error("User.Age int not found")
	}
	if !foundEmail {
		t.Error("User.Email *string not found")
	}
}

func TestEmit_WithParamOverride_SliceType(t *testing.T) {
	src := `
{{/* @param Items []struct{ID int64; Title string} */}}
{{ range .Items }}{{ .ID }}: {{ .Title }}{{ end }}
`
	u := gen.Unit{
		TemplateName: "tpl",
		Pkg:          "x",
		FilePath:    "tpl.tmpl",
		Source:       src,
	}

	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	f := parseCode(t, result.MainCode)

	// Check TplItemsItem has ID int64 and Title string (新しいフォーマット)
	item := findType(f, "TplItemsItem")
	if item == nil {
		t.Fatal("TplItemsItem type not found")
	}

	foundID := false
	foundTitle := false
	for _, field := range item.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		if name == "ID" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "int64" {
				foundID = true
			}
		}
		if name == "Title" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "string" {
				foundTitle = true
			}
		}
	}

	if !foundID {
		t.Error("TplItemsItem.ID int64 not found")
	}
	if !foundTitle {
		t.Error("TplItemsItem.Title string not found")
	}
}

func TestEmit_TemplateWithBacktick(t *testing.T) {
	src := "Code example: `{{ .Code }}`"
	u := gen.Unit{
		TemplateName: "tpl",
		Pkg:          "x",
		FilePath:    "tpl.tmpl",
		Source:       src,
	}

	result, err := gen.Emit([]gen.Unit{u})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// SourcesCode should use double quotes instead of backticks
	if strings.Contains(result.SourcesCode, "= `") {
		t.Fatalf("SourcesCode should not use backticks when template contains backtick\n%s", result.SourcesCode)
	}
	if !strings.Contains(result.SourcesCode, "= \"") {
		t.Fatalf("SourcesCode should use double quotes\n%s", result.SourcesCode)
	}

	// Verify it compiles and contains the original content
	if !strings.Contains(result.SourcesCode, "Code example:") {
		t.Fatalf("SourcesCode missing template content\n%s", result.SourcesCode)
	}

	// MainCode should parse correctly
	parseCode(t, result.MainCode)
	// SourcesCode should parse correctly
	parseCode(t, result.SourcesCode)
}

