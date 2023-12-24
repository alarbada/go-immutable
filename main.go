package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gookit/goutil/dump"
	"golang.org/x/tools/go/packages"
)

func main() {
	// dump.Std().MaxDepth = 10
	flag.Parse()
	args := flag.Args()
	pattern := args[0]

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		panic(err)
	}

	funcParams := make(map[string][]string)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					if x.Type.Params != nil {
						var params []string
						for _, p := range x.Type.Params.List {
							for _, n := range p.Names {
								params = append(params, n.Name)
							}
						}

						funcDeclId := getFuncDeclId(pkg, x)
						funcParams[funcDeclId] = params
					}
				}
				return true
			})
		}
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			fset := pkg.Fset

			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					ast.Inspect(x.Body, func(n ast.Node) bool {
						dump.P(n.Pos())
						return true
					})

				case *ast.CallExpr:
					id := constructIdentifier(x, pkg, file)
					params, ok := funcParams[id]
					if !ok {
						return true
					}

					for i, arg := range x.Args {
						currParam := params[i]
						if arg, ok := arg.(*ast.Ident); ok {
							if checkArgument(fset, currParam, arg) {
								return true
							}
						}
					}

				case *ast.AssignStmt:
					if len(x.Lhs) == 1 && len(x.Rhs) == 1 {
						if ident, ok := x.Lhs[0].(*ast.Ident); ok && !ident.IsExported() {
							// Check if it's a new variable declaration with :=
							if x.Tok == token.DEFINE {
								return true
							}
							if checkVariableName(fset, ident.Name, ident.Pos()) {
								return true
							}
						}
					}
					for _, lhs := range x.Lhs {
						switch v := lhs.(type) {
						case *ast.Ident: // Variable
							if checkVariableName(fset, v.Name, v.Pos()) {
								fmt.Println(lhs)
								return true
							}
						case *ast.SelectorExpr: // Struct field
							if ident, ok := v.X.(*ast.Ident); ok {
								if checkStructName(fset, ident.Name, v.Sel.Name, v.Sel.Pos()) {
									return true
								}
							}
						}
					}
				}
				return true
			})

		}
	}
}

func getFuncDeclId(pkg *packages.Package, x *ast.FuncDecl) string {
	funcDeclId := ""
	if x.Recv != nil {
		recvTypeName := x.Recv.List[0].Type.(*ast.Ident).Name
		funcDeclId = pkg.ID + "." + recvTypeName + "." + x.Name.Name
	} else {
		funcDeclId = pkg.ID + "." + x.Name.Name
	}

	return funcDeclId
}

func checkStructName(
	fset *token.FileSet,
	structName, fieldName string,
	pos token.Pos,
) (checked bool) {

	if !strings.HasPrefix(fieldName, "mut") && !strings.HasPrefix(fieldName, "Mut") {
		name := structName + "." + fieldName

		fmt.Println("checkStructName")
		fmt.Printf(
			"%s: Variable '%s' should be prefixed with 'mut' or 'Mut'\n",
			fset.Position(pos), name,
		)
		return true
	}

	return false
}

func checkVariableName(fset *token.FileSet, name string, pos token.Pos) (checked bool) {
	if name == "_" {
		return false
	}

	if !strings.HasPrefix(name, "mut") && !strings.HasPrefix(name, "Mut") {
		fmt.Println("checkVariableName")
		fmt.Printf(
			"%s: Variable '%s' should be prefixed with 'mut' or 'Mut'\n",
			fset.Position(pos), name,
		)
		return true
	}

	return false
}

func checkArgument(
	fset *token.FileSet,
	currParam string,
	arg *ast.Ident,
) (checked bool) {
	isCurrParamMut := strings.HasPrefix(currParam, "mut") || strings.HasPrefix(currParam, "Mut")

	isArgMut := strings.HasPrefix(arg.Name, "mut") || strings.HasPrefix(arg.Name, "Mut")

	if isCurrParamMut && !isArgMut {
		fmt.Println("checkArgument")
		fmt.Printf(
			"%s: Argument '%s' should be prefixed with 'mut' or 'Mut'\n",
			fset.Position(arg.Pos()), arg.Name,
		)
		return true
	}

	return false
}

func isPackageName(name string, file *ast.File) bool {
	for _, imp := range file.Imports {
		// Strip the double quotes from the import path
		importPath := strings.Trim(imp.Path.Value, "\"")

		// If the import has a name (alias), compare it with 'name'
		if imp.Name != nil {
			if imp.Name.Name == name {
				return true
			}
		} else {
			// If no alias is used, extract the actual package name from the path and compare it
			parts := strings.Split(importPath, "/")
			actualPkgName := parts[len(parts)-1]
			if actualPkgName == name {
				return true
			}
		}
	}
	return false
}

func constructIdentifier(
	callExpr *ast.CallExpr,
	currPkg *packages.Package,
	file *ast.File,
) string {
	currentPkgId := currPkg.ID

	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		// Local function call
		return currentPkgId + "." + fun.Name

	case *ast.SelectorExpr:
		switch x := fun.X.(type) {
		case *ast.Ident:
			if isPackageName(x.Name, file) {
				// It's a package function call
				return resolvePackagePath(x.Name, file) + "." + fun.Sel.Name
			} else {
				var packageName, typeName string
				valueSpec, ok := x.Obj.Decl.(*ast.ValueSpec)
				if !ok {
					return ""
				}

				switch valueSpecType := valueSpec.Type.(type) {

				case *ast.Ident: // local type from the same package
					packageName = currentPkgId
					typeName = valueSpecType.Name

				case *ast.SelectorExpr: // external type from another package
					selectorExpr := valueSpecType
					packageName = selectorExpr.X.(*ast.Ident).Name
					packageName = resolvePackagePath(packageName, file)
					typeName = selectorExpr.Sel.Name
				}

				return packageName + "." + typeName + "." + fun.Sel.Name
			}
		}
	}
	return ""
}

func resolvePackagePath(alias string, file *ast.File) string {
	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == alias {
			return strings.Trim(imp.Path.Value, "\"")
		}
		if imp.Name == nil {
			parts := strings.Split(strings.Trim(imp.Path.Value, "\""), "/")
			if parts[len(parts)-1] == alias {
				return strings.Trim(imp.Path.Value, "\"")
			}
		}
	}
	return ""
}

func resolvePackagePathForType(typeName string, file *ast.File) string {
	for _, imp := range file.Imports {
		// Check if the import alias matches the type name
		if imp.Name != nil && imp.Name.Name == typeName {
			return strings.Trim(imp.Path.Value, "\"")
		}

		// Check if the last element of the import path matches the type name
		impPath := strings.Trim(imp.Path.Value, "\"")
		parts := strings.Split(impPath, "/")
		if parts[len(parts)-1] == typeName {
			return impPath
		}
	}
	return ""
}

func handlePanic(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()
	fn()
	return nil
}

func processFunctionDecl(fDecl *ast.FuncDecl) {
	funcScope := make(map[string]bool)

	if fDecl.Type.Params != nil {
		for _, p := range fDecl.Type.Params.List {
			for _, n := range p.Names {
				funcScope[n.Name] = true
			}
		}
	}

	// Traverse the function body
	ast.Inspect(fDecl.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			// Handle assignments
			for _, lhs := range x.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					_, declared := funcScope[ident.Name]
					if x.Tok == token.DEFINE && !declared {
						// New declaration
						funcScope[ident.Name] = true
					}

				}
			}
		}
		return true
	})
}
