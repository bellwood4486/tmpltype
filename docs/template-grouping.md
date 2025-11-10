# Template Grouping

Organize your templates logically in subdirectories to create nested namespaces and improve project structure.

## Table of Contents

- [Overview](#overview)
- [Why Use Template Grouping?](#why-use-template-grouping)
- [Directory Structure](#directory-structure)
- [Generated Code Structure](#generated-code-structure)
- [Usage Patterns](#usage-patterns)
- [Naming Conventions](#naming-conventions)
- [Best Practices](#best-practices)
- [Complete Example](#complete-example)

## Overview

Template grouping allows you to organize related templates into subdirectories. Each subdirectory becomes a nested namespace in the generated code, providing better organization and avoiding name conflicts.

**Basic concept:**
```
templates/
├── footer.tmpl          # Flat template → Footer
└── mail_invite/         # Group → MailInvite namespace
    ├── title.tmpl       #   → MailInvite.Title
    └── content.tmpl     #   → MailInvite.Content
```

## Why Use Template Grouping?

### ✅ Logical Organization

Group related templates together:

```
templates/
├── mail_invite/
│   ├── title.tmpl
│   └── content.tmpl
├── mail_welcome/
│   ├── title.tmpl
│   └── content.tmpl
└── mail_reset_password/
    ├── title.tmpl
    └── content.tmpl
```

All invitation email templates are in one directory, welcome email templates in another, etc.

### ✅ Avoid Name Conflicts

Without grouping, you'd need unique names for every template:

```
# ❌ Without grouping - verbose names
templates/
├── mail_invite_title.tmpl
├── mail_invite_content.tmpl
├── mail_welcome_title.tmpl
├── mail_welcome_content.tmpl
└── mail_reset_password_title.tmpl
```

With grouping, you can reuse simple names within each group:

```
# ✅ With grouping - clean names
templates/
├── mail_invite/
│   ├── title.tmpl
│   └── content.tmpl
└── mail_welcome/
    ├── title.tmpl
    └── content.tmpl
```

### ✅ Better Navigation

Find templates quickly:
- All mail-related templates are under `mail_*/`
- All dashboard-related templates are under `dashboard_*/`
- Shared templates stay at the root

### ✅ Type-Safe Namespaces

Generated code reflects your organization:

```go
// Access templates through nested namespaces
Template.MailInvite.Title
Template.MailInvite.Content
Template.MailWelcome.Title
Template.MailWelcome.Content
```

## Directory Structure

### Flat vs Grouped Templates

You can mix both approaches in the same project:

```
templates/
├── header.tmpl                    # Flat (root level)
├── footer.tmpl                    # Flat (root level)
├── mail_invite/                   # Group
│   ├── title.tmpl
│   └── content.tmpl
└── mail_account_created/          # Group
    ├── title.tmpl
    └── content.tmpl
```

### Scanning Depth

**Important:** `tmpltype` scans templates at **depth 0 and 1 only**.

**✅ Scanned:**
```
templates/
├── email.tmpl              ← depth 0 (flat)
└── mail/
    └── invite.tmpl         ← depth 1 (grouped)
```

**❌ Not scanned:**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl       ← depth 2 (not scanned)
```

### Recommended Structure

**Pattern:** `<category>_<name>/`

```
templates/
├── shared_header.tmpl              # Shared templates (flat)
├── shared_footer.tmpl
├── mail_invite/                    # Email: Invite
│   ├── title.tmpl
│   └── content.tmpl
├── mail_welcome/                   # Email: Welcome
│   ├── title.tmpl
│   └── content.tmpl
├── dashboard_summary/              # Dashboard: Summary
│   └── widget.tmpl
└── dashboard_activity/             # Dashboard: Activity
    └── widget.tmpl
```

**Benefits:**
- Clear categorization (mail, dashboard, etc.)
- Easy to find related templates
- Numeric prefixes for ordering (optional)

## Generated Code Structure

### Template Namespace

`tmpltype` generates a nested `Template` struct:

```go
var Template = struct {
    // Flat templates
    SharedFooter TemplateName
    SharedHeader TemplateName

    // Grouped templates
    MailInvite struct {
        Title   TemplateName
        Content TemplateName
    }
    MailWelcome struct {
        Title   TemplateName
        Content TemplateName
    }
    DashboardSummary struct {
        Widget TemplateName
    }
    DashboardActivity struct {
        Widget TemplateName
    }
}{
    SharedFooter: "shared_footer",
    SharedHeader: "shared_header",
    MailInvite: struct {
        Title   TemplateName
        Content TemplateName
    }{
        Title:   "mail_invite/title",
        Content: "mail_invite/content",
    },
    // ...
}
```

### Type-Safe Render Functions

Each template gets its own render function:

**Flat templates:**
```go
func RenderSharedHeader(w io.Writer, p SharedHeader) error
func RenderSharedFooter(w io.Writer, p SharedFooter) error
```

**Grouped templates:**
```go
func RenderMailInviteTitle(w io.Writer, p MailInviteTitle) error
func RenderMailInviteContent(w io.Writer, p MailInviteContent) error
func RenderMailWelcomeTitle(w io.Writer, p MailWelcomeTitle) error
func RenderMailWelcomeContent(w io.Writer, p MailWelcomeContent) error
```

### Type Names

**Pattern:** `<GroupName><TemplateName>`

| Template Path | Type Name | Render Function |
|--------------|-----------|-----------------|
| `footer.tmpl` | `Footer` | `RenderFooter()` |
| `mail_invite/title.tmpl` | `MailInviteTitle` | `RenderMailInviteTitle()` |
| `mail_invite/content.tmpl` | `MailInviteContent` | `RenderMailInviteContent()` |
| `dashboard_summary/widget.tmpl` | `DashboardSummaryWidget` | `RenderDashboardSummaryWidget()` |

## Usage Patterns

### Type-Safe Rendering (Recommended)

Use the generated type-safe functions:

```go
var buf bytes.Buffer

// Render grouped template
err := RenderMailInviteTitle(&buf, MailInviteTitle{
    SiteName:    "MyApp",
    InviterName: "Alice",
})

// Render flat template
err = RenderSharedFooter(&buf, SharedFooter{
    Year:    2025,
    Company: "MyCompany",
})
```

**Benefits:**
- ✅ Compile-time type checking
- ✅ IDE autocompletion
- ✅ Catch errors early

### Dynamic Rendering

Use the generic `Render()` function with the `Template` namespace:

```go
var buf bytes.Buffer

// Select template dynamically
templateName := getTemplateFromConfig()

// Use Template namespace for type-safe template name
err := Render(&buf, Template.MailInvite.Title, data)
```

**Use cases:**
- Configuration-driven template selection
- Dynamic template switching based on user preferences
- A/B testing with different templates

### Mixing Both Approaches

```go
func sendEmail(emailType string, data any) error {
    var buf bytes.Buffer

    // Dynamic selection
    var templateName TemplateName
    switch emailType {
    case "invite":
        templateName = Template.MailInvite.Content
    case "welcome":
        templateName = Template.MailWelcome.Content
    default:
        return fmt.Errorf("unknown email type: %s", emailType)
    }

    // Generic render
    err := Render(&buf, templateName, data)
    if err != nil {
        return err
    }

    // Send email...
    return nil
}

func renderInviteEmail(inviter string) error {
    var buf bytes.Buffer

    // Type-safe render
    err := RenderMailInviteContent(&buf, MailInviteContent{
        InviterName: inviter,
        SiteName:    "MyApp",
    })

    // Send email...
    return err
}
```

## Naming Conventions

### Directory Naming

**Pattern:** `lowercase_with_underscores`

```
templates/
├── mail_invite/              ✅ Good
├── dashboard_summary/        ✅ Good
├── user_profile/             ✅ Good
├── MailInvite/               ❌ Avoid (PascalCase)
├── mail-invite/              ❌ Avoid (hyphens)
```

### Numeric Prefixes

Use numeric prefixes for ordering (they're removed from generated names):

```
templates/
├── 01_mail_invite/
│   └── title.tmpl
├── 02_mail_welcome/
│   └── title.tmpl
└── 03_mail_password_reset/
    └── title.tmpl
```

**Generated names:**
- `MailInvite.Title` (01_ removed)
- `MailWelcome.Title` (02_ removed)
- `MailPasswordReset.Title` (03_ removed)

### Template File Naming

Use simple, descriptive names within each group:

```
mail_invite/
├── title.tmpl        ✅ Simple and clear
├── content.tmpl      ✅ Simple and clear
├── html.tmpl         ✅ Simple and clear
```

Avoid redundancy (the group name already provides context):

```
mail_invite/
├── mail_invite_title.tmpl      ❌ Redundant
├── invite_email_content.tmpl   ❌ Redundant
```

## Best Practices

### ✅ DO

**Group related templates:**
```
templates/
├── mail_invite/
│   ├── title.tmpl
│   ├── content.tmpl
│   └── footer.tmpl
```

**Use consistent naming:**
```
templates/
├── mail_invite/
├── mail_welcome/
└── mail_reset_password/
```

**Mix flat and grouped as needed:**
```
templates/
├── shared_header.tmpl      # Used by all pages
├── shared_footer.tmpl      # Used by all pages
└── mail_invite/            # Specific to invites
    └── content.tmpl
```

**Use numeric prefixes for ordering:**
```
templates/
├── 01_header/
├── 02_nav/
└── 03_footer/
```

### ❌ DON'T

**Don't nest deeper than 1 level:**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl    ❌ Not scanned (depth 2)
```

**Don't mix naming styles:**
```
templates/
├── mail_invite/         # underscore
├── mail-welcome/        # hyphen
└── MailResetPassword/   # PascalCase
```

**Don't use redundant names:**
```
mail_invite/
├── mail_invite_title.tmpl      ❌ Redundant
└── mail_invite_content.tmpl    ❌ Redundant
```

## Complete Example

### Directory Structure

```
myproject/
├── gen.go
├── main.go
├── template_gen.go (generated)
└── templates/
    ├── footer.tmpl
    ├── 01_mail_invite/
    │   ├── title.tmpl
    │   └── content.tmpl
    ├── 02_mail_account_created/
    │   ├── title.tmpl
    │   └── content.tmpl
    └── 03_mail_article_created/
        ├── title.tmpl
        └── content.tmpl
```

### Generate Code

**gen.go:**
```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

Run:
```bash
go generate
```

### Use the Generated Code

**main.go:**
```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    var buf bytes.Buffer

    // Type-safe rendering
    _ = RenderMailInviteTitle(&buf, MailInviteTitle{
        SiteName:    "MyApp",
        InviterName: "Alice",
    })
    fmt.Println("Invite Title:", buf.String())

    buf.Reset()

    _ = RenderMailInviteContent(&buf, MailInviteContent{
        InviteURL:   "https://myapp.com/invite/abc123",
        InviterName: "Alice",
        SiteName:    "MyApp",
    })
    fmt.Println("Invite Content:", buf.String())

    buf.Reset()

    // Dynamic rendering
    _ = Render(&buf, Template.MailAccountCreated.Title, MailAccountCreatedTitle{
        SiteName: "MyApp",
        UserName: "Bob",
    })
    fmt.Println("Account Created Title:", buf.String())

    // Flat template
    buf.Reset()
    _ = RenderFooter(&buf, Footer{
        Company: "MyCompany",
        Year:    2025,
    })
    fmt.Println("Footer:", buf.String())
}
```

### See the Complete Example

Run the working example:

```bash
cd examples/07_grouping
go generate
go run .
```

This example demonstrates:
- Mixing flat and grouped templates
- Numeric prefixes for ordering
- Type-safe and dynamic rendering
- Nested namespace structure

## See Also

- [Getting Started Guide](getting-started.md) - Tutorial and basics
- [CLI Reference](cli-reference.md) - Command-line options (directory scanning)
- [Template Syntax](template-syntax.md) - Supported template constructs
- [Example 03: Multi Template](../examples/03_multi_template/) - Multiple flat templates
- [Example 07: Grouping](../examples/07_grouping/) - Complete grouping example
