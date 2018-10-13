package impl

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"strings"
)

const (
	FieldBaseUrl = "baseUrl"
	FieldHeader  = "header"
	FieldCookies = "cookies"
)

const (
	// packages
	HttpPkg = "net/http"
)

const (
	// types
	IntToken    = "int"
	StringToken = "string"
	ErrorToken  = "error"
)

const (
	// ids
	IdResp       = "resp"
	IdReq        = "req"
	IdStatusCode = "statusCode"
	IdError      = "err"
)

func Impl(service *Service, pkg string) (code string, err error) {
	service.newFunc = "New" + service.name
	service.implName = strings.ToLower(service.name) + "Impl"
	service.self = strings.ToLower(service.name)
	service.pkg = pkg

	file := NewFilePath(pkg)
	service.resolveCode(file)
	err = service.resolveMetadata()
	if err == nil {
		code = fmt.Sprintf("%#v", file)
	}
	return
}

// *net/http.Response -> Op("*").Qual("net/http", "Response")
func getQual(typ string) *Statement {
	switch typ {
	case IntToken:
		return Int()
	case StringToken:
		return String()
	case ErrorToken:
		return Err()
	default:
		var statement *Statement
		if strings.HasPrefix(typ, "*") {
			statement = Op("*")
			typ = typ[1:]
		}
		var pkg string
		dot := strings.LastIndex(typ, ".")
		if dot != -1 {
			pkg = typ[:dot]
			typ = typ[dot+1:]
		}

		//fmt.Printf("pkg: %s; typ: %s\n", pkg, typ)
		qual := Qual(pkg, typ)
		if statement == nil {
			statement = qual
		} else {
			statement = statement.Add(qual)
		}
		return statement
	}
}
