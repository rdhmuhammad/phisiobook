package tokenizer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"

	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"golang.org/x/exp/slices"
)

func Load() *Tokenizer {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, os.Getenv("GOFILE"), nil, parser.ParseComments)
	if err != nil {
		logger.Error(err)
		return nil
	}

	var result = Tokenizer{
		groups: make(map[string]group),
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Recv == nil {
			continue
		}

		for _, param := range fn.Type.Params.List {
			if starExpr, ok := param.Type.(*ast.StarExpr); ok {
				switch t := starExpr.X.(type) {
				case *ast.SelectorExpr:
					if pkg, ok := t.X.(*ast.Ident); ok {
						if fn.Name.Name == "Route" && pkg.Name == "gin" && t.Sel.Name == "RouterGroup" {
							parsingBody(fn, &result, param.Names[0].Name)
							break
						}
					}
				}
			}
		}
	}

	// request body
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Recv == nil {
			continue
		}

		for _, gr := range result.groups {
			rt, ok := gr.routers[fn.Name.Name]
			if !ok {
				continue
			}

			parsingRequestBody(fn, &result, rt)
		}
	}

	return &result
}

func parsingRequestBody(fn *ast.FuncDecl, result *Tokenizer, rt router) {
	for _, statement := range fn.Body.List {
		if declStmt, ok := statement.(*ast.DeclStmt); ok {
			gendcl, ok := declStmt.Decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gendcl.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if spec.Names[0].Name == "request" {
					expr, ok := spec.Type.(*ast.SelectorExpr)
					if !ok {
						if spec.Values == nil {
							continue
						}

						lit, ok := spec.Values[0].(*ast.CompositeLit)
						if !ok {
							continue
						}

						exp, ok := lit.Type.(*ast.SelectorExpr)
						if !ok {
							continue
						}

						var reqPaths []string
						pkg, ok := exp.X.(*ast.Ident)
						if ok {
							reqPaths = append(reqPaths, pkg.Name)
						}
						reqPaths = append(reqPaths, exp.Sel.Name)
						rt.requestBodyDir = strings.Join(reqPaths, ".")
						result.editRouter(rt)
						break
					}
					var reqPaths []string
					pkg, ok := expr.X.(*ast.Ident)
					if ok {
						reqPaths = append(reqPaths, pkg.Name)
					}

					reqPaths = append(reqPaths, expr.Sel.Name)
					rt.requestBodyDir = strings.Join(reqPaths, ".")
					result.editRouter(rt)
					break
				}
			}
		}
	}
}

func parsingBody(fn *ast.FuncDecl, result *Tokenizer, routerVarName string) {
	for _, statement := range fn.Body.List {
		if assign, ok := statement.(*ast.AssignStmt); ok {
			if gr, ok := parsingGroupAssign(assign, routerVarName); ok {
				result.addGroup(gr)
				continue
			}
		}

		if expr, ok := statement.(*ast.ExprStmt); ok {
			if call, ok := parsingRouteCall(expr); ok {
				result.addRoute(call.varGroupName, call)
			}
		}
	}
}

func parsingRouteCall(expr *ast.ExprStmt) (router, bool) {
	callExpr, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return router{}, false
	}

	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return router{}, false
	}

	var result router
	groupIdent, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return router{}, false
	}
	result.varGroupName = groupIdent.Name
	httpMethod := selectorExpr.Sel.Name
	if !isHTTPMethod(httpMethod) {
		return router{}, false
	}
	result.method = httpMethod
	var middlewareSel = []string{
		"Validate",
		"Authorize",
		"Idempotent",
	}
	for _, param := range callExpr.Args {
		if e, ok := stringLiteral(param); ok {
			result.path = e
			continue
		}

		if p, ok := param.(*ast.SelectorExpr); ok &&
			slices.Contains(middlewareSel, p.Sel.Name) {
			continue
		} else if ok {
			result.handlerFunc = p.Sel.Name
		}

	}

	logger.Infof("Group %s, 1 Endpoint added => %s %s", result.varGroupName, result.method, result.path)
	return result, true
}

func isHTTPMethod(s string) bool {
	switch s {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func parsingGroupAssign(assign *ast.AssignStmt, routerVar string) (group, bool) {
	// has left and right
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return group{}, false
	}

	// had variable
	lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return group{}, false
	}

	// exp not null
	callExpr, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return group{}, false
	}

	// using selector :=
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return group{}, false
	}

	if selector.Sel == nil || selector.Sel.Name != "Group" {
		return group{}, false
	}

	parentIdent, ok := selector.X.(*ast.Ident)
	if !ok {
		return group{}, false
	}

	if parentIdent.Name != routerVar {
		return group{}, false
	}

	if len(callExpr.Args) < 1 {
		return group{}, false
	}

	groupPath, ok := stringLiteral(callExpr.Args[0])
	if !ok {
		return group{}, false
	}

	logger.Infof("1 Group added called => %s", lhsIdent.Name)
	return group{path: groupPath, VarName: lhsIdent.Name, routers: make(map[string]router)}, true
}

func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return v, true
}

func isGinRouterGroupPtr(t types.Type) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}

	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Name() == "RouterGroup" &&
		obj.Pkg().Path() == "github.com/gin-gonic/gin"
}
