package filter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"wsf/registry"

	"github.com/pkg/errors"
)

var (
	rootPath       = "/"
	rootImportPath = "/"
)

// Interface is a filter interface
type Interface interface {
	Filter(value interface{}) (interface{}, error)
	Defaults() error
}

// Register registers a filter
func Register(filterName string, filterType reflect.Type) {
	ft := registry.Get("filterTypes")
	filters := make(map[string]reflect.Type)
	if ft != nil {
		filters = ft.(map[string]reflect.Type)
	}

	filters[filterName] = filterType
	registry.Set("filterTypes", filters)
}

func init() {
	//_, filename, _, _ := runtime.Caller(0)
	//filename = filepath.Dir(filename)
	//fsWalk(filename, filename, processPath)
}

func fsWalk(fname string, linkName string, walkFn filepath.WalkFunc) error {
	fsWalkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var name string
		name, err = filepath.Rel(fname, path)
		if err != nil {
			return err
		}

		path = filepath.Join(linkName, name)

		if err == nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
			var symlinkPath string
			symlinkPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			info, err = os.Lstat(symlinkPath)

			if err != nil {
				return walkFn(path, info, err)
			}

			if info.IsDir() {
				return fsWalk(symlinkPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}

	err := filepath.Walk(fname, fsWalkFunc)
	return err
}

func processPath(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("Error scanning source: %s", err)
	}

	if !info.IsDir() || info.Name() == "tmp" {
		return nil
	}

	// Get the import path of the package
	pkgImportPath := rootImportPath
	fmt.Println("pkgImportPath:", pkgImportPath)
	if rootPath != path {
		pkgImportPath = rootImportPath + "/" + filepath.ToSlash(path[len(rootPath)+1:])
	}

	// Parse files within the path
	var pkgs map[string]*ast.Package
	fset := token.NewFileSet()
	pkgs, err = parser.ParseDir(
		fset,
		path,
		func(f os.FileInfo) bool {
			return !f.IsDir() && !strings.HasPrefix(f.Name(), ".") && strings.HasSuffix(f.Name(), ".go")
		},
		0)

	if err != nil {
		ast.Print(nil, err)
	}

	// Skip "main" packages
	delete(pkgs, "main")

	// Ignore packages that end with _test
	// These cannot be included in source code that is not generated specifically as a test
	for i := range pkgs {
		if len(i) > 6 {
			if string(i[len(i)-5:]) == "_test" {
				delete(pkgs, i)
			}
		}
	}

	// If there is no code in this directory, skip it.
	if len(pkgs) == 0 {
		return nil
	}

	// There should be only one package in this directory.
	if len(pkgs) > 1 {
		for i := range pkgs {
			println("Found package ", i)
		}
	}

	var pkg *ast.Package
	for _, v := range pkgs {
		pkg = v
	}

	if pkg != nil {
		processPackage(fset, pkgImportPath, path, pkg)
	}

	return nil
}

func processPackage(fset *token.FileSet, pkgImportPath, pkgPath string, pkg *ast.Package) {
	for fname, file := range pkg.Files {
		fmt.Println("fname:", fname)
		//imports := map[string]reflect.Type{}

		// For each declaration in the source file...
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok != token.TYPE {
				continue
			}

			spec := genDecl.Specs[0].(*ast.TypeSpec)
			_, ok = spec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			//imports[spec.Name.Name]
			fmt.Println("spec:", spec.Type)
			//fmt.Println("spec:", reflect.TypeOf((spec.Type)(nil)).Elem())
			/*addImports(imports, decl, pkgPath)

			if scanControllers {
				// Match and add both structs and methods
				structSpecs = appendStruct(fname, structSpecs, pkgImportPath, pkg, decl, imports, fset)
				appendAction(fset, methodSpecs, decl, pkgImportPath, pkg.Name, imports)
			} else if scanTests {
				structSpecs = appendStruct(fname, structSpecs, pkgImportPath, pkg, decl, imports, fset)
			}

			// If this is a func... (ignore nil for external (non-Go) function)
			if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Body != nil {
				// Scan it for validation calls
				lineKeys := GetValidationKeys(fname, fset, funcDecl, imports)
				if len(lineKeys) > 0 {
					validationKeys[pkgImportPath+"."+getFuncName(funcDecl)] = lineKeys
				}

				// Check if it's an init function.
				if funcDecl.Name.Name == "init" {
					initImportPaths = []string{pkgImportPath}
				}
			}*/
		}
	}
}
