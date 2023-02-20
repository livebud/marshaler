package json

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"strings"
	"text/template"
)

type Unmarshaler struct {
	// Import path we're generating code into
	TargetPath string
	// Find the type spec for the given import path and name
	Find func(importPath string, name string) (ast.Expr, error)
	// Add an import to the generated code
	Import func(path string) (name string, err error)
}

//go:embed unmarshaler.gotext
var unmarshalerTemplate string

var generator = template.Must(template.New("unmarshaler").Parse(unmarshalerTemplate))

var generatorImports = []string{
	"fmt",
	"github.com/livebud/marshaler/json/scanner",
	"bytes",
}

// TODO: allow the import path name to be customized
func (u *Unmarshaler) typeName(importPath, name string) (string, error) {
	if u.TargetPath == importPath {
		return name, nil
	}
	importName, err := u.Import(importPath)
	if err != nil {
		return "", err
	}
	return importName + "." + name, nil
}

func (u *Unmarshaler) Generate(importPath, name string) ([]byte, error) {
	expr, err := u.Find(importPath, name)
	if err != nil {
		return nil, err
	}
	schema, err := fromExpr(expr, "in")
	if err != nil {
		return nil, err
	}
	for _, importPath := range generatorImports {
		if _, err := u.Import(importPath); err != nil {
			return nil, err
		}
	}
	typeName, err := u.typeName(importPath, name)
	if err != nil {
		return nil, err
	}
	state := State{
		Schema: schema,
		Name:   typeName,
	}
	code := new(bytes.Buffer)
	if err := generator.Execute(code, state); err != nil {
		return nil, err
	}
	return code.Bytes(), nil
}

func fromExpr(x ast.Expr, target string) (Type, error) {
	switch x := x.(type) {
	case *ast.Ident:
		return fromIdent(x, target)
	case *ast.StructType:
		return fromStruct(x, target)
	case *ast.MapType:
		return fromMap(x, target)
	case *ast.ArrayType:
		return fromArray(x, target)
	case *ast.StarExpr:
		return fromStar(x, target)
	default:
		return nil, fmt.Errorf("fromExpr: %T not implemented", x)
	}
}

func fromStruct(s *ast.StructType, target string) (*Struct, error) {
	var fields []StructField
	for _, f := range s.Fields.List {
		dataType, err := fromExpr(f.Type, "&"+target+"."+f.Names[0].Name)
		if err != nil {
			return nil, err
		}
		fields = append(fields, StructField{
			Key:  f.Names[0].Name,
			Type: dataType,
		})
	}
	return &Struct{
		Fields: fields,
	}, nil
}

func fromIdent(i *ast.Ident, target string) (Type, error) {
	switch i.Name {
	case "string":
		return String{target}, nil
	case "int":
		return Int{target}, nil
	case "float64":
		return Float64{target}, nil
	case "bool":
		return Bool{target}, nil
	}
	return nil, fmt.Errorf("fromIdent: %q not implemented", i.Name)
}

func fromMap(m *ast.MapType, target string) (*Map, error) {
	keyType, err := fromExpr(m.Key, target)
	if err != nil {
		return nil, err
	}
	// Static target because it's defined in the template
	valueType, err := fromExpr(m.Value, "&val")
	if err != nil {
		return nil, err
	}
	return &Map{
		Key:   keyType,
		Value: valueType,
		// For maps, we pull the value out of the target first and you Go doesn't
		// support `val := &target["key"]`, so we do `val := target["key"]` and
		// then `&val` instead.
		target: strings.TrimPrefix(target, "&"),
	}, nil
}

func fromArray(a *ast.ArrayType, target string) (*Array, error) {
	// Static target because it's defined in the template
	dataType, err := fromExpr(a.Elt, "&val")
	if err != nil {
		return nil, err
	}
	return &Array{
		Elt: dataType,
		// For arrays, we pull the value out of the target first and you Go doesn't
		// support `&target := append(&target, val)`, so we do
		// `target := append(target, val)` instead.
		target: strings.TrimPrefix(target, "&"),
	}, nil
}

func fromStar(s *ast.StarExpr, target string) (*Star, error) {
	// Static target because it's defined in the template
	dataType, err := fromExpr(s.X, "val")
	if err != nil {
		return nil, err
	}
	return &Star{
		X:      dataType,
		target: strings.TrimPrefix(target, "&"),
	}, nil
}

type State struct {
	Schema Type
	Name   string
}

type Type interface {
	Type() string
	String() string
	Target() string
}

func (String) Type() string  { return "string" }
func (Int) Type() string     { return "int" }
func (Float64) Type() string { return "float64" }
func (Bool) Type() string    { return "bool" }
func (Struct) Type() string  { return "struct" }
func (Array) Type() string   { return "array" }
func (Map) Type() string     { return "map" }
func (Star) Type() string    { return "star" }

type String struct {
	target string
}

type Int struct {
	target string
}

type Float64 struct {
	target string
}

type Bool struct {
	target string
}

func (String) String() string  { return "string" }
func (Int) String() string     { return "int" }
func (Float64) String() string { return "float64" }
func (Bool) String() string    { return "bool" }

type Struct struct {
	Fields []StructField
	target string
}

func (s *Struct) String() string {
	out := new(strings.Builder)
	out.WriteString("struct {\n")
	for _, f := range s.Fields {
		out.WriteString(fmt.Sprintf("  %s %s\n", f.Key, f.Type.String()))
	}
	out.WriteString("}")
	return out.String()
}

type StructField struct {
	Key  string
	Type Type
}

type Map struct {
	Key    Type
	Value  Type
	target string
}

func (m Map) String() string {
	return fmt.Sprintf("map[%s]%s", m.Key.String(), m.Value.String())
}

type Array struct {
	Elt    Type
	target string
}

func (s Array) String() string {
	return fmt.Sprintf("[]%s", s.Elt.String())
}

type Star struct {
	X      Type
	target string
}

func (s Star) String() string {
	return fmt.Sprintf("*%s", s.X.String())
}

func (t String) Target() string  { return t.target }
func (t Int) Target() string     { return t.target }
func (t Float64) Target() string { return t.target }
func (t Bool) Target() string    { return t.target }
func (t Struct) Target() string  { return t.target }
func (t Array) Target() string   { return t.target }
func (t Map) Target() string     { return t.target }
func (t Star) Target() string    { return t.target }
