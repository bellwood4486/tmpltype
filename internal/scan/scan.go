package scan

import "sort"

// Kind は推論されたフィールド種別を表します。
type Kind int

const (
	KindString Kind = iota
	KindStruct
	KindSlice
	KindMap
)

// Field は推論スキーマ木のノードです。
type Field struct {
	Name     string
	Kind     Kind
	Elem     *Field            // Slice/Map の要素
	Children map[string]*Field // Struct の子
}

// Schema はトップレベル（Params直下）のフィールド集合です。
type Schema struct {
	Fields map[string]*Field
}

// ScanTemplate は Go テンプレートを AST 解析して、.(ドット）スコープを追跡して
// フィールド参照からスキーマ木を推論します。
// 既定では葉はすべて string として扱い、 range は []struct{} (子フィールドがあれば) または []string (なければ),
// index は map[string]string を推論します。
func ScanTemplate(src string) (Schema, error) {
	insp, err := inspect(src)
	if err != nil {
		return Schema{}, err
	}
	logInspection(insp)

	schema := buildSchema(insp)
	logSchema(schema)

	return schema, nil
}

// logInspection は inspection の内容をログ出力します。
func logInspection(insp inspection) {
	logRefCount(len(insp.refs))
	for _, ref := range insp.refs {
		logRef(ref.path, ref.usage)
	}
}

// logSchema は Schema のツリー構造をログ出力します。
func logSchema(schema Schema) {
	logSchemaFieldCount(len(schema.Fields))
	names := make([]string, 0, len(schema.Fields))
	for name := range schema.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		field := schema.Fields[name]
		logField(field, 0)
	}
}

// logField は Field をログ出力します。
func logField(f *Field, depth int) {
	logSchemaField(f.Name, f.Kind, depth)

	if f.Elem != nil {
		logSchemaElem(f.Name, f.Elem.Name, f.Elem.Kind, depth+1)
		if f.Elem.Children != nil {
			childNames := make([]string, 0, len(f.Elem.Children))
			for name := range f.Elem.Children {
				childNames = append(childNames, name)
			}
			sort.Strings(childNames)
			for _, name := range childNames {
				logField(f.Elem.Children[name], depth+2)
			}
		}
	}

	if f.Children != nil {
		childNames := make([]string, 0, len(f.Children))
		for name := range f.Children {
			childNames = append(childNames, name)
		}
		sort.Strings(childNames)
		for _, name := range childNames {
			logField(f.Children[name], depth+1)
		}
	}
}
