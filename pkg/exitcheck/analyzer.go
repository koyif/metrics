package exitcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is the static analyzer that checks for:
// - usage of panic function
// - calls to log.Fatal outside main function in main package
// - calls to os.Exit outside main function in main package
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "checks for panic, log.Fatal and os.Exit usage outside main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Check for function calls
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check for panic calls
			if isBuiltinPanic(callExpr) {
				pass.Reportf(callExpr.Pos(), "usage of panic is not allowed")
				return true
			}

			// Check if we're inside main function of main package
			inMainFunc := isInMainFunc(file, callExpr)
			isMainPkg := pass.Pkg.Name() == "main"

			// Check for log.Fatal calls
			if isLogFatal(callExpr, pass.TypesInfo) {
				if !isMainPkg || !inMainFunc {
					pass.Reportf(callExpr.Pos(), "log.Fatal should only be called in main function of main package")
				}
				return true
			}

			// Check for os.Exit calls
			if isOsExit(callExpr, pass.TypesInfo) {
				if !isMainPkg || !inMainFunc {
					pass.Reportf(callExpr.Pos(), "os.Exit should only be called in main function of main package")
				}
				return true
			}

			return true
		})
	}

	return nil, nil
}

// isBuiltinPanic checks if the call expression is a call to the built-in panic function
func isBuiltinPanic(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "panic"
}

// isLogFatal checks if the call expression is a call to log.Fatal
func isLogFatal(call *ast.CallExpr, info *types.Info) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if !isInPackage("log", ident, info) {
		return false
	}

	return sel.Sel.Name == "Fatal"
}

// isOsExit checks if the call expression is a call to os.Exit
func isOsExit(call *ast.CallExpr, info *types.Info) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if !isInPackage("os", ident, info) {
		return false
	}

	return sel.Sel.Name == "Exit"
}

func isInPackage(packagePath string, ident *ast.Ident, info *types.Info) bool {
	obj := info.Uses[ident]
	if obj == nil {
		return false
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}

	return pkgName.Imported().Path() == packagePath
}

// isInMainFunc checks if the call expression is inside the main function
func isInMainFunc(file *ast.File, call *ast.CallExpr) bool {
	var inMain bool

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if this is the main function
		if funcDecl.Name.Name != "main" {
			return true
		}

		// Check if the call is within this function's body
		if funcDecl.Body != nil {
			start := funcDecl.Body.Pos()
			end := funcDecl.Body.End()
			callPos := call.Pos()

			if callPos >= start && callPos <= end {
				inMain = true
				return false
			}
		}

		return true
	})

	return inMain
}
