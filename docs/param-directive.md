# `@param` Directive Reference

The `@param` directive allows you to explicitly specify types for template parameters, overriding automatic type inference. This is essential for complex types like specific integer sizes, optional fields (pointers), and structured data.

## Table of Contents

- [Syntax](#syntax)
- [Why Use @param?](#why-use-param)
- [Supported Types](#supported-types)
  - [Basic Types](#basic-types)
  - [Pointer Types](#pointer-types-optionalnullable)
  - [Slices](#slices)
  - [Maps](#maps)
  - [Nested Struct Fields](#nested-struct-fields-dot-notation)
  - [Slice of Structs](#slice-of-structs)
  - [Optional Slices](#optional-slices)
- [Known Limitations](#known-limitations)
- [Best Practices](#best-practices)
- [Complete Examples](#complete-examples)

## Syntax

```go
{{/* @param <FieldPath> <Type> */}}
```

**Parameters:**
- `<FieldPath>`: Dot-separated field path (e.g., `User.Name`, `Items`, `Config.Database.Host`)
- `<Type>`: Go type expression (see supported types below)

**Example:**
```go
{{/* @param User.Age int */}}
{{/* @param Items []struct{ID int64; Title string} */}}
```

## Why Use @param?

By default, `tmpltype` infers all fields as `string`. Use `@param` when you need:

### ✅ Specific Numeric Types

```go
{{/* @param User.ID int64 */}}        // Database ID
{{/* @param Price float64 */}}         // Decimal precision
{{/* @param Count int */}}             // Integer count
```

### ✅ Optional Fields

```go
{{/* @param Email *string */}}         // May be nil
{{/* @param Score *int */}}            // May be nil
```

### ✅ Complex Structures

```go
{{/* @param Items []struct{ID int64; Name string; Price float64} */}}
{{/* @param Config map[string]int */}}
```

### ✅ Type Safety

```go
// Without @param: runtime error if you pass int
{{/* Template uses {{ .Age }} */}}
RenderTemplate(w, Params{Age: "25"})  // OK (string)
RenderTemplate(w, Params{Age: 25})    // ❌ Compile error

// With @param: compile-time safety
{{/* @param Age int */}}
{{/* Template uses {{ .Age }} */}}
RenderTemplate(w, Params{Age: 25})    // ✅ OK (int)
RenderTemplate(w, Params{Age: "25"})  // ❌ Compile error
```

## Supported Types

### Basic Types

All Go basic types are supported:

```go
{{/* @param Name string */}}
{{/* @param Age int */}}
{{/* @param UserID int64 */}}
{{/* @param Score int32 */}}
{{/* @param Tiny int8 */}}

{{/* @param Count uint */}}
{{/* @param Size uint64 */}}

{{/* @param Price float64 */}}
{{/* @param Rating float32 */}}

{{/* @param Active bool */}}

{{/* @param Data byte */}}
{{/* @param Char rune */}}

{{/* @param Value any */}}  // Interface{} equivalent
```

**Supported types:** `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `byte`, `rune`, `any`

### Pointer Types (Optional/Nullable)

Wrap any type with `*` to make it optional:

```go
{{/* @param Email *string */}}
{{/* @param Age *int */}}
{{/* @param Score *int64 */}}
{{/* @param Price *float64 */}}
{{/* @param Active *bool */}}
```

**Generated:**
```go
type TemplateParams struct {
    Email  *string
    Age    *int
    Score  *int64
    Price  *float64
    Active *bool
}
```

**Usage:**
```go
email := "alice@example.com"
params := TemplateParams{
    Email: &email,  // Pointer to string
    Age:   nil,     // Can be nil
}
```

**Template usage:**
```html
{{/* @param Email *string */}}
{{ if .Email }}
  <p>Email: {{ .Email }}</p>
{{ else }}
  <p>No email provided</p>
{{ end }}
```

### Slices

Slices of any basic type:

```go
{{/* @param Tags []string */}}
{{/* @param IDs []int */}}
{{/* @param Scores []int64 */}}
{{/* @param Prices []float64 */}}
{{/* @param Flags []bool */}}
```

**Generated:**
```go
type TemplateParams struct {
    Tags   []string
    IDs    []int
    Scores []int64
    Prices []float64
    Flags  []bool
}
```

**Usage:**
```go
params := TemplateParams{
    Tags:   []string{"go", "template", "codegen"},
    IDs:    []int{1, 2, 3},
    Prices: []float64{9.99, 19.99, 29.99},
}
```

### Maps

Maps with `string` keys:

```go
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}
{{/* @param Scores map[string]int64 */}}
{{/* @param Prices map[string]float64 */}}
{{/* @param Flags map[string]bool */}}
```

**Generated:**
```go
type TemplateParams struct {
    Metadata map[string]string
    Counters map[string]int
    Scores   map[string]int64
    Prices   map[string]float64
    Flags    map[string]bool
}
```

**Usage:**
```go
params := TemplateParams{
    Metadata: map[string]string{
        "author": "Alice",
        "version": "1.0",
    },
    Counters: map[string]int{
        "views": 100,
        "likes": 42,
    },
}
```

**⚠️ Important:** Map keys must always be `string`. Other key types are not supported.

```go
// ❌ Not supported
{{/* @param Lookup map[int]string */}}
{{/* @param Index map[int64]bool */}}
```

### Nested Struct Fields (Dot Notation)

Use dot notation to define nested struct fields:

```go
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
{{/* @param User.Email string */}}
{{/* @param User.Age int */}}
```

**Generated:**
```go
type TemplateParamsUser struct {
    Age   int
    Email string
    ID    int64
    Name  string
}

type TemplateParams struct {
    User TemplateParamsUser
}
```

**Deep nesting:**
```go
{{/* @param Config.Database.Host string */}}
{{/* @param Config.Database.Port int */}}
{{/* @param Config.Database.MaxConns int */}}
```

**Generated:**
```go
type TemplateParamsConfigDatabase struct {
    Host     string
    MaxConns int
    Port     int
}

type TemplateParamsConfig struct {
    Database TemplateParamsConfigDatabase
}

type TemplateParams struct {
    Config TemplateParamsConfig
}
```

### Slice of Structs

Define inline struct types for slice elements:

```go
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
```

**Generated:**
```go
type TemplateParamsItemsItem struct {
    ID    int64
    Price float64
    Title string
}

type TemplateParams struct {
    Items []TemplateParamsItemsItem
}
```

**Nested fields within struct:**
```go
{{/* @param Records []struct{Name string; Tags []string; Score *int} */}}
```

**Generated:**
```go
type TemplateParamsRecordsItem struct {
    Name  string
    Score *int
    Tags  []string
}

type TemplateParams struct {
    Records []TemplateParamsRecordsItem
}
```

**⚠️ Important:** Use **semicolons** (`;`) to separate struct fields, not commas.

```go
// ❌ Wrong - commas not allowed
{{/* @param Item struct{Name string, ID int} */}}

// ✅ Correct - use semicolons
{{/* @param Item struct{Name string; ID int} */}}
```

### Optional Slices

Make the entire slice optional with `*`:

```go
{{/* @param Tags *[]string */}}
{{/* @param Scores *[]int */}}
```

**Generated:**
```go
type TemplateParams struct {
    Tags   *[]string
    Scores *[]int
}
```

**Usage:**
```go
tags := []string{"go", "template"}
params := TemplateParams{
    Tags:   &tags,  // Pointer to slice
    Scores: nil,    // Can be nil
}
```

## Known Limitations

### ❌ Nested Slices/Maps

Directly nested slices and maps are not supported:

```go
// ❌ Does NOT work - generates invalid syntax
{{/* @param Matrix [][]string */}}
{{/* @param Groups map[string][]string */}}
{{/* @param Data []map[string]int */}}
```

**Workaround:** Use slice of structs:

```go
// ✅ Works - wrap in struct
{{/* @param Matrix []struct{Row []string} */}}
{{/* @param Groups []struct{Key string; Values []string} */}}
{{/* @param Data []struct{Items map[string]int} */}}
```

**Example:**
```go
// Instead of [][]string, use:
{{/* @param Matrix []struct{Row []string} */}}

// Generated:
type TemplateParamsMatrixItem struct {
    Row []string
}

type TemplateParams struct {
    Matrix []TemplateParamsMatrixItem
}

// Usage:
params := TemplateParams{
    Matrix: []TemplateParamsMatrixItem{
        {Row: []string{"a", "b", "c"}},
        {Row: []string{"d", "e", "f"}},
    },
}
```

### ❌ Inline Struct at Top Level

Cannot use inline `struct{...}` directly at top level:

```go
// ❌ Does NOT work - generates invalid Go code
{{/* @param User struct{ID int64; Name string} */}}
```

**Workaround:** Use dot notation:

```go
// ✅ Works
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
```

### ❌ Deeply Nested Paths with Inline Structs

Cannot combine deep paths with inline struct definitions:

```go
// ❌ Does NOT work - generates type names with dots
{{/* @param Complex.Nested.User struct{ID int64; Name string} */}}
```

**Workaround:** Flatten the structure:

```go
// ✅ Works
{{/* @param Complex.Nested.User.ID int64 */}}
{{/* @param Complex.Nested.User.Name string */}}
```

### ❌ Non-String Map Keys

Map keys must always be `string`:

```go
// ❌ Not supported
{{/* @param Lookup map[int]string */}}
{{/* @param Index map[int64]bool */}}
{{/* @param IDs map[uint64]string */}}
```

**Workaround:** Convert keys to strings in your code before passing to template.

## Best Practices

### ✅ DO

**Always use trim markers (`{{-` and `-}}`) with `@param` directives:**

Since `@param` directives don't produce any output, they should always use Go template trim markers to avoid blank lines in the rendered output.

```go
// ✅ Recommended: No blank lines in output
{{- /* @param User.Name string */ -}}
{{- /* @param User.Age int */ -}}
{{- /* @param Items []struct{ID int64; Title string} */ -}}
<div>Content starts here</div>

// ❌ Not recommended: Creates blank lines
{{/* @param User.Name string */}}
{{/* @param User.Age int */}}
{{/* @param Items []struct{ID int64; Title string} */}}
<div>Content starts here</div>  {{/* 3 blank lines before this */}}
```

**Use dot notation for nested structures:**
```go
{{- /* @param User.Name string */ -}}
{{- /* @param Config.Database.Host string */ -}}
```

**Use `[]struct{...}` for complex collections:**
```go
{{- /* @param Items []struct{ID int64; Name string; Price float64} */ -}}
```

**Use pointer types for optional fields:**
```go
{{- /* @param Email *string */ -}}
{{- /* @param Score *int */ -}}
```

**Keep field paths relatively flat (1-2 levels):**
```go
// ✅ Good
{{/* @param User.Name string */}}
{{/* @param Config.Host string */}}

// ⚠️ Works but verbose
{{/* @param App.Config.Database.Connection.Pool.MaxSize int */}}
```

**Use semicolons in struct fields:**
```go
{{/* @param Item struct{Name string; ID int} */}}
```

### ❌ DON'T

**Don't use inline `struct{...}` at top level:**
```go
// ❌ Wrong
{{/* @param User struct{Name string} */}}

// ✅ Right
{{/* @param User.Name string */}}
```

**Don't nest slices/maps directly:**
```go
// ❌ Wrong
{{/* @param Matrix [][]string */}}

// ✅ Right
{{/* @param Matrix []struct{Row []string} */}}
```

**Don't combine deep paths with inline structs:**
```go
// ❌ Wrong
{{/* @param A.B.C.D struct{X int} */}}

// ✅ Right
{{/* @param A.B.C.D.X int */}}
```

**Don't use commas in struct definitions:**
```go
// ❌ Wrong
{{/* @param Item struct{Name string, ID int} */}}

// ✅ Right
{{/* @param Item struct{Name string; ID int} */}}
```

## Complete Examples

### Example 1: E-commerce Product

**Template:**
```html
{{/* @param Product.ID int64 */}}
{{/* @param Product.Price float64 */}}
{{/* @param Product.InStock bool */}}
{{/* @param Product.Description *string */}}
{{/* @param Tags []string */}}
{{/* @param Reviews []struct{Rating int; Comment string; Author string} */}}

<div class="product">
  <h2>{{ .Product.Name }} (#{{ .Product.ID }})</h2>
  <p class="price">${{ .Product.Price }}</p>

  {{ if .Product.InStock }}
    <span class="badge">In Stock</span>
  {{ else }}
    <span class="badge out">Out of Stock</span>
  {{ end }}

  {{ if .Product.Description }}
    <p>{{ .Product.Description }}</p>
  {{ end }}

  <div class="tags">
    {{ range .Tags }}
      <span class="tag">{{ . }}</span>
    {{ end }}
  </div>

  <div class="reviews">
    {{ range .Reviews }}
      <div class="review">
        <span class="rating">{{ .Rating }}/5</span>
        <p>{{ .Comment }}</p>
        <small>- {{ .Author }}</small>
      </div>
    {{ end }}
  </div>
</div>
```

### Example 2: User Profile

**Template:**
```html
{{/* @param User.ID int64 */}}
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{/* @param User.Bio *string */}}
{{/* @param Metadata map[string]string */}}
{{/* @param Stats map[string]int */}}

<div class="profile">
  <h1>{{ .User.Name }} (ID: {{ .User.ID }})</h1>
  <p>Age: {{ .User.Age }}</p>

  {{ if .User.Email }}
    <p>Email: {{ .User.Email }}</p>
  {{ end }}

  {{ if .User.Bio }}
    <blockquote>{{ .User.Bio }}</blockquote>
  {{ end }}

  <dl class="metadata">
    {{ range $key, $val := .Metadata }}
      <dt>{{ $key }}</dt>
      <dd>{{ $val }}</dd>
    {{ end }}
  </dl>

  <dl class="stats">
    {{ range $key, $val := .Stats }}
      <dt>{{ $key }}</dt>
      <dd>{{ $val }}</dd>
    {{ end }}
  </dl>
</div>
```

### Example 3: Complete Reference

See [Example 05: All Param Types](../examples/05_all_param_types/) for a comprehensive, runnable example demonstrating:
- All supported basic types
- Pointer types (optional fields)
- Slices and maps
- Nested struct fields
- Slice of structs
- All known limitations with workarounds

Run the example:
```bash
cd examples/05_all_param_types
go generate
go run .
```

## See Also

- [Getting Started Guide](getting-started.md) - Tutorial and basics
- [Template Syntax](template-syntax.md) - Supported template constructs
- [CLI Reference](cli-reference.md) - Command-line options
- [Example 02: Param Directive](../examples/02_param_directive/) - Basic `@param` usage
- [Example 05: All Param Types](../examples/05_all_param_types/) - Complete reference
