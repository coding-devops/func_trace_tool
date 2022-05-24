package ast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

type instrumenter struct {
	traceImport string
	tracePkg    string
	traceFunc   string
}

func New(traceImport, tracePkg, traceFunc string) *instrumenter {
	return &instrumenter{
		traceImport: traceImport,
		tracePkg:    tracePkg,
		traceFunc:   traceFunc,
	}
}

func hasFuncDecl(f *ast.File) bool {
	if len(f.Decls) == 0 {
		return false
	}

	for _, decl := range f.Decls {
		_, ok := decl.(*ast.FuncDecl)
		if ok {
			return true
		}
	}

	return false
}

func (a instrumenter) Instrument(filename string) ([]byte, error) {

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("dasd")
	}

	// 如果整个源文件都不包含函数的话就直接返回，无法注入
	if !hasFuncDecl(f) {
		return nil, errors.New("no func declare")
	}
	// ast.Print(fset, f)
	// // 先在ast上添加包导入
	astutil.AddNamedImport(fset, f, "trace", a.traceImport)
	//再添加函数声明
	a.addDeferTraceIntoFuncDecls(f)
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// ast.Print(fset, f)

	buf := &bytes.Buffer{}

	err = format.Node(buf, fset, f)

	if err != nil {
		return nil, fmt.Errorf("error formatting new code: %w", err)
	}

	return buf.Bytes(), nil // 返回转换后的Go源码
}

func (a instrumenter) addDeferTraceIntoFuncDecls(f *ast.File) {
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			a.addFunDDeferStmt(fd)
		}
	}
}

func (a instrumenter) addFunDDeferStmt(fd *ast.FuncDecl) {
	stmts := fd.Body.List

	for _, stmt := range stmts {
		ds, qw := stmt.(*ast.DeferStmt)
		if !qw {
			continue
		}
		ce, ok := ds.Call.Fun.(*ast.CallExpr)
		if !ok {
			continue
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		//这里找到了defer函数 现在确定这个defer函数是不是我们的trace.Trace()()
		ident, ok := se.X.(*ast.Ident)
		if ident.Name == a.tracePkg && se.Sel.Name == a.traceFunc && ok {
			return
		}
	}

	ds := &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: a.tracePkg,
					},
					Sel: &ast.Ident{
						Name: a.traceFunc,
					},
				},
			},
		},
	}

	newList := make([]ast.Stmt, len(stmts)+1)
	copy(newList[1:], stmts)
	newList[0] = ds
	fd.Body.List = newList
}
