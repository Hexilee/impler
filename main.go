package main

import (
	"fmt"
	"github.com/rady-io/http-service/impl"
	. "github.com/rady-io/http-service/log"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"strings"
)

const (
	GoFileKey = "GOFILE"
	GoPkgKey  = "GOPACKAGE"
	ZeroStr   = ""
)

var (
	GoFile = os.Getenv(GoFileKey)
	GoPkg  = os.Getenv(GoPkgKey)
)

func main() {
	if GoFile == ZeroStr || GoPkg == ZeroStr {
		Log.Fatal("$GOFILE and $GOPACKAGE cannot be empty")
	}
	serviceName := os.Args[1]
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, GoFile, nil, parser.ParseComments)
	if err != nil {
		Log.Fatal(err.Error())
	}
	cmap := ast.NewCommentMap(fset, file, file.Comments)
	conf := types.Config{Importer: importer.Default()}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	_, err = conf.Check(GoPkg, fset, []*ast.File{file}, info)
	if err != nil {
		Log.Fatal(err.Error()) // type error
	}

	service := impl.NewService(serviceName, info).InitComments(cmap)
	code, err := impl.Impl(service, GoPkg)
	if err != nil {
		Log.Fatal(err.Error())
	}

	implFileName := fmt.Sprintf("%s_impl.go", strings.ToLower(serviceName))
	ioutil.WriteFile(implFileName, []byte(code), 0644)
}
