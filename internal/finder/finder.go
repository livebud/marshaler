package finder

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func New(dir string) *Finder {
	return &Finder{
		dir: dir,
	}
}

type Finder struct {
	dir string
}

func (f *Finder) Find(importPath string, name string) (ast.Expr, error) {
	gomod, err := os.ReadFile(filepath.Join(f.dir, "go.mod"))
	if err != nil {
		return nil, err
	}
	modFile, err := modfile.Parse("go.mod", gomod, nil)
	if err != nil {
		return nil, err
	}
	modPath := modFile.Module.Mod.Path
	// If the import path has the modPath prefix, then it's a local import
	importPackage := f.importLocal
	if !strings.HasPrefix(importPath, modPath) {
		importPackage = f.importRemote
	}
	// Import the package
	pkg, err := importPackage(modFile, importPath, name)
	if err != nil {
		return nil, err
	}
	// Parse each valid Go file
	fset := token.NewFileSet()
	for _, filename := range pkg.GoFiles {
		filename = filepath.Join(f.dir, filename)
		code, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(fset, filename, code, parser.DeclarationErrors)
		if err != nil {
			return nil, err
		}
		// Look for the type spec
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if ts.Name.Name == name {
							return ts.Type, nil
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("finder:could not find type definition for %q.%s", importPath, name)
}

func (f *Finder) importLocal(modFile *modfile.File, importPath string, name string) (*build.Package, error) {
	dir := filepath.Join(f.dir, trimModulePath(modFile.Module.Mod.Path, importPath))
	return build.Import(".", dir, build.ImportMode(0))
}

func (f *Finder) importRemote(modFile *modfile.File, importPath string, name string) (*build.Package, error) {
	return nil, fmt.Errorf("find remote for %q.%s not implemneted yet", importPath, name)
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

func trimModulePath(modulePath string, importPath string) string {
	if modulePath == importPath {
		return ""
	}
	return strings.TrimPrefix(importPath, modulePath+"/")
}
