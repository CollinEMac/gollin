package main

import (
    "fmt"
    "os"
    "log"
    "strings"
    // "go/parser"
    "go/token"
    "go/ast"
    "go/types"
	"path/filepath"
	"golang.org/x/tools/go/packages"
)

func main() {
    // reads the contents of gollin at the given path and generates a .go file

    if len(os.Args) < 2 {
        log.Fatal("gollin file path required.")
        os.Exit(1)
    }
    path:= os.Args[1]

    var gollinPath string

    if strings.HasSuffix(path, ".gol") {
       gollinPath = path
    } else {
        // Build the full filepath
        // Should probably check for extensions before doing this
        var gollinBuilder strings.Builder
        gollinBuilder.WriteString(path)
        gollinBuilder.WriteString(".gol")
        gollinPath = gollinBuilder.String()
    }

    // Manipulate the code here
    goCode := parse(gollinPath);

    // we know gollinPath has suffix .gol at this point
    newFilePath := strings.Split(gollinPath, ".")[0]
    var goPath strings.Builder
    goPath.WriteString(newFilePath)
    goPath.WriteString(".go")

    // Spit out the go code
    os.WriteFile(goPath.String(), goCode, 0777);
}

func parse(gollinPath string) []byte {
    // Get gollin code
    gollinCode, err := os.ReadFile(gollinPath)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Begin parsing");
    fset := token.NewFileSet()
    // file, err := parser.ParseFile(fset, gollinPath, gollinCode, 0)
    if err != nil {
        log.Fatal("Parse error: ", err)
    }

	// Get the project directory from the filename
	dir := filepath.Dir(gollinPath)

	// Get package info for this file's dependencies
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  dir,
		Fset: fset,
	}

	// Load the current package and all its dependencies
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		log.Fatal("Failed to load package:", err)
	}

	if len(pkgs) == 0 {
		log.Fatal("No packages found")
	}

	pkg := pkgs[0]

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for function calls (CallExpr) in the Node
			if call, ok := n.(*ast.CallExpr); ok {

				if _, ok := call.Fun.(*ast.SelectorExpr); !ok {
					// This is not a package function like "os.Open()"
					// TODO: Handle *ast.Ident functions that return an error
					return true
				}

				// Look up our list of functions that return errors
				if returnsError(call, pkg.TypesInfo) {
					funcName := getFunctionName(call)
					fmt.Printf("Found error-returning call: %s at line %d\n",
						funcName, fset.Position(call.Pos()).Line)
					// TODO: Check if all errors are behing handled (try/catch)
				}
			}
			return true
		})
	}

    // For now, just return the original code
    // Later we'll transform it
    return []byte(gollinCode)
}

func getFunctionName(call *ast.CallExpr) string {
	// Package function: os.Open()
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok {
			return pkg.Name + "." + sel.Sel.Name
		}
	}
	return ""
}

func returnsError(call *ast.CallExpr, info *types.Info) bool {
    // Get the type of the function being called

    funcType, ok := info.Types[call.Fun]
    if !ok {
		fmt.Printf("error 1\n")
        return false
    }

    sig, ok := funcType.Type.(*types.Signature)
    if !ok {
		fmt.Printf("error 2")
        return false
    }

    // Get the results (return values)
    results := sig.Results()
    if results == nil || results.Len() == 0 {
        return false
    }

	// Return true if the very last result is an error type
    lastResult := results.At(results.Len() - 1)
    return lastResult.Type().String() == "error"
}
