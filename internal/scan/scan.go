package scan

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
	return buildSchema(insp), nil
}
