package defercheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/cfg"
)

const Doc = `check references to variables inside a defer`

var Analyzer = &analysis.Analyzer{
	Name: "defercheck",
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		ctrlflow.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		var fType *ast.FuncType
		var cfg *cfg.CFG
		switch n := n.(type) {
		case *ast.FuncDecl:
			fType = n.Type
			cfg = cfgs.FuncDecl(n)
		case *ast.FuncLit:
			fType = n.Type
			cfg = cfgs.FuncLit(n)
		}

		if cfg == nil {
			return
		}
		newVisitor(pass, fType).Analyze(cfg)
	})

	return nil, nil
}
