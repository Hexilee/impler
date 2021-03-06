package impl

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/rady-io/annotation-processor"
	. "github.com/rady-io/http-service/log"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"time"
)

const (
	TokenType = "type"
	ZeroStr   = ""
	LF        = "\n"
)

func NewService(name string, info *types.Info) *Service {
	return &Service{
		info:    info,
		methods: make(map[token.Pos]*Method),
		name:    name,
		ServiceMeta: &ServiceMeta{
			idList:     make([]string, 0),
			headerVars: make([]*PatternMeta, 0),
			cookieVars: make([]*PatternMeta, 0),
		},
	}
}

type (
	Service struct {
		info        *types.Info
		methods     map[token.Pos]*Method
		name        string
		commentText string
		service     *types.Interface
		*ServiceMeta
	}

	ServiceMeta struct {
		idList                       []string
		baseUrl                      *PatternMeta
		headerVars                   []*PatternMeta
		cookieVars                   []*PatternMeta
		self, pkg, implName, newFunc string
	}
)

func (srv *Service) resolveCode(file *File) (err error) {
	file.HeaderComment(fmt.Sprintf(`Implement of %s.%s
This file is generated by github.com/Hexilee/impler at %s
DON'T EDIT IT!
`, srv.pkg, srv.name, time.Now()))
	file.Func().Id(srv.newFunc).Params(srv.getParams()).Qual(srv.pkg, srv.name).BlockFunc(func(group *Group) {
		group.Id(srv.self).Op(":=").Op("&").Id(srv.implName).Values(Dict{
			Id(FieldHeader):  Make(Qual(HttpPkg, "Header")),
			Id(FieldCookies): Make(Index().Op("*").Qual(HttpPkg, "Cookie"), Lit(0)),
		})
		srv.setBaseUrl(group)
		srv.addHeader(group)
		srv.addCookies(group)
		group.Return(Id(srv.self))

	})

	file.Type().Id(srv.implName).Struct(
		Id(FieldBaseUrl).String(),
		Id(FieldHeader).Qual(HttpPkg, "Header"),
		Id(FieldCookies).Index().Op("*").Qual(HttpPkg, "Cookie"),
	)

	for _, method := range srv.methods {
		Log.Infof("Implement method: %s", method.String())
		err = method.resolveMetadata()
		if err != nil {
			break
		}
		method.resolveCode(file)
	}
	return
}

func (srv *Service) getParams() (params Code) {
	paramList := make([]Code, 0)
	for _, id := range srv.idList {
		paramList = append(paramList, Id(id))
	}

	if len(paramList) == 0 {
		params = List(paramList...)
	} else {
		params = List(paramList...).Add(String())
	}
	return
}

func (srv *Service) setBaseUrl(group *Group) {
	if len(srv.baseUrl.ids) == 0 {
		group.Id(srv.self).Dot(FieldBaseUrl).Op("=").Lit(srv.baseUrl.pattern)
	} else if srv.baseUrl.pattern == StringPlaceholder {
		group.Id(srv.self).Dot(FieldBaseUrl).Op("=").Id(srv.baseUrl.ids[0])
	} else {
		// add Format pkg
		group.Id(srv.self).Dot(FieldBaseUrl).Op("=").
			Qual(FormatPkg, "Sprintf").Call(Lit(srv.baseUrl.pattern), List(genIds(srv.baseUrl.ids)...))
	}
}

func (srv *Service) addHeader(group *Group) {
	for _, pattern := range srv.headerVars {
		if len(pattern.ids) == 0 {
			group.Id(srv.self).Dot(FieldHeader).Dot("Add").Call(Lit(pattern.key), Lit(pattern.pattern))
		} else if pattern.pattern == StringPlaceholder {
			group.Id(srv.self).Dot(FieldHeader).Dot("Add").Call(Lit(pattern.key), Id(pattern.ids[0]))
		} else {
			group.Id(srv.self).Dot(FieldHeader).
				Dot("Add").Call(Lit(pattern.key),
				Qual(FormatPkg, "Sprintf").Call(Lit(pattern.pattern), List(genIds(pattern.ids)...)))
		}
	}
}

func (srv *Service) addCookies(group *Group) {
	for _, pattern := range srv.cookieVars {
		if len(pattern.ids) == 0 {
			group.Id(srv.self).Dot(FieldCookies).Op("=").Append(
				Id(srv.self).Dot(FieldCookies),
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Lit(pattern.pattern),
				}),
			)
		} else if pattern.pattern == StringPlaceholder {
			group.Id(srv.self).Dot(FieldCookies).Op("=").Append(
				Id(srv.self).Dot(FieldCookies),
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Id(pattern.ids[0]),
				}),
			)
		} else {
			group.Id(srv.self).Dot(FieldCookies).Op("=").Append(
				Id(srv.self).Dot(FieldCookies),
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Qual(FormatPkg, "Sprintf").Call(Lit(pattern.pattern), List(genIds(pattern.ids)...)),
				}),
			)
		}
	}
}

func (srv *Service) InitComments(cmap ast.CommentMap) *Service {
	for node := range cmap {
		if tok, ok := node.(*ast.GenDecl); ok &&  !srv.Complete() {
			srv.TrySetNode(tok)
		}
	}
	for node := range cmap {
		if tok, ok := node.(*ast.Field); ok {
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
		for _, spec := range node.Specs {
			typ := spec.(*ast.TypeSpec)
			name := typ.Name.String()
			if name == srv.name {
				obj := srv.info.Defs[typ.Name]
				if obj != nil {
					if service, ok := obj.Type().Underlying().(*types.Interface); ok {
						srv.setMethods(service)
						srv.service = service
						srv.commentText = combineComments(node.Doc.Text(), typ.Doc.Text())
					}
				}
			}
		}
	}
	return
}

func combineComments(typeSpecComments, interfaceComments string) (comments string) {
	typeSpecComments = strings.Trim(typeSpecComments, LF)
	interfaceComments = strings.Trim(interfaceComments, LF)
	comments = typeSpecComments + LF + interfaceComments
	return
}

func (srv *Service) setMethods(service *types.Interface) {
	for i := 0; i < service.NumExplicitMethods(); i++ {
		rawMethod := service.ExplicitMethod(i)
		srv.SetMethod(rawMethod)
	}
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
	err = processor.NewProcessor(srv.commentText).Scan(func(ann, key, value string) (err error) {
		switch ann {
		case BaseAnn:
			err = srv.ServiceMeta.trySetBaseUrl(value)
		case HeaderAnn:
			srv.ServiceMeta.addHeader(key, value)
		case CookieAnn:
			srv.ServiceMeta.addCookie(key, value)
		}
		return
	})
	if err == nil {
		srv.resolveBaseUrl()
	}
	return
}

func (meta *ServiceMeta) trySetBaseUrl(baseUrl string) (err error) {
	if meta.baseUrl != nil {
		err = DuplicatedAnnotationError(BaseAnn)
	}
	if err == nil {
		Log.Debugf("Set BaseURL: %s", baseUrl)
		meta.baseUrl = meta.genPatternMeta("baseUrl", baseUrl)
	}
	return
}

func (meta *ServiceMeta) addHeader(key, value string) {
	Log.Debugf("Add Header: %s(%s)", key, value)
	var patternMeta *PatternMeta
	patternMeta = meta.genPatternMeta(key, value)
	meta.headerVars = append(meta.headerVars, patternMeta)
}

func (meta *ServiceMeta) addCookie(key, value string) {
	Log.Debugf("Add Cookie: %s(%s)", key, value)
	var patternMeta *PatternMeta
	patternMeta = meta.genPatternMeta(key, value)
	meta.cookieVars = append(meta.cookieVars, patternMeta)
	return
}

func (meta *ServiceMeta) resolveBaseUrl() {
	if meta.baseUrl == nil {
		baseUrl := "{baseUrl}"
		Log.Debugf("Set BaseURL: %s", baseUrl)
		meta.baseUrl = meta.genPatternMeta("baseUrl", baseUrl)
	}
}

func (meta *ServiceMeta) genPatternMeta(key, pattern string) (patternMeta *PatternMeta) {
	patternMeta = &PatternMeta{
		key: key,
		ids: make([]string, 0),
	}
	patterns := IdRe.FindAllString(pattern, -1)
	for _, pattern := range patterns {
		id := getIdFromPattern(pattern)
		meta.idList = append(meta.idList, id)
		patternMeta.ids = append(patternMeta.ids, id)
	}
	patternMeta.pattern = IdRe.ReplaceAllStringFunc(pattern, meta.findAndReplace)
	return
}

func (meta ServiceMeta) findAndReplace(pattern string) (placeholder string) {
	placeholder = StringPlaceholder
	return
}
