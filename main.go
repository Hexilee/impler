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
	pkg, err := conf.Check(GoPkg, fset, []*ast.File{file}, nil)
	if err != nil {
		Log.Fatal(err.Error()) // type error
	}

	rawService, ok := pkg.Scope().Lookup(serviceName).Type().Underlying().(*types.Interface)
	if !ok {
		Log.Fatal(serviceName + " is not a interface")
	}

	service := impl.NewService(serviceName, rawService).InitComments(cmap)
	code, err := impl.Impl(service, GoPkg)
	if err != nil {
		Log.Fatal(err.Error())
	}

	implFileName := fmt.Sprintf("%s_impl.go", strings.ToLower(serviceName))
	ioutil.WriteFile(implFileName, []byte(code), 0644)
}
