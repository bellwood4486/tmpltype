# Supported Template Syntax

`tmpltype` supports a wide range of Go template syntax patterns. The template scanner analyzes these patterns to automatically infer types for your template parameters.

## Table of Contents

- [Basic Patterns](#basic-patterns)
  - [Field Reference](#1-field-reference)
  - [Nested Field Reference](#2-nested-field-reference)
  - [Deep Nested Paths](#3-deep-nested-paths)
- [Control Flow](#control-flow)
  - [Conditional Statements (if)](#4-conditional-statements-if)
  - [With Statement](#5-with-statement)
  - [Else Clauses](#6-else-clauses)
- [Collections](#collections)
  - [Range Over Slice](#7-range-over-slice)
  - [Range Over Map](#8-range-over-map)
  - [Map Access](#9-map-access-with-index-function)
- [Advanced Patterns](#advanced-patterns)
  - [Nested Structures](#10-nested-structures-with--range)
  - [Multiple Ranges](#11-multiple-ranges)
  - [Complex Nesting](#12-complex-nesting)
- [Type Inference Rules](#type-inference-rules)
- [Limitations](#limitations)
- [Examples](#examples)

## Basic Patterns

### 1. Field Reference

**Template:**
```go
{{ .Title }}
```

**Inferred Type:**
```go
type TemplateParams struct {
    Title string
}
```

**Use Case:** Display a simple value

### 2. Nested Field Reference

**Template:**
```go
{{ .User.Name }}
{{ .User.Email }}
```

**Inferred Type:**
```go
type TemplateParamsUser struct {
    Email string
    Name  string
}

type TemplateParams struct {
    User TemplateParamsUser
}
```

**Use Case:** Access properties of an object

### 3. Deep Nested Paths

**Template:**
```go
{{ .Company.Department.Team.Manager.Name }}
```

**Inferred Type:**
```go
type TemplateParamsCompanyDepartmentTeamManager struct {
    Name string
}

type TemplateParamsCompanyDepartmentTeam struct {
    Manager TemplateParamsCompanyDepartmentTeamManager
}

type TemplateParamsCompanyDepartment struct {
    Team TemplateParamsCompanyDepartmentTeam
}

type TemplateParamsCompany struct {
    Department TemplateParamsCompanyDepartment
}

type TemplateParams struct {
    Company TemplateParamsCompany
}
```

**Use Case:** Deeply nested object hierarchies

**Note:** Type names can get long. Consider using `@param` for cleaner names or flattening your structure.

## Control Flow

### 4. Conditional Statements (if)

**Template:**
```go
{{ if .Status }}
  <p>Status: {{ .Status }}</p>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParams struct {
    Status string
}
```

**Key Points:**
- The field is inferred as `string` if used directly
- If it has nested fields, it's inferred as a struct
- Use `*string` if you need to distinguish between empty string and nil

### 5. With Statement

**Template:**
```go
{{ with .Summary }}
  <p>{{ .Content }}</p>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsSummary struct {
    Content string
}

type TemplateParams struct {
    Summary TemplateParamsSummary
}
```

**Key Points:**
- Changes the dot (`.`) context within the block
- The scanner correctly tracks scope changes
- Fields inside `with` block are relative to `.Summary`

### 6. Else Clauses

**Template:**
```go
{{ with .Summary }}
  <p>{{ .Content }}</p>
{{ else }}
  <p>{{ .DefaultMessage }}</p>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsSummary struct {
    Content string
}

type TemplateParams struct {
    DefaultMessage string
    Summary        TemplateParamsSummary
}
```

**Key Points:**
- Fields in `else` block are at the parent scope
- Both branches are analyzed for type inference

## Collections

### 7. Range Over Slice

**Template:**
```go
{{ range .Items }}
  <li>{{ .Title }} - {{ .ID }}</li>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsItemsItem struct {
    ID    string
    Title string
}

type TemplateParams struct {
    Items []TemplateParamsItemsItem
}
```

**Key Points:**
- `.Items` is inferred as a slice
- Fields accessed inside `range` become fields of the item struct
- Item type is named `<FieldName>Item` by convention

**Override with `@param` for specific types:**
```go
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
```

### 8. Range Over Map

**Template:**
```go
{{ range $key, $value := .Meta }}
  <dt>{{ $key }}</dt>
  <dd>{{ $value }}</dd>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParams struct {
    Meta map[string]string
}
```

**Key Points:**
- Two-variable range (`$k, $v := .Field`) is inferred as `map[string]string`
- Single-variable or no-variable range is inferred as slice

**Override for slice with index:**
```go
{{/* @param Products []struct{Name string} */}}
{{ range $i, $item := .Products }}
  <li>{{ $i }}: {{ $item.Name }}</li>
{{ end }}
```

### 9. Map Access with Index Function

**Template:**
```go
{{ index .Meta "key" }}
{{ index .Meta "author" }}
{{ index .Settings "theme" }}
```

**Inferred Type:**
```go
type TemplateParams struct {
    Meta     map[string]string
    Settings map[string]string
}
```

**Key Points:**
- `index` function signals map type
- Map key is always `string`
- Map value is inferred as `string` (override with `@param` for other types)

**Override for different value types:**
```go
{{/* @param Counters map[string]int */}}
{{ index .Counters "visits" }}
```

## Advanced Patterns

### 10. Nested Structures (with + range)

**Template:**
```go
{{ with .Project }}
  <h3>{{ .Name }}</h3>
  {{ range .Tasks }}
    <p>{{ .Title }} - {{ .Status }}</p>
  {{ end }}
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsProjectTasksItem struct {
    Status string
    Title  string
}

type TemplateParamsProject struct {
    Name  string
    Tasks []TemplateParamsProjectTasksItem
}

type TemplateParams struct {
    Project TemplateParamsProject
}
```

**Use Case:** Object containing a collection

### 11. Multiple Ranges

**Template:**
```go
{{ range .Authors }}
  <h4>{{ .Name }}</h4>
{{ end }}

{{ range .Tags }}
  <span>{{ . }}</span>
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsAuthorsItem struct {
    Name string
}

type TemplateParams struct {
    Authors []TemplateParamsAuthorsItem
    Tags    []string  // Simple slice ({{ . }} refers to item itself)
}
```

**Key Points:**
- `{{ . }}` in range creates `[]string` (slice of simple values)
- `{{ .Field }}` in range creates `[]struct{Field string}`

### 12. Complex Nesting

**Template:**
```go
{{ range .Categories }}
  <h2>{{ .Name }}</h2>
  {{ with .Settings }}
    <p>Theme: {{ index .Options "theme" }}</p>
  {{ end }}
  {{ range .Items }}
    <li>{{ .Title }}</li>
  {{ end }}
{{ end }}
```

**Inferred Type:**
```go
type TemplateParamsCategoriesItemSettingsOptions struct {
    // map[string]string inferred from index
}

type TemplateParamsCategoriesItemSettings struct {
    Options map[string]string
}

type TemplateParamsCategoriesItemItemsItem struct {
    Title string
}

type TemplateParamsCategoriesItem struct {
    Items    []TemplateParamsCategoriesItemItemsItem
    Name     string
    Settings TemplateParamsCategoriesItemSettings
}

type TemplateParams struct {
    Categories []TemplateParamsCategoriesItem
}
```

**Use Case:** Complex hierarchical data structures

## Type Inference Rules

### Default Type Inference

| Pattern | Inferred Type | Example |
|---------|---------------|---------|
| `{{ .Field }}` | `string` | `Title string` |
| `{{ .Obj.Field }}` | nested struct with `string` | `User struct{Name string}` |
| `{{ range .Items }}...{{ end }}` | `[]struct{...}` | `Items []struct{...}` |
| `{{ range $k, $v := .Map }}...{{ end }}` | `map[string]string` | `Meta map[string]string` |
| `{{ index .Map "key" }}` | `map[string]string` | `Meta map[string]string` |

### Scope Tracking

The scanner maintains a scope stack:

```go
// Root scope: TemplateParams
{{ .Title }}                    // TemplateParams.Title

{{ with .User }}                // Enter User scope
  {{ .Name }}                   // User.Name
  {{ range .Roles }}            // Enter Roles item scope
    {{ .Name }}                 // Roles[].Name
  {{ end }}                     // Exit Roles scope
{{ end }}                       // Exit User scope

{{ .Footer }}                   // Back to TemplateParams.Footer
```

### Field Naming

Generated struct fields:
- Are sorted **alphabetically**
- Use **exported** names (capitalized)
- Follow Go naming conventions

**Template:**
```go
{{ .user_name }}
{{ .user-email }}
{{ .UserID }}
```

**Generated:**
```go
type TemplateParams struct {
    UserEmail string  // from user-email
    UserID    string  // from UserID
    UserName  string  // from user_name (sorted alphabetically)
}
```

## Limitations

### Not Supported

❌ **Variables (assignments)**
```go
// Not analyzed for type inference
{{ $var := .Something }}
{{ $var.Field }}
```

**Workaround:** Access fields directly without variables, or use `@param`

❌ **Function calls on fields**
```go
// Function calls not analyzed
{{ .Title | upper }}
{{ printf "%s" .Name }}
```

**Workaround:** Fields used in functions are still inferred, but function results are not

❌ **Template definitions**
```go
// Template names not analyzed
{{ template "header" . }}
{{ define "footer" }}...{{ end }}
```

**Workaround:** Each template file is analyzed independently

❌ **Complex expressions**
```go
// Not fully analyzed
{{ if and .A .B }}
{{ if not .Flag }}
```

**Workaround:** Simple field references work; use `@param` for complex logic

❌ **Field access on `index` result**
```go
// Field access on index result not tracked
{{ (index .Users "admin").Name }}
```

**Workaround:** Use two-variable range instead: `{{ range $k, $v := .Users }}{{ $v.Name }}{{ end }}`, or use `@param` to specify the map value type

## Examples

### Complete Example

See [Example 04: Comprehensive Template](../examples/04_comprehensive_template/) for a working example demonstrating all supported patterns.

**Template files:**
- `basic_fields.tmpl` - Basic and nested field references
- `control_flow.tmpl` - If, with, else statements
- `collections.tmpl` - Range and map access
- `advanced.tmpl` - Complex nested structures

### Quick Reference by Use Case

| Use Case | Template Pattern | See Example |
|----------|------------------|-------------|
| Simple data | `{{ .Field }}` | [01_basic](../examples/01_basic/) |
| Nested objects | `{{ .User.Name }}` | [01_basic](../examples/01_basic/) |
| Lists | `{{ range .Items }}{{ .Title }}{{ end }}` | [04_comprehensive](../examples/04_comprehensive_template/) |
| Maps | `{{ index .Meta "key" }}` | [04_comprehensive](../examples/04_comprehensive_template/) |
| Conditionals | `{{ if .Flag }}...{{ end }}` | [04_comprehensive](../examples/04_comprehensive_template/) |
| Type override | `{{/* @param Age int */}}` | [02_param_directive](../examples/02_param_directive/) |
| Complex types | `{{/* @param Items []struct{...} */}}` | [05_all_param_types](../examples/05_all_param_types/) |

## See Also

- [Getting Started Guide](getting-started.md) - Tutorial and basics
- [`@param` Directive](param-directive.md) - Override inferred types
- [Examples](../examples/) - Working code examples
