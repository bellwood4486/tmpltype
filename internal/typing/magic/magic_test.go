package magic

import (
	"testing"
)

func TestParseParams_Simple(t *testing.T) {
	src := `
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{ .User.Name }}
`
	directives, err := ParseParams(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(directives) != 2 {
		t.Errorf("expected 2 directives, got %d", len(directives))
	}

	// Check first directive
	if directives[0].Path != "User.Age" {
		t.Errorf("unexpected path: %s", directives[0].Path)
	}
	if directives[0].Type.Kind != TypeKindBase || directives[0].Type.BaseType != "int" {
		t.Errorf("unexpected type for User.Age")
	}

	// Check second directive
	if directives[1].Path != "User.Email" {
		t.Errorf("unexpected path: %s", directives[1].Path)
	}
	if directives[1].Type.Kind != TypeKindPointer {
		t.Errorf("expected pointer type for User.Email")
	}
}

func TestParseParams_Multiple(t *testing.T) {
	src := `
{{/* @param Items []string */}}
{{/* @param Meta map[string]int */}}
{{/* @param Count int64 */}}
`
	directives, err := ParseParams(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(directives) != 3 {
		t.Errorf("expected 3 directives, got %d", len(directives))
	}
}

func TestParseParams_NoDirectives(t *testing.T) {
	src := `{{ .User.Name }}`
	directives, err := ParseParams(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(directives) != 0 {
		t.Errorf("expected 0 directives, got %d", len(directives))
	}
}

func TestTypeResolver_GetType(t *testing.T) {
	src := `
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
`
	resolver, err := NewTypeResolver(src)
	if err != nil {
		t.Fatal(err)
	}

	// Test User.Age type
	typ, ok := resolver.GetType([]string{"User", "Age"})
	if !ok {
		t.Error("expected User.Age to have type override")
	}
	if typ != "int" {
		t.Errorf("expected User.Age to be 'int', got %s", typ)
	}

	// Test User.Email type
	typ, ok = resolver.GetType([]string{"User", "Email"})
	if !ok {
		t.Error("expected User.Email to have type override")
	}
	if typ != "*string" {
		t.Errorf("expected User.Email to be '*string', got %s", typ)
	}

	// Test non-existent path
	_, ok = resolver.GetType([]string{"User", "Name"})
	if ok {
		t.Error("expected User.Name to not have type override")
	}
}

func TestTypeResolver_SliceType(t *testing.T) {
	src := `
{{/* @param Items []int */}}
{{/* @param Tags []string */}}
`
	resolver, err := NewTypeResolver(src)
	if err != nil {
		t.Fatal(err)
	}

	// Test Items type
	typ, ok := resolver.GetType([]string{"Items"})
	if !ok {
		t.Error("expected Items to have type override")
	}
	if typ != "[]int" {
		t.Errorf("expected Items to be '[]int', got %s", typ)
	}

	// Test Tags type
	typ, ok = resolver.GetType([]string{"Tags"})
	if !ok {
		t.Error("expected Tags to have type override")
	}
	if typ != "[]string" {
		t.Errorf("expected Tags to be '[]string', got %s", typ)
	}
}

func TestParseParams_TrimMarkers(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected int
	}{
		{
			name: "left trim marker without space",
			src: `{{-/* @param User.Age int */}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "left trim marker with space",
			src: `{{- /* @param User.Age int */}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "right trim marker without space",
			src: `{{/* @param User.Age int */-}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "right trim marker with space",
			src: `{{/* @param User.Age int */ -}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "both trim markers",
			src: `{{-/* @param User.Age int */-}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "both trim markers with spaces",
			src: `{{- /* @param User.Age int */ -}}
{{ .User.Name }}`,
			expected: 1,
		},
		{
			name: "multiple directives with trim markers",
			src: `{{-/* @param User.Age int */}}
{{-/* @param User.Email *string */}}
{{-/* @param Items []string */-}}
{{ .User.Name }}`,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directives, err := ParseParams(tt.src)
			if err != nil {
				t.Fatal(err)
			}

			if len(directives) != tt.expected {
				t.Errorf("expected %d directives, got %d", tt.expected, len(directives))
			}

			// Verify first directive is parsed correctly
			if len(directives) > 0 {
				if directives[0].Path != "User.Age" {
					t.Errorf("unexpected path: %s", directives[0].Path)
				}
				if directives[0].Type.Kind != TypeKindBase || directives[0].Type.BaseType != "int" {
					t.Errorf("unexpected type for User.Age")
				}
			}
		})
	}
}
