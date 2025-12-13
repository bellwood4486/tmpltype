package scan_test

import (
	"testing"

	"github.com/bellwood4486/tmpltype/internal/scan"
)

func TestScanTemplate_Range_MapWithFieldRef_MakesMapStruct(t *testing.T) {
	// $v.Name を参照 → map[string]struct{Name string}
	src := `{{ range $k, $v := .Users }}{{ $v.Name }}{{ end }}`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	users := getTop(t, sch, "Users")
	assertKind(t, users, scan.KindMap)
	if users.Elem == nil {
		t.Fatal("Users.Elem is nil")
	}
	assertKind(t, users.Elem, scan.KindStruct)
	name := getChild(t, users.Elem, "Name")
	assertKind(t, name, scan.KindString)
}

func TestScanTemplate_Range_MapWithValueOnly_MakesMapString(t *testing.T) {
	// $v のみを参照 → map[string]string
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
