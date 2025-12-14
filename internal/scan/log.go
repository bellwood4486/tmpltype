package scan

import (
	"sort"
	"strings"

	"github.com/bellwood4486/tmpltype/internal/logger"
)

// ログ出力のアダプター層
// 本処理からフォーマットの詳細を分離するため、ここでslog用の可変長引数に変換を行う

func logTemplate(name string) {
	logger.Debug("template", "name", name)
}

func logRefCount(count int) {
	logger.Debug("ref", "count", count)
}

func logRef(path []string, u usage) {
	logger.Debug("ref", "path", strings.Join(path, "."), "usage", usageString(u))
}

func logPathInfoCount(count int) {
	logger.Debug("pathinfo", "count", count)
}

func logPathInfo(path string, pi *pathInfo) {
	usages := make([]string, 0, len(pi.usages))
	for u := range pi.usages {
		usages = append(usages, usageString(u))
	}
	sort.Strings(usages)
	logger.Debug("pathinfo",
		"path", path,
		"usages", strings.Join(usages, ","),
		"hasChild", pi.hasChild,
	)
}

func logKindCount(count int) {
	logger.Debug("kind", "count", count)
}

func logKind(path string, k Kind) {
	logger.Debug("kind", "path", path, "kind", kindString(k))
}

func logSchemaFieldCount(count int) {
	logger.Debug("schema", "fields", count)
}

func logSchemaField(name string, kind Kind, depth int) {
	indent := strings.Repeat("  ", depth)
	logger.Debug("schema", "field", indent+name, "kind", kindString(kind))
}

func logSchemaElem(parentName string, elemName string, kind Kind, depth int) {
	indent := strings.Repeat("  ", depth)
	logger.Debug("schema", "field", indent+"[elem] "+elemName, "kind", kindString(kind))
}

// usageString は usage を文字列に変換します。
func usageString(u usage) string {
	switch u {
	case usageLeaf:
		return "leaf"
	case usageRange:
		return "range"
	case usageRangeMap:
		return "rangemap"
	case usageIndex:
		return "index"
	case usageScope:
		return "scope"
	default:
		return "unknown"
	}
}

// kindString は Kind を文字列に変換します。
func kindString(k Kind) string {
	switch k {
	case KindString:
		return "String"
	case KindStruct:
		return "Struct"
	case KindSlice:
		return "Slice"
	case KindMap:
		return "Map"
	default:
		return "Unknown"
	}
}
