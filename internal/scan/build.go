package scan

import (
	"sort"
	"strings"

	"github.com/bellwood4486/tmpltype/internal/util"
)

// pathInfo はあるパスについて集計された情報を保持します。
type pathInfo struct {
	usages   map[usage]bool
	hasChild bool // より長いパス（子孫）が存在するか
}

// buildSchema は inspection からスキーマを構築します。
// 全てのフィールド参照を見て、各パスの型を決定します。
func buildSchema(insp inspection) Schema {
	if len(insp.refs) == 0 {
		return Schema{Fields: map[string]*Field{}}
	}

	// 1. パスごとに usage を集計
	info := make(map[string]*pathInfo)
	for _, ref := range insp.refs {
		key := strings.Join(ref.path, ".")
		if info[key] == nil {
			info[key] = &pathInfo{usages: make(map[usage]bool)}
		}
		info[key].usages[ref.usage] = true
	}

	// 2. 親パスを補完（子パスが存在する場合、親パスも info に追加）
	parentPaths := make(map[string]bool)
	for key := range info {
		parts := strings.Split(key, ".")
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], ".")
			parentPaths[parentPath] = true
		}
	}
	for parentPath := range parentPaths {
		if info[parentPath] == nil {
			info[parentPath] = &pathInfo{usages: make(map[usage]bool)}
		}
	}

	// 3. 各パスについて子パスが存在するか判定
	for key := range info {
		for otherKey := range info {
			if otherKey != key && strings.HasPrefix(otherKey, key+".") {
				info[key].hasChild = true
				break
			}
		}
	}

	// 4. 各パスの Kind を決定
	kindMap := make(map[string]Kind)
	for key, pi := range info {
		kindMap[key] = determineKind(pi)
	}

	// 5. スキーマ構造を構築
	schema := Schema{Fields: map[string]*Field{}}
	buildTree(&schema, info, kindMap)

	return schema
}

// determineKind はパス情報から Kind を決定します。
func determineKind(pi *pathInfo) Kind {
	// 優先順位: Map > Slice > Struct > String
	// usageScope（if/with の基点）は hasChild がある場合のみ Struct になる
	// 例: {{ if .Status }}{{ .Status }}{{ end }} → Status は String
	// 例: {{ with .User }}{{ .Name }}{{ end }} → User は Struct（User.Name があるため）
	if pi.usages[usageRangeMap] || pi.usages[usageIndex] {
		return KindMap
	}
	if pi.usages[usageRange] {
		return KindSlice
	}
	if pi.hasChild {
		return KindStruct
	}
	return KindString
}

// buildTree は pathInfo と kindMap から Schema のツリー構造を構築します。
func buildTree(schema *Schema, info map[string]*pathInfo, kindMap map[string]Kind) {
	// パスをソート（短いパスから処理するため）
	paths := make([]string, 0, len(info))
	for key := range info {
		paths = append(paths, key)
	}
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})

	// パスを処理（短い順なので親が先に処理される）
	for _, path := range paths {
		parts := strings.Split(path, ".")
		kind := kindMap[path]

		if len(parts) == 1 {
			// トップレベルフィールド
			name := parts[0]
			if schema.Fields[name] == nil {
				schema.Fields[name] = createField(name, kind, path, info)
			}
		} else {
			// 子フィールド（親は既に存在）
			insertField(schema, parts, kind, info)
		}
	}
}

// insertField はパスに対応するフィールドをスキーマに挿入します。
// トップレベルフィールドは buildTree で既に作成されています。
func insertField(schema *Schema, parts []string, kind Kind, info map[string]*pathInfo) {
	if len(parts) <= 1 {
		return
	}

	// 親フィールドを辿る（トップレベルは必ず存在する）
	cur := schema.Fields[parts[0]]
	for i := 1; i < len(parts); i++ {
		// Slice の場合は要素に潜る
		if cur.Kind == KindSlice {
			if cur.Elem == nil {
				// 親パスから Elem の Kind を決定
				parentPath := strings.Join(parts[:i], ".")
				elemKind := KindString
				var children map[string]*Field
				if pi, ok := info[parentPath]; ok && pi.hasChild {
					elemKind = KindStruct
					children = map[string]*Field{}
				}
				cur.Elem = &Field{
					Name:     cur.Name + "Item",
					Kind:     elemKind,
					Children: children,
				}
			}
			cur = cur.Elem
		}

		// Map の場合は値は String なので子フィールドは追加しない
		if cur.Kind == KindMap {
			return
		}

		if cur.Children == nil {
			cur.Children = map[string]*Field{}
		}

		segName := parts[i]
		if i == len(parts)-1 {
			// 最後のセグメント - 新しいフィールドを作成
			if cur.Children[segName] == nil {
				fullPath := strings.Join(parts, ".")
				cur.Children[segName] = createField(segName, kind, fullPath, info)
			}
		} else {
			// 中間セグメント - 既に存在するはず（短い順にソート済み）
			cur = cur.Children[segName]
		}
	}
}

// createField は名前と Kind からフィールドを作成します。
// path はフルパス（例: "Section.Items"）で、Slice 要素の Kind 判定に使用します。
func createField(name string, kind Kind, path string, info map[string]*pathInfo) *Field {
	field := &Field{
		Name: util.Export(name),
		Kind: kind,
	}

	// Struct の場合は Children を初期化
	if kind == KindStruct {
		field.Children = map[string]*Field{}
	}

	// Slice の場合は Elem を作成（子パスの有無で Kind を決定）
	if kind == KindSlice {
		elemKind := KindString
		var children map[string]*Field
		if pi, ok := info[path]; ok && pi.hasChild {
			elemKind = KindStruct
			children = map[string]*Field{}
		}
		field.Elem = &Field{
			Name:     util.Export(name) + "Item",
			Kind:     elemKind,
			Children: children,
		}
	}

	// Map の場合は Elem を作成
	if kind == KindMap {
		field.Elem = &Field{
			Name: util.Export(name) + "Value",
			Kind: KindString,
		}
	}

	return field
}
