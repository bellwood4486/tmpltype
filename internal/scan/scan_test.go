package scan_test

import (
	"testing"

	"github.com/bellwood4486/tmpltype/internal/scan"
)

func TestScanTemplate_SimpleFieldsAndNested(t *testing.T) {
	src := `
{{ .User.Name }}
{{ .Message }}
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	user := getTop(t, sch, "User")
	assertKind(t, user, scan.KindStruct)
	name := getChild(t, user, "Name")
	assertKind(t, name, scan.KindString)

	msg := getTop(t, sch, "Message")
	assertKind(t, msg, scan.KindString)
}

func TestScanTemplate_DeepNestedPath(t *testing.T) {
	src := `{{ .User.Address.City }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	user := getTop(t, sch, "User")
	addr := getChild(t, user, "Address")
	city := getChild(t, addr, "City")
	assertKind(t, city, scan.KindString)
}

func TestScanTemplate_WithScope_ElseRestoresDot(t *testing.T) {
	src := `
{{ with .User }}
	{{ .Name }}
{{ else }}
	{{ .Message }}
{{ end }}
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	user := getTop(t, sch, "User")
	assertKind(t, user, scan.KindStruct)
	name := getChild(t, user, "Name")
	assertKind(t, name, scan.KindString)

	// else 側は元のドット（トップレベル）に戻るはず
	msg := getTop(t, sch, "Message")
	assertKind(t, msg, scan.KindString)
}

func TestScanTemplate_Range_MakesSliceAndElementStruct(t *testing.T) {
	src := `
<ul>
{{ range .Items }}
	<li>{{ .Title }} #{{ .ID }}</li>
{{ end }}
</ul>
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	items := getTop(t, sch, "Items")
	assertKind(t, items, scan.KindSlice)
	if items.Elem == nil {
		t.Fatal("Items.Elem is nil")
	}
	assertKind(t, items.Elem, scan.KindStruct)
	title := getChild(t, items.Elem, "Title")
	assertKind(t, title, scan.KindString)
	id := getChild(t, items.Elem, "ID")
	assertKind(t, id, scan.KindString)
}

func TestScanTemplate_Range_EmptyBody_MakesSliceString(t *testing.T) {
	// range のbodyが空の場合、要素は []string になる
	src := `{{ range .Items }}{{ end }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	items := getTop(t, sch, "Items")
	assertKind(t, items, scan.KindSlice)
	if items.Elem == nil {
		t.Fatal("Items.Elem is nil")
	}
	assertKind(t, items.Elem, scan.KindString)
	if items.Elem.Children != nil {
		t.Errorf("Items.Elem.Children should be nil, got %v", items.Elem.Children)
	}
}

func TestScanTemplate_Range_DotOnly_MakesSliceString(t *testing.T) {
	// range のbodyでドット自体のみを参照する場合、要素は []string になる
	src := `{{ range .Tags }}{{ . }}{{ end }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	tags := getTop(t, sch, "Tags")
	assertKind(t, tags, scan.KindSlice)
	if tags.Elem == nil {
		t.Fatal("Tags.Elem is nil")
	}
	assertKind(t, tags.Elem, scan.KindString)
	if tags.Elem.Children != nil {
		t.Errorf("Tags.Elem.Children should be nil, got %v", tags.Elem.Children)
	}
}

func TestScanTemplate_Index_MakesMapString(t *testing.T) {
	src := `{{ index .Meta "env" }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	meta := getTop(t, sch, "Meta")
	assertKind(t, meta, scan.KindMap)
	if meta.Elem == nil {
		t.Fatal("Meta.Elem is nil")
	}
	assertKind(t, meta.Elem, scan.KindString)
}

func TestScanTemplate_Range_TwoVariables_MakesMap(t *testing.T) {
	// range $k, $v := .Field → map[string]string
	src := `{{ range $k, $v := .Meta }}{{ $k }}={{ $v }}{{ end }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	meta := getTop(t, sch, "Meta")
	assertKind(t, meta, scan.KindMap)
	if meta.Elem == nil {
		t.Fatal("Meta.Elem is nil")
	}
	assertKind(t, meta.Elem, scan.KindString)
}

func TestScanTemplate_Range_OneVariable_MakesSlice(t *testing.T) {
	// range $item := .Items → []struct{}
	src := `{{ range $item := .Items }}{{ $item }}{{ end }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	items := getTop(t, sch, "Items")
	assertKind(t, items, scan.KindSlice)
}

func TestScanTemplate_WithThenRange_NestedUnderPrefix(t *testing.T) {
	src := `
{{ with .Section }}
	{{ range .Items }}{{ .Title }}{{ end }}
{{ end}}
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	section := getTop(t, sch, "Section")
	items := getChild(t, section, "Items")
	assertKind(t, items, scan.KindSlice)
	if items.Elem == nil {
		t.Fatal("Items.Elem is nil")
	}
	assertKind(t, items.Elem, scan.KindStruct)
	title := getChild(t, items.Elem, "Title")
	assertKind(t, title, scan.KindString)
}

func TestScanTemplate_IfPipeLine_fieldAndChild(t *testing.T) {
	src := `
{{ if .User }}ok{{ end }}
{{ .User.Name }}
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	user := getTop(t, sch, "User")
	assertKind(t, user, scan.KindStruct)
	name := getChild(t, user, "Name")
	assertKind(t, name, scan.KindString)
}

func getTop(t *testing.T, s scan.Schema, name string) *scan.Field {
	t.Helper()
	f := s.Fields[name]
	if f == nil {
		t.Fatalf("top-level field %q not found", name)
	}
	return f
}

func getChild(t *testing.T, f *scan.Field, name string) *scan.Field {
	t.Helper()
	if f.Children == nil {
		t.Fatalf("field %q has no children", f.Name)
	}
	ch, ok := f.Children[name]
	if !ok || ch == nil {
		t.Fatalf("child %q not found under %q", name, f.Name)
	}
	return ch
}

func assertKind(t *testing.T, got *scan.Field, want scan.Kind) {
	t.Helper()
	if got.Kind != want {
		t.Fatalf("kind mismatch: got=%v want=%v", got.Kind, want)
	}
}
