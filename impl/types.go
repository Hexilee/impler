package impl

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
)

const (
	Src = `
package types

import (
	"io"
)

var (
	IOReader io.Reader
)
`
)

const (
	TypeIOReader = "IOReader"
)

func GetType(name string) types.Type {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "types.go", Src, 0)
	if err != nil {
		log.Fatal(err)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("impler/types", fset, []*ast.File{file}, nil)
	return pkg.Scope().Lookup(name).Type()
}
