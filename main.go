package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"impler/impl"
	"io/ioutil"
	"log"
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
		log.Fatal("$GOFILE and $GOPACKAGE cannot be empty")
	}
	serviceName := os.Args[1]
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, GoFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	cmap := ast.NewCommentMap(fset, file, file.Comments)
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(GoPkg, fset, []*ast.File{file}, nil)
	if err != nil {
		log.Fatal(err) // type error
	}

	rawService, ok := pkg.Scope().Lookup(serviceName).Type().Underlying().(*types.Interface)
	if !ok {
		log.Fatal(serviceName + " is not a interface")
	}

	service := impl.NewService(serviceName, rawService).InitComments(cmap)
	code := impl.Impl(service, GoPkg)
	implFileName := fmt.Sprintf("%s_impl.go", strings.ToLower(serviceName))
	ioutil.WriteFile(implFileName, []byte(code), 0644)
}
