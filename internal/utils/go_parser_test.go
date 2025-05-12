package utils

import (
	"fmt"
	"testing"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		validate func(*testing.T, *GoFile, error)
	}{
		{
			name: "basic package and imports",
			source: `
package main

import (
	"fmt"
	"strings"
)
`,
			validate: func(t *testing.T, file *GoFile, err error) {
				if err != nil {
					t.Fatalf("ParseSource failed: %v", err)
				}
				if file.Package != "main" {
					t.Errorf("Expected package 'main', got '%s'", file.Package)
				}
				if len(file.Imports) != 2 {
					t.Fatalf("Expected 2 imports, got %d", len(file.Imports))
				}
				expectedImports := []string{"fmt", "strings"}
				for i, imp := range file.Imports {
					if imp.Path != expectedImports[i] {
						t.Errorf("Import %d: expected '%s', got '%s'", i, expectedImports[i], imp.Path)
					}
				}
			},
		},
		{
			name: "struct with methods",
			source: `
package model

type User struct {
	Name string
	Age  int
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}
`,
			validate: func(t *testing.T, file *GoFile, err error) {
				if err != nil {
					t.Fatalf("ParseSource failed: %v", err)
				}
				if len(file.Structs) != 1 {
					t.Fatalf("Expected 1 struct, got %d", len(file.Structs))
				}
				structDef := file.Structs[0]
				if structDef.Name != "User" {
					t.Errorf("Expected struct 'User', got '%s'", structDef.Name)
				}
				if len(file.Methods) != 1 {
					t.Fatalf("Expected 1 method, got %d", len(file.Methods))
				}
				method := file.Methods[0]
				if method.Receiver.Type != "*User" {
					t.Errorf("Expected receiver '*User', got '%s'", method.Receiver.Type)
				}
				if method.Function.Name != "Greet" {
					t.Errorf("Expected method 'Greet', got '%s'", method.Function.Name)
				}
			},
		},
		{
			name: "interface and functions",
			source: `
package storage

type Repository interface {
	Get(id string) (interface{}, error)
	Save(data interface{}) error
}

func NewMemoryRepository() Repository {
	return &memoryRepo{}
}
`,
			validate: func(t *testing.T, file *GoFile, err error) {
				if err != nil {
					t.Fatalf("ParseSource failed: %v", err)
				}
				if len(file.Interfaces) != 1 {
					t.Fatalf("Expected 1 interface, got %d", len(file.Interfaces))
				}
				iface := file.Interfaces[0]
				if iface.Name != "Repository" {
					t.Errorf("Expected interface 'Repository', got '%s'", iface.Name)
				}
				if len(iface.Methods) != 2 {
					t.Errorf("Expected 2 methods in interface, got %d", len(iface.Methods))
				}
				if len(file.Functions) != 1 {
					t.Fatalf("Expected 1 function, got %d", len(file.Functions))
				}
				if file.Functions[0].Name != "NewMemoryRepository" {
					t.Errorf("Expected function 'NewMemoryRepository', got '%s'", file.Functions[0].Name)
				}
			},
		},
		{
			name: "variables and constants",
			source: `
package config

const (
	DefaultTimeout = 30
	MaxRetries     = 3
)

var (
	APIEndpoint = "https://api.example.com"
	DebugMode   = false
)
`,
			validate: func(t *testing.T, file *GoFile, err error) {
				if err != nil {
					t.Fatalf("ParseSource failed: %v", err)
				}
				if len(file.Constants) != 2 {
					t.Fatalf("Expected 2 constants, got %d", len(file.Constants))
				}
				if len(file.Variables) != 2 {
					t.Fatalf("Expected 2 variables, got %d", len(file.Variables))
				}
				if file.Variables[0].Name != "APIEndpoint" || file.Variables[0].Value != `"https://api.example.com"` {
					t.Errorf("Unexpected variable: %+v", file.Variables[0])
				}
				if file.Constants[1].Name != "MaxRetries" || file.Constants[1].Value != "3" {
					t.Errorf("Unexpected constant: %+v", file.Constants[1])
				}
			},
		},
		{
			name: "invalid syntax",
			source: `
package main

func broken() {
	missing closing brace
`,
			validate: func(t *testing.T, _ *GoFile, err error) {
				if err == nil {
					t.Error("Expected parse error for invalid syntax, got nil")
				}
			},
		},
	}

	parser := NewGoCodeParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseSource(tt.source)
			tt.validate(t, file, err)
		})
	}
}

func TestParseFile(t *testing.T) {
	// 注意：这个测试需要实际文件，可以在测试前创建临时文件
	t.Run("parse actual file", func(t *testing.T) {
		parser := NewGoCodeParser()
		_, err := parser.ParseFile("H:\\code\\codepilot\\internal\\api\\handlers.go") // 需要准备测试文件
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}
		// 可以添加更多具体断言
	})

	t.Run("non-existent file", func(t *testing.T) {
		parser := NewGoCodeParser()
		_, err := parser.ParseFile("H:\\code\\codepilot\\internal\\api\\handlers111.go")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

// 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	parser := NewGoCodeParser()

	t.Run("position calculation", func(t *testing.T) {
		src := `package main

func foo() {
}
`
		file, err := parser.ParseSource(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(file.Functions) != 1 {
			t.Fatal("Expected one function")
		}
		pos := file.Functions[0].Position
		if pos.StartLine != 3 || pos.EndLine != 4 {
			t.Errorf("Expected lines 3-4, got %d-%d", pos.StartLine, pos.EndLine)
		}
	})

	t.Run("complex type parsing", func(t *testing.T) {
		src := `package types

type Complex struct {
	Ch chan<- *[]string
	Fn func(int) (string, error)
}
`
		file, err := parser.ParseSource(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(file.Structs) != 1 {
			t.Fatal("Expected one struct")
		}
		fields := file.Structs[0].Fields
		fmt.Println(fields)
		if fields[0].Type != "chan<- *[]string" {
			t.Errorf("Unexpected type for field 0: %s", fields[0].Type)
		}
		if fields[1].Type != "func( int) ( string,  error)" {
			t.Errorf("Unexpected type for field 1: %s", fields[1].Type)
		}
	})
}
