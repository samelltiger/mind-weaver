package utils

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// GoCodeParser 专门用于解析 Go 代码
type GoCodeParser struct {
	fset *token.FileSet
}

// NewGoCodeParser 创建新的 Go 代码解析器
func NewGoCodeParser() *GoCodeParser {
	return &GoCodeParser{
		fset: token.NewFileSet(),
	}
}

// ParseFile 解析整个 Go 文件
func (p *GoCodeParser) ParseFile(filename string) (*GoFile, error) {
	file, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return p.parse(file), nil
}

// ParseSource 解析 Go 源代码字符串
func (p *GoCodeParser) ParseSource(src string) (*GoFile, error) {
	file, err := parser.ParseFile(p.fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return p.parse(file), nil
}

// GoFile 表示解析后的 Go 文件结构
type GoFile struct {
	Package    string      `json:"package"`
	Imports    []Import    `json:"imports"`
	Functions  []Function  `json:"functions"`
	Methods    []Method    `json:"methods"`
	Structs    []Struct    `json:"structs"`
	Interfaces []Interface `json:"interfaces"`
	Variables  []Variable  `json:"variables"`
	Constants  []Constant  `json:"constants"`
}

// Import 表示导入声明
type Import struct {
	Name string `json:"name"` // 别名（如有）
	Path string `json:"path"` // 导入路径
}

// Function 表示函数声明
type Function struct {
	Name       string   `json:"name"`
	Parameters []Field  `json:"parameters"`
	Results    []Field  `json:"results"`
	Doc        string   `json:"doc"`
	Position   Position `json:"position"`
}

// Method 表示方法声明
type Method struct {
	Receiver Field    `json:"receiver"`
	Function Function `json:"function"`
}

// Struct 表示结构体类型
type Struct struct {
	Name     string   `json:"name"`
	Fields   []Field  `json:"fields"`
	Doc      string   `json:"doc"`
	Position Position `json:"position"`
}

// Interface 表示接口类型
type Interface struct {
	Name     string     `json:"name"`
	Methods  []Function `json:"methods"`
	Doc      string     `json:"doc"`
	Position Position   `json:"position"`
}

// Field 表示结构体字段或接口方法
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

// Variable 表示变量声明
type Variable struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Value    string   `json:"value,omitempty"`
	Position Position `json:"position"`
}

// Constant 表示常量声明
type Constant struct {
	Name     string   `json:"name"`
	Type     string   `json:"type,omitempty"`
	Value    string   `json:"value"`
	Position Position `json:"position"`
	Doc      string   `json:"doc"`
}

// Param 表示函数参数或返回值
type Param struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

// Position 表示代码位置信息
type Position struct {
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
}

// parse 将 ast.File 转换为我们的 GoFile 结构
func (p *GoCodeParser) parse(file *ast.File) *GoFile {
	result := &GoFile{
		Package: file.Name.Name,
	}

	// 处理导入
	for _, imp := range file.Imports {
		importDecl := Import{
			Path: strings.Trim(imp.Path.Value, `"`),
		}
		if imp.Name != nil {
			importDecl.Name = imp.Name.Name
		}
		result.Imports = append(result.Imports, importDecl)
	}

	// 遍历所有声明
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			p.parseFuncDecl(d, result)
		case *ast.GenDecl:
			p.parseGenDecl(d, result)
		}
	}

	return result
}

// parseFuncDecl 处理函数和方法声明
func (p *GoCodeParser) parseFuncDecl(fn *ast.FuncDecl, file *GoFile) {
	funcDecl := Function{
		Name:       fn.Name.Name,
		Parameters: p.parseFieldList(fn.Type.Params),
		Results:    p.parseFieldList(fn.Type.Results),
		Doc:        fn.Doc.Text(),
		Position:   p.getPosition(fn.Pos(), fn.End()),
	}

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		// 这是方法
		receiver := p.parseField(fn.Recv.List[0])
		file.Methods = append(file.Methods, Method{
			Receiver: receiver,
			Function: funcDecl,
		})
	} else {
		// 这是普通函数
		file.Functions = append(file.Functions, funcDecl)
	}
}

// parseGenDecl 处理通用声明（类型、变量、常量等）
func (p *GoCodeParser) parseGenDecl(decl *ast.GenDecl, file *GoFile) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			p.parseTypeSpec(s, decl.Doc, file)
		case *ast.ValueSpec:
			p.parseValueSpec(s, decl, file)
		}
	}
}

// parseTypeSpec 处理类型声明
func (p *GoCodeParser) parseTypeSpec(spec *ast.TypeSpec, doc *ast.CommentGroup, file *GoFile) {
	pos := p.getPosition(spec.Pos(), spec.End())

	switch t := spec.Type.(type) {
	case *ast.StructType:
		file.Structs = append(file.Structs, Struct{
			Name:     spec.Name.Name,
			Fields:   p.parseStructFields(t.Fields),
			Doc:      doc.Text(),
			Position: pos,
		})
	case *ast.InterfaceType:
		file.Interfaces = append(file.Interfaces, Interface{
			Name:     spec.Name.Name,
			Methods:  p.parseInterfaceMethods(t.Methods),
			Doc:      doc.Text(),
			Position: pos,
		})
	}
}

// parseValueSpec 处理变量和常量声明
func (p *GoCodeParser) parseValueSpec(spec *ast.ValueSpec, decl *ast.GenDecl, file *GoFile) {
	varType := ""
	if spec.Type != nil {
		varType = p.exprToString(spec.Type)
	}

	doc := decl.Doc.Text()
	tok := decl.Tok.String()

	for i, name := range spec.Names {
		pos := p.getPosition(name.Pos(), name.End())
		value := ""
		if len(spec.Values) > i {
			value = p.exprToString(spec.Values[i])
		}

		switch tok {
		case "var":
			file.Variables = append(file.Variables, Variable{
				Name:     name.Name,
				Type:     varType,
				Value:    value,
				Position: pos,
			})
		case "const":
			file.Constants = append(file.Constants, Constant{
				Name:     name.Name,
				Type:     varType,
				Value:    value,
				Position: pos,
				Doc:      doc,
			})
		}
	}
}

// parseStructFields 解析结构体字段
func (p *GoCodeParser) parseStructFields(fields *ast.FieldList) []Field {
	var result []Field
	if fields == nil {
		return result
	}

	for _, f := range fields.List {
		field := p.parseField(f)
		result = append(result, field)
	}
	return result
}

// parseInterfaceMethods 解析接口方法
func (p *GoCodeParser) parseInterfaceMethods(methods *ast.FieldList) []Function {
	var result []Function
	if methods == nil {
		return result
	}

	for _, m := range methods.List {
		if len(m.Names) == 0 {
			continue // 可能是嵌入接口
		}

		if fnType, ok := m.Type.(*ast.FuncType); ok {
			result = append(result, Function{
				Name:       m.Names[0].Name,
				Parameters: p.parseFieldList(fnType.Params),
				Results:    p.parseFieldList(fnType.Results),
			})
		}
	}
	return result
}

// parseFieldList 解析参数或返回值列表
func (p *GoCodeParser) parseFieldList(fl *ast.FieldList) []Field {
	var params []Field
	if fl == nil {
		return params
	}

	for _, f := range fl.List {
		param := p.parseField(f)
		params = append(params, param)
	}
	return params
}

// parseField 解析单个字段/参数
func (p *GoCodeParser) parseField(f *ast.Field) Field {
	param := Field{
		Type: p.exprToString(f.Type),
	}
	if len(f.Names) > 0 {
		param.Name = f.Names[0].Name
	}
	return param
}

// exprToString 将表达式转换为字符串表示
func (p *GoCodeParser) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return p.exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + p.exprToString(e.Key) + "]" + p.exprToString(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return p.funcTypeToString(e)
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return "chan<- " + p.exprToString(e.Value)
		case ast.RECV:
			return "<-chan " + p.exprToString(e.Value)
		default:
			return "chan " + p.exprToString(e.Value)
		}
	case *ast.Ellipsis:
		return "..." + p.exprToString(e.Elt)
	case *ast.BasicLit:
		return e.Value
	default:
		return "<complex_type>"
	}
}

// funcTypeToString 将函数类型转换为字符串
func (p *GoCodeParser) funcTypeToString(ft *ast.FuncType) string {
	var params, results []string

	for _, p := range p.parseFieldList(ft.Params) {
		params = append(params, p.Name+" "+p.Type)
	}

	for _, r := range p.parseFieldList(ft.Results) {
		results = append(results, r.Name+" "+r.Type)
	}

	sig := "func(" + strings.Join(params, ", ") + ")"
	if len(results) > 0 {
		sig += " (" + strings.Join(results, ", ") + ")"
	} else if len(results) == 1 {
		sig += " " + results[0]
	}

	return sig
}

// getPosition 获取代码位置信息
func (p *GoCodeParser) getPosition(start, end token.Pos) Position {
	startPos := p.fset.Position(start)
	endPos := p.fset.Position(end)
	return Position{
		StartLine: startPos.Line,
		EndLine:   endPos.Line,
	}
}
