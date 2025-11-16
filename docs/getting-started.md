# Getting Started with tmpltype

This guide will walk you through your first steps with `tmpltype`, from installation to creating your first type-safe template.

## Table of Contents

- [Installation](#installation)
- [Your First Template](#your-first-template)
- [Understanding Type Inference](#understanding-type-inference)
- [Using Type Directives](#using-type-directives)
- [Working with Multiple Templates](#working-with-multiple-templates)
- [Next Steps](#next-steps)

## Installation

Install `tmpltype` using `go install`:

```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

Or add it to your project:

```bash
go get github.com/bellwood4486/tmpltype
```

Verify the installation:

```bash
tmpltype -h
```

## Your First Template

Let's create a simple email template and generate type-safe code for it.

### Step 1: Create Your Project Structure

```bash
mkdir myproject
cd myproject
go mod init myproject
```

Create a templates directory:

```bash
mkdir templates
```

### Step 2: Create a Template File

Create `templates/email.tmpl`:

```html
<h1>Hello {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
<p>Sent at: {{ .Timestamp }}</p>
```

### Step 3: Add a go:generate Directive

Create `gen.go` in your project root:

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

### Step 4: Generate Code

Run code generation:

```bash
go generate
```

This creates `template_gen.go` with:
- A `Email` struct with fields: `User`, `Message`, `Timestamp`
- A nested `EmailUser` struct with `Name` field
- A `RenderEmail()` function for type-safe rendering
- Template initialization and management functions

### Step 5: Use the Generated Code

Create `main.go`:

```go
package main

import (
    "bytes"
    "fmt"
    "time"
)

func main() {
    // Initialize templates (required)
    InitTemplates()

    var buf bytes.Buffer

    err := RenderEmail(&buf, Email{
        User: EmailUser{
            Name: "Alice",
        },
        Message:   "Welcome to tmpltype!",
        Timestamp: time.Now().Format(time.RFC3339),
    })

    if err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
}
```

Run your program:

```bash
go run .
```

**Output:**
```html
<h1>Hello Alice</h1>
<p>Welcome to tmpltype!</p>
<p>Sent at: 2025-11-10T15:04:05Z</p>
```

## Understanding Type Inference

`tmpltype` automatically infers types from your template syntax:

| Template Syntax | Inferred Type | Why |
|----------------|---------------|-----|
| `{{ .Name }}` | `string` | Simple field reference |
| `{{ .User.Name }}` | Nested struct with `string` field | Nested field reference |
| `{{ range .Items }}{{ .Title }}{{ end }}` | `[]struct{Title string}` | Range over collection |
| `{{ index .Meta "key" }}` | `map[string]string` | Map access with index |

### Example: Complex Template

```html
<!-- templates/blog.tmpl -->
<article>
  <h1>{{ .Title }}</h1>
  <p>By {{ .Author.Name }} ({{ .Author.Email }})</p>

  <ul>
  {{ range .Tags }}
    <li>{{ . }}</li>
  {{ end }}
  </ul>

  <div>{{ .Content }}</div>
</article>
```

Generated struct:

```go
type BlogAuthor struct {
    Email string
    Name  string
}

type Blog struct {
    Author  BlogAuthor
    Content string
    Tags    []string
    Title   string
}
```

**Key Points:**
- All field references become `string` by default
- Nested paths create nested structs
- `range` creates slices
- Fields are sorted alphabetically

## Using Type Directives

When you need specific types (not `string`), use `@param` directives in your template comments.

### Example: User Profile with Typed Fields

Create `templates/profile.tmpl`:

```html
{{/* @param User.ID int64 */}}
{{/* @param User.Age int */}}
{{/* @param User.Bio *string */}}
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}

<div class="profile">
  <h2>User #{{ .User.ID }}: {{ .User.Name }}</h2>
  <p>Age: {{ .User.Age }}</p>

  {{ if .User.Bio }}
  <blockquote>{{ .User.Bio }}</blockquote>
  {{ end }}

  <h3>Recent Purchases</h3>
  <ul>
  {{ range .Items }}
    <li>#{{ .ID }}: {{ .Title }} - ${{ .Price }}</li>
  {{ end }}
  </ul>
</div>
```

Generated types:

```go
type ProfileItemsItem struct {
    ID    int64
    Price float64
    Title string
}

type ProfileUser struct {
    Age  int
    Bio  *string  // Optional field (pointer)
    ID   int64
    Name string
}

type Profile struct {
    Items []ProfileItemsItem
    User  ProfileUser
}
```

Usage:

```go
bio := "Go enthusiast and coffee lover"

var buf bytes.Buffer
err := RenderProfile(&buf, Profile{
    User: ProfileUser{
        ID:   123,
        Name: "Alice",
        Age:  30,
        Bio:  &bio,  // Optional field
    },
    Items: []ProfileItemsItem{
        {ID: 1, Title: "Go Programming Book", Price: 49.99},
        {ID: 2, Title: "Mechanical Keyboard", Price: 129.99},
    },
})
```

### Common Type Patterns

```go
// Basic types
{{/* @param Age int */}}
{{/* @param Price float64 */}}
{{/* @param Active bool */}}

// Optional fields (pointers)
{{/* @param Email *string */}}
{{/* @param Score *int */}}

// Collections
{{/* @param Tags []string */}}
{{/* @param Scores []int */}}

// Maps
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}

// Complex structures
{{/* @param Items []struct{ID int64; Name string; Active bool} */}}

// Nested fields
{{/* @param Config.Database.Host string */}}
{{/* @param Config.Database.Port int */}}
```

For complete type reference, see the [`@param` Directive documentation](param-directive.md).

## Working with Multiple Templates

`tmpltype` can process multiple templates at once using the `-dir` option.

### Example Project Structure

```
myproject/
├── gen.go
├── main.go
├── template_gen.go (generated)
└── templates/
    ├── email.tmpl
    ├── sms.tmpl
    └── push.tmpl
```

### Single Generation Command

```go
//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

This generates:
- `Email`, `Sms`, `Push` structs
- `RenderEmail()`, `RenderSms()`, `RenderPush()` functions
- A `Template` namespace for dynamic rendering
- A generic `Render()` function

### Type-Safe Rendering

```go
// Type-safe: compile-time errors for wrong fields
var buf bytes.Buffer
_ = RenderEmail(&buf, Email{...})
_ = RenderSms(&buf, Sms{...})
```

### Dynamic Rendering

```go
// Dynamic: runtime template selection
var buf bytes.Buffer
templateName := getTemplateNameFromConfig()
_ = Render(&buf, templateName, data)
```

For organizing templates in subdirectories, see the [Template Grouping documentation](template-grouping.md).

## Using Custom Template Functions

`tmpltype` supports custom template functions through a functional option pattern.

### Example: Email Template with Custom Functions

Create `templates/email.tmpl`:

```html
{{/* @param CreatedAt time.Time */}}
<h1>{{ .Title | upper }}</h1>
<p>Created: {{ formatDate .CreatedAt }}</p>
<p>{{ myCustomFunction .Message }}</p>
```

Create `funcs.go` with your custom functions:

```go
package main

import (
    "html/template"
    "strings"
    "time"
)

func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "upper": strings.ToUpper,
        "formatDate": func(t time.Time) string {
            return t.Format("2006-01-02")
        },
        "myCustomFunction": func(s string) string {
            return "✨ " + s + " ✨"
        },
    }
}
```

Initialize templates with your custom functions:

```go
package main

import (
    "bytes"
    "fmt"
    "time"
)

func main() {
    // Initialize templates with custom functions
    InitTemplates(WithFuncs(GetTemplateFuncs()))

    var buf bytes.Buffer
    err := RenderEmail(&buf, Email{
        Title:     "Welcome",
        Message:   "Hello World",
        CreatedAt: time.Now(),
    })

    if err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
}
```

**Key Points:**
- Call `InitTemplates()` with `WithFuncs()` option before rendering
- Custom functions work seamlessly with type inference
- Functions not in the preset list are automatically handled during code generation

For a complete example, see [Example 08: Custom Functions](../examples/08_custom_functions/).

## Next Steps

Now that you understand the basics, explore:

- **[CLI Reference](cli-reference.md)** - All command-line options
- **[Template Syntax](template-syntax.md)** - Supported Go template constructs
- **[`@param` Directive](param-directive.md)** - Complete type directive reference
- **[Template Grouping](template-grouping.md)** - Organize templates in subdirectories
- **[Examples](../examples/)** - Working code examples

### Recommended Learning Path

1. **[Example 01: Basic](../examples/01_basic/)** - Start here for basic type inference
2. **[Example 02: Param Directive](../examples/02_param_directive/)** - Learn `@param` usage
3. **[Example 03: Multi Template](../examples/03_multi_template/)** - Multiple templates
4. **[Example 07: Grouping](../examples/07_grouping/)** - Template organization
5. **[Example 08: Custom Functions](../examples/08_custom_functions/)** - Using custom template functions

Each example is fully runnable with `go generate && go run .`
