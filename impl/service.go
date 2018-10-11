package impl

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"strings"
)

const (
	TokenType = "type"
	ZeroStr   = ""
	LF        = "\n"
)

func NewService(name string, service *types.Interface) *Service {
	return &Service{
		methods: make(map[token.Pos]*Method),
		name:    name,
		service: service,
		meta: &ServiceMeta{
			header:  make(http.Header),
			cookies: make([]*http.Cookie, 0),
		},
	}
}

type (
	Service struct {
		methods     map[token.Pos]*Method
		name        string
		commentText string
		service     *types.Interface
		meta        *ServiceMeta
	}

	ServiceMeta struct {
		baseUrl string
		header  http.Header
		cookies []*http.Cookie
	}
)

func (srv *Service) InitComments(cmap ast.CommentMap) *Service {
	for i := 0; i < srv.service.NumExplicitMethods(); i++ {
		rawMethod := srv.service.ExplicitMethod(i)
		srv.SetMethod(rawMethod)
	}
	for node := range cmap {
		switch tok := node.(type) {
		case *ast.GenDecl:
			if !srv.Complete() {
				srv.TrySetNode(tok)
			}
		case *ast.Field:
			srv.TryAddField(tok)
		}
	}
	return srv
}

func (srv *Service) SetMethod(rawMethod *types.Func) {
	srv.methods[rawMethod.Pos()] = NewMethod(srv, rawMethod)
}

func (srv *Service) TrySetNode(node *ast.GenDecl) {
	success := true
	if node.Tok.String() != TokenType {
		success = false
	}

	if success {
		for i := 0; i < srv.service.NumExplicitMethods(); i++ {
			method := srv.service.ExplicitMethod(i)
			if method.Pos() < node.Pos() || method.Pos() > node.End() {
				success = false
				break
			}
		}
	}

	if success {
		srv.commentText = strings.Trim(node.Doc.Text(), LF)
	}
	return
}

func (srv *Service) TryAddField(node *ast.Field) {
	if method, ok := srv.methods[node.Pos()]; ok {
		if len(node.Names) == 1 && method.Name() == node.Names[0].String() {
			method.commentText = strings.Trim(node.Doc.Text(), LF)
		}
	}
}

func (srv *Service) Complete() bool {
	return srv.commentText != ZeroStr
}

func (srv Service) String() string {
	str := new(strings.Builder)
	fmt.Fprintf(str, "/*\n%s\n*/\n", srv.commentText)
	fmt.Fprintf(str, "type %s interface {\n", srv.name)
	for _, method := range srv.methods {
		fmt.Fprintf(str, "\t/*\n%s\n\t*/\n", method.commentText)
		fmt.Fprintf(str, "\t%s(", method.Name())
		params := method.signature.Params()
		results := method.signature.Results()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			fmt.Fprintf(str, "%s %s", param.Name(), param.Type())
			if i != params.Len()-1 {
				fmt.Fprint(str, ", ")
			}
		}
		fmt.Fprint(str, ") (")
		for i := 0; i < results.Len(); i++ {
			result := results.At(i)
			fmt.Fprintf(str, "%s", result.Type())
			if i != results.Len()-1 {
				fmt.Fprint(str, ", ")
			}
		}
		fmt.Fprint(str, ")\n\n")
	}
	fmt.Fprintln(str, "}")
	return str.String()
}

func (srv *Service) resolveMetadata() (err error) {
	NewProcessor(srv.commentText).Scan(func(ann, key, value string) (err error) {
		switch ann {
		case BaseAnn:
			if srv.meta.baseUrl != ZeroStr {
				err = DuplicatedAnnotationError(BaseAnn)
			}
			srv.meta.baseUrl = value
		case HeaderAnn:
			srv.meta.header.Add(key, value)
		case CookieAnn:
			srv.meta.cookies = append(srv.meta.cookies, &http.Cookie{Name: key, Value: value})
		}
		return
	})

	for _, method := range srv.methods {
		err = method.resolveMetadata()
		if err != nil {
			break
		}
	}
	return
}