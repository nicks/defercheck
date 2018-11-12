package defercheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/cfg"
)

type visitor struct {
	pass  *analysis.Pass
	fType *ast.FuncType

	// Keep track of the captured variables when we visit each block.
	// If the captured variables change, we'll need to reanalyze.
	visitedBlocks map[int32]visitorState
}

func newVisitor(pass *analysis.Pass, fType *ast.FuncType) *visitor {
	return &visitor{
		pass:          pass,
		fType:         fType,
		visitedBlocks: make(map[int32]visitorState),
	}
}

func (v *visitor) Analyze(cfg *cfg.CFG) {
	if len(cfg.Blocks) == 0 {
		return
	}

	v.analyzeBlock(cfg.Blocks[0], visitorState{})
}

// Traverse the CFG, analyzing the blocks in order.
func (v *visitor) analyzeBlock(b *cfg.Block, state visitorState) {
	lastVisit, ok := v.visitedBlocks[b.Index]
	if ok && lastVisit.Equals(state) {
		// We already visited this block with the same visitorState, so
		// we can skip this.
		return
	}
	v.visitedBlocks[b.Index] = state

	for _, n := range b.Nodes {
		state = v.analyzeNode(n, state)
	}

	for _, succ := range b.Succs {
		v.analyzeBlock(succ, state)
	}
}

// Traverse a node. We want to check for two things:
// 1) If the node has a defer that evals vars, and
// 2) If the node assigns any vars
func (v *visitor) analyzeNode(n ast.Node, state visitorState) visitorState {
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit, *ast.FuncDecl:
			return false
		case *ast.DeferStmt:
			state = v.analyzeDefer(n, state)
			return false
		case *ast.ReturnStmt:
			if v.fType.Results == nil {
				return false
			}

			for _, retField := range v.fType.Results.List {
				if len(retField.Names) != 1 {
					continue
				}

				ident := retField.Names[0]
				obj, ok := v.pass.TypesInfo.Defs[ident]
				if !ok {
					continue
				}

				evalNode, evaled := state.deferEvaledVars.Use(obj)
				if !evaled {
					continue
				}

				v.pass.Reportf(evalNode.Pos(),
					"variable %s evaluated by defer, then returned later", ident.Name)
			}
			return false

		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				ident, isIdent := lhs.(*ast.Ident)
				if !isIdent {
					continue
				}

				obj, ok := v.pass.TypesInfo.Uses[ident]
				if !ok {
					continue
				}

				evalNode, evaled := state.deferEvaledVars.Use(obj)
				if !evaled {
					continue
				}

				v.pass.Reportf(evalNode.Pos(),
					"variable %s evaluated by defer, then reassigned later", ident.Name)
			}
			return false
		default:
			return true
		}
	})
	return state
}

func (v *visitor) analyzeDefer(n ast.Node, state visitorState) visitorState {
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit, *ast.FuncDecl:
			return false
		case *ast.Ident:
			obj, ok := v.pass.TypesInfo.Uses[n]
			if !ok {
				return false
			}

			state = visitorState{
				state.deferEvaledVars.Add(obj, n),
			}
			return false
		default:
			return true
		}
	})
	return state
}

type visitorState struct {
	deferEvaledVars varSet
}

func (a visitorState) Equals(b visitorState) bool {
	return a.deferEvaledVars.Equals(b.deferEvaledVars)
}

// Immutable list of vars captured in defers, keyed by Object ID
type varSet map[string]ast.Node

func (vs varSet) Contains(obj types.Object) bool {
	_, exists := vs[obj.Id()]
	return exists
}

func (vs varSet) Use(obj types.Object) (ast.Node, bool) {
	n, exists := vs[obj.Id()]
	return n, exists
}

func (vs varSet) Add(obj types.Object, n ast.Node) varSet {
	_, exists := vs[obj.Id()]
	if exists {
		return vs
	}

	newMap := make(map[string]ast.Node, len(vs)+1)
	for k, v := range vs {
		newMap[k] = v
	}
	newMap[obj.Id()] = n
	return varSet(newMap)
}

func (a varSet) Equals(b varSet) bool {
	if len(a) != len(b) {
		return false
	}

	// It's good enough to compare the keys
	// because they are IDs.
	for k := range a {
		_, exists := b[k]
		if !exists {
			return false
		}
	}
	return true
}
