package scan

import (
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

	// 2. 各パスについて子パスが存在するか判定
	for key := range info {
		for otherKey := range info {
			if otherKey != key && strings.HasPrefix(otherKey, key+".") {
				info[key].hasChild = true
				break
			}
		}
	}

	// 3. 各パスの Kind を決定
	kindMap := make(map[string]Kind)
	for key, pi := range info {
		kindMap[key] = determineKind(pi)
	}

	// 4. スキーマ構造を構築
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
	sortByLength(paths)

	// 各パスについてフィールドを作成
	for _, key := range paths {
		parts := strings.Split(key, ".")
		kind := kindMap[key]

		// 親パスが Slice の場合、このパスは要素のフィールドとして扱う
		insertField(schema, parts, kind, kindMap, info)
	}
}

// sortByLength はパスを長さ順（短い順）にソートします。
func sortByLength(paths []string) {
	// シンプルなバブルソート（パス数は少ないので十分）
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if len(paths[i]) > len(paths[j]) {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}
}

// insertField はパスに対応するフィールドをスキーマに挿入します。
func insertField(schema *Schema, parts []string, kind Kind, kindMap map[string]Kind, info map[string]*pathInfo) {
	if len(parts) == 0 {
		return
	}

	// トップレベルフィールドを取得または作成
	name := parts[0]
	cur := schema.Fields[name]
	if cur == nil {
		topKind := KindStruct
		if len(parts) == 1 {
			topKind = kind
		} else if k, ok := kindMap[name]; ok {
			topKind = k
		}
		cur = &Field{
			Name: util.Export(name),
			Kind: topKind,
		}
		schema.Fields[name] = cur
	}

	// Slice/Map の場合は Elem を確保
	ensureElem(cur, name, info)

	// 単一セグメントの場合は終了
	if len(parts) == 1 {
		return
	}

	// 中間パスを辿る
	pathSoFar := name
	for i := 1; i < len(parts); i++ {
		// Slice の場合は要素に潜る
		if cur.Kind == KindSlice {
			if cur.Elem == nil {
				// Slice 要素の Kind を決定: 子パスがあれば Struct、なければ String
				elemKind := KindString
				var children map[string]*Field
				if pi, ok := info[pathSoFar]; ok && pi.hasChild {
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

		// Map の場合は要素に潜る
		if cur.Kind == KindMap {
			if cur.Elem == nil {
				cur.Elem = &Field{
					Name: cur.Name + "Value",
					Kind: KindString,
				}
			}
			// Map の値は通常 String なので、子フィールドは追加しない
			return
		}

		if cur.Children == nil {
			cur.Children = map[string]*Field{}
		}

		segName := parts[i]
		pathSoFar = pathSoFar + "." + segName

		if i == len(parts)-1 {
			// 最後のセグメント
			if cur.Children[segName] == nil {
				cur.Children[segName] = &Field{
					Name: util.Export(segName),
					Kind: kind,
				}
			}
		} else {
			// 中間セグメント
			if cur.Children[segName] == nil {
				midKind := KindStruct
				if k, ok := kindMap[pathSoFar]; ok {
					midKind = k
				}
				cur.Children[segName] = &Field{
					Name: util.Export(segName),
					Kind: midKind,
				}
			}
			cur = cur.Children[segName]
		}
	}
}

// ensureElem は Slice/Map フィールドの Elem を確保します。
// Slice 要素の Kind は info から子パスの有無を判定して決定します。
func ensureElem(f *Field, path string, info map[string]*pathInfo) {
	if f == nil {
		return
	}
	if f.Kind == KindSlice && f.Elem == nil {
		// Slice 要素の Kind を決定: 子パスがあれば Struct、なければ String
		elemKind := KindString
		var children map[string]*Field
		if pi, ok := info[path]; ok && pi.hasChild {
			elemKind = KindStruct
			children = map[string]*Field{}
		}
		f.Elem = &Field{
			Name:     f.Name + "Item",
			Kind:     elemKind,
			Children: children,
		}
	}
	if f.Kind == KindMap && f.Elem == nil {
		f.Elem = &Field{
			Name: f.Name + "Value",
			Kind: KindString,
		}
	}
}

