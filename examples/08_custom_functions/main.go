package main

import (
	"bytes"
	"fmt"
	"time"
)

func main() {
	// Initialize templates with custom functions
	InitTemplates(WithFuncs(GetTemplateFuncs()))

	fmt.Println("=== Example 1: Email Template ===")
	var buf1 bytes.Buffer
	err := RenderEmail(&buf1, Email{
		Title: "welcome email",
		User: EmailUser{
			Name:  "Alice Smith",
			Email: "ALICE@EXAMPLE.COM",
		},
		CreatedAt: time.Date(2025, 11, 16, 14, 30, 0, 0, time.UTC),
		Message:   "Welcome to our service!\nWe're glad to have you.",
		Price:     12345,
		URL:       "", // Empty string will use default value
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(buf1.String())

	// Test custom function with dynamic resolution
	fmt.Println("\n=== Example 2: Custom Functions Test ===")
	var buf2 bytes.Buffer
	err = RenderCustomFuncTest(&buf2, CustomFuncTest{
		Title:   "Hello World",
		Message: "Testing Custom Functions",
		URL:     "https://example.com",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(buf2.String())

	// Example without custom functions (should fail with clear error)
	fmt.Println("\n=== Without initialization ===")
	testWithoutInit()
}

func testWithoutInit() {
	// Reset for demonstration (in real code, don't do this)
	// This simulates what happens if InitTemplates is not called
	fmt.Println("Note: In a real application, forgetting to call InitTemplates() will result in a clear error message.")
}
