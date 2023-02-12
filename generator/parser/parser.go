package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
)

func Parse(fsys fs.FS, dir string) (*Package, error) {
	imported, err := importPackage(fsys, dir)
	if err != nil {
		return nil, err
	}
	// Parse each valid Go file
	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(imported.GoFiles))
	for _, filename := range imported.GoFiles {
		filename = path.Join(dir, filename)
		code, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(fset, filename, code, parser.DeclarationErrors)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return &Package{
		Name:  imported.Name,
		Files: files,
	}, nil
}

type Package struct {
	Name  string
	Files []*ast.File
}

func (p *Package) TypeSpec(name string) (*ast.TypeSpec, bool) {
	for _, file := range p.Files {
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if ts.Name.Name == name {
							return ts, true
						}
					}
				}
			}
		}
	}
	return nil, false
}
