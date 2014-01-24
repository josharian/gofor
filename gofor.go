// gofor analyzes the types of for loops encountered in Go code.
// It is geared towards detecting counting loops -- loops of the form
//   for i := min; i < max; i += stride
// and its categorizations are mostly steps on the way to ruling out
// non-counting loops and then detecting interesting types of counting
// loops -- special values for min and stride (0 and 1 respectively)
// and literal vs non-literal min/max/stride.
//
// It is intended to be used in conjunction with sort and uniq.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/josharian/gofor/github.com/kr/fs"
)

type visitor struct{}

func (v visitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return v
	}

	if _, ok := node.(*ast.RangeStmt); ok {
		fmt.Println("range")
		return v
	}

	// for _ := range _ {
	stmt, ok := node.(*ast.ForStmt)
	if !ok {
		return v
	}

	// for {
	if stmt.Init == nil && stmt.Cond == nil && stmt.Post == nil {
		fmt.Println("bare for")
		return v
	}

	// for scan.Scan() {
	if stmt.Init == nil && stmt.Post == nil {
		fmt.Println("cond only")
		return v
	}

	if stmt.Init == nil {
		fmt.Println("missing init")
		return v
	}

	if stmt.Post == nil {
		fmt.Println("missing post")
		return v
	}

	// condition not a < b or a <= b
	binExpr, ok := stmt.Cond.(*ast.BinaryExpr)
	if !ok || (binExpr.Op != token.LSS && binExpr.Op != token.LEQ) {
		fmt.Println("cond not < or <=")
		return v
	}

	// cond lhs not identifier
	condLhs, ok := binExpr.X.(*ast.Ident)
	if !ok {
		fmt.Println("cond lhs not identifier")
		return v
	}

	initAssign, ok := stmt.Init.(*ast.AssignStmt)
	// init not i := n
	if !ok {
		fmt.Println("init not i := n")
		return v
	}

	// init has multiple values on lhs/rhs
	if len(initAssign.Lhs) != 1 || len(initAssign.Rhs) != 1 {
		fmt.Println("init multiple values")
		return v
	}

	initLhs, ok := initAssign.Lhs[0].(*ast.Ident)
	// init lhs not an identifier?!
	if !ok {
		fmt.Println("init lhs not identifier")
		return v
	}

	if initLhs.Name != condLhs.Name {
		fmt.Println("init lhs != cond lhs")
		return v
	}

	postAssign, postAssignOk := stmt.Post.(*ast.AssignStmt)
	postIncDec, postIncDecOk := stmt.Post.(*ast.IncDecStmt)
	if !postAssignOk && !postIncDecOk {
		fmt.Println("post not assign or inc/dec")
		return v
	}

	if postAssignOk {
		if len(postAssign.Lhs) != 1 || len(postAssign.Rhs) != 1 {
			fmt.Println("post assign multiple values")
			return v
		}
		postAssignLhs, ok := postAssign.Lhs[0].(*ast.Ident)
		if !ok {
			fmt.Println("post assign lhs not ident")
			return v
		}
		if postAssignLhs.Name != initLhs.Name {
			fmt.Println("init lhs != post assign lhs")
			return v
		}
		if postAssign.Tok != token.ADD_ASSIGN && postAssign.Tok != token.SUB_ASSIGN {
			fmt.Println("post assign not += or -= (but might be i = i - 1, oh well)")
			return v
		}
	}

	if postIncDecOk {
		postIncDecLhs, ok := postIncDec.X.(*ast.Ident)
		if !ok {
			fmt.Println("post incdec lhs not ident")
			return v
		}
		if postIncDecLhs.Name != initLhs.Name {
			fmt.Println("init lhs != post incdec lhs")
			return v
		}
	}

	// Ok, at this point we know we have a statement of one of these three forms:
	// for i := ?; i <[=] ?; i += [?] {
	// for i := ?; i <[=] ?; i++ {
	// for i := ?; i <[=] ?; i-- {

	// For each of the question marks here -- min, max, stride -- there are multiple
	// cases we might want to distinguish. Figure out which it is.

	var min string
	if initRhsLit, ok := initAssign.Rhs[0].(*ast.BasicLit); ok {
		if initRhsLit.Value == "0" {
			min = "0"
		} else {
			min = "literal"
		}
	} else {
		min = "non-literal"
	}

	var max string
	if _, ok := binExpr.Y.(*ast.BasicLit); ok {
		max = "literal"
	} else {
		max = "non-literal"
	}

	var stride string
	var postAssignRhsLit *ast.BasicLit
	var postAssignRhsLitOk bool
	if postAssignOk {
		postAssignRhsLit, postAssignRhsLitOk = postAssign.Rhs[0].(*ast.BasicLit)
	}
	switch {
	case postIncDecOk && postIncDec.Tok == token.INC:
		stride = "1"
	case postIncDecOk: // must be DEC
		stride = "literal" // -1
	case !postAssignRhsLitOk:
		stride = "non-literal"
	case postAssignRhsLit.Value == "1" && postAssign.Tok == token.ADD_ASSIGN:
		stride = "1"
		// ignore possibility of code doing i -= -1. That's dumb.
	default:
		stride = "literal"
	}

	fmt.Printf("counting loop min %s, max %s, stride %s\n", min, max, stride)
	return v
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: gofor <dir> | sort | uniq -c | sort -n -r")
		os.Exit(1)
	}
	walker := fs.Walk(os.Args[1])

	var v visitor

	for walker.Step() {
		if err := walker.Err(); err != nil {
			fmt.Printf("Error during filesystem walk: %v\n", err)
			continue
		}

		if walker.Stat().IsDir() || !strings.HasSuffix(walker.Path(), ".go") {
			continue
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, walker.Path(), nil, 0)
		// fmt.Println(walker.Path())
		if err != nil {
			// don't print err here; it is too chatty, due to (un?)surprising
			// amounts of broken code in the wild
			continue
		}

		ast.Walk(v, f)
	}
}
