package impl

import (
	"fmt"
	. "github.com/Hexilee/impler/log"
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
	HttpPkg      = "net/http"
	EncodingJSON = "encoding/json"
	EncodingXML  = "encoding/xml"
	Bytes        = "bytes"
	Ioutil       = "io/ioutil"
	NetURL       = "net/url"
	IO           = "io"
	MultipartPkg = "mime/multipart"
	Textproto    = "net/textproto"
	OS           = "os"
	StringsPkg   = "strings"
	UnHTMLPkg    = "github.com/Hexilee/unhtml"
)

const (
	// types
	IntToken    = "int"
	StringToken = "string"
	ErrorToken  = "error"
)

func Impl(service *Service, pkg string) (code string, err error) {
	Log.Infof("Implement Service: %s", service.name)
	service.newFunc = "New" + service.name
	service.implName = strings.ToLower(service.name) + "Impl"
	service.self = strings.ToLower(service.name)
	service.pkg = pkg
	file := NewFilePath(pkg)
	err = service.resolveMetadata()
	if err == nil {
		service.resolveCode(file)
		code = fmt.Sprintf("%#v", file)
	}
	return
}

// *net/http.Response -> Op("*").Qual("net/http", "Response")
func getQual(typ string) *Statement {
	if !strings.Contains(typ, ".") {
		return Id(typ)
	}
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
