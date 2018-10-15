package impl

import (
	"errors"
	"github.com/Hexilee/impler/headers"
	. "github.com/Hexilee/impler/log"
	. "github.com/dave/jennifer/jen"
	"go/types"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	IdRegexp          = `\{[a-zA-Z_][0-9a-zA-Z_]*\}`
	StringPlaceholder = "%s"
	IntPlaceholder    = "%d"
)

const (
	// ids
	IdResult      = "genResult"
	IdRequest     = "genRequest"
	IdResponse    = "genResponse"
	IdClient      = "genClient"
	IdStatusCode  = "genStatusCode"
	IdError       = "genErr"
	IdUri         = "genUri"
	IdUrl         = "genFinalUrl"
	IdData        = "genData"
	IdResultData  = "genResultData"
	IdBody        = "genBody"
	IdDataMap     = "genDataMap"
	IdPartWriter  = "genPartWriter"
	IdBodyWriter  = "genBodyWriter"
	IdFile        = "genFile"
	IdFilePath    = "genFilePath"
	IdCookie      = "genCookie"
	IdHeaderKey   = "genHeaderKey"
	IdHeaderSlice = "genHeaderSlice"
	IdHeaderValue = "genHeaderValue"
	IdHeader      = "genHeader"
)

var (
	IdRe = regexp.MustCompile(IdRegexp)
)

type (
	Method struct {
		*types.Func
		commentText string
		service     *Service
		signature   *types.Signature
		*MethodMeta
	}

	MethodMeta struct {
		idList      IdList // delete when a id is used; to get left ids
		httpMethod  string
		uri         *PatternMeta
		headerVars  []*PatternMeta
		cookieVars  []*PatternMeta
		bodyVars    []*BodyMeta // left params as '@Param(id) {id}'
		totalIds    map[string]*ParamMeta
		responseIds []string
		resultType  BodyType
		requestType BodyType
		singleBody  bool // json || xml
	}

	ParamMeta struct {
		key string
		typ ParamType
	}

	PatternMeta struct {
		key     string
		pattern string
		ids     []string
	}

	BodyMeta struct {
		*PatternMeta
		typ ParamType
	}
)

func NewMethod(srv *Service, rawMethod *types.Func) *Method {
	method := &Method{
		Func:      rawMethod,
		service:   srv,
		signature: rawMethod.Type().(*types.Signature),
		MethodMeta: &MethodMeta{
			idList:      make(IdList),
			headerVars:  make([]*PatternMeta, 0),
			cookieVars:  make([]*PatternMeta, 0),
			totalIds:    make(map[string]*ParamMeta),
			bodyVars:    make([]*BodyMeta, 0),
			responseIds: make([]string, 0),
		},
	}

	params := method.signature.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		method.idList.addKey(param.Name())
		method.totalIds[param.Name()] = NewParamMeta(param)
	}

	return method
}

func NewParamMeta(param *types.Var) (meta *ParamMeta) {
	meta = &ParamMeta{key: param.Name()}
	switch typ := param.Type().(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.Int:
			meta.typ = TypeInt
		case types.String:
			meta.typ = TypeString
		default:
			meta.typ = Other
		}
	default:
		meta.typ = Other
	}
	if GetType(TypeIOReader).String() == param.Type().String() {
		meta.typ = IOReader
	}
	return
}

func genIds(ids []string) []Code {
	results := make([]Code, 0)
	for _, id := range ids {
		results = append(results, Id(id))
	}
	return results
}

func (method *Method) resolveCode(file *File) {
	service := method.service
	paramList := make([]Code, 0)
	resultList := make([]Code, 0)
	params := method.signature.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		paramList = append(paramList, Id(param.Name()).Add(getQual(param.Type().String())))
	}

	results := method.signature.Results()
	resultList = append(resultList, Id(IdResult).Add(getQual(results.At(0).Type().String())))
	if results.Len() == 2 {
		resultList = append(resultList, Id(IdError).Add(getQual(results.At(1).Type().String())))
	}

	if results.Len() == 3 {
		resultList = append(resultList, Id(IdStatusCode).Add(getQual(results.At(1).Type().String())))
		resultList = append(resultList, Id(IdError).Add(getQual(results.At(2).Type().String())))
	}

	file.Func().
		Params(Id(service.self).Qual(service.pkg, service.implName)).
		Id(method.Name()).
		Params(paramList...).Params(resultList...).
		BlockFunc(method.genMethodBody)
	return
}

func (method *Method) genMethodBody(group *Group) {
	group.Var().Id(IdBody).Qual(IO, "ReadWriter")
	group.Var().Id(IdRequest).Op("*").Qual(HttpPkg, "Request")
	if len(method.uri.ids) == 0 {
		group.Id(IdUri).Op(":=").Lit(method.uri.pattern)
	} else if method.uri.pattern == StringPlaceholder {
		group.Id(IdUri).Op(":=").Id(method.uri.ids[0])
	} else {
		group.Id(IdUri).Op(":=").Qual("fmt", "Sprintf").Call(Lit(method.uri.pattern), List(genIds(method.uri.ids)...))
	}
	method.genBody(group)
	method.genRequest(group)
	// TODO: check @Header, cannot set contentType
	method.addHeader(group)
	method.addCookies(group)
	method.setContentType(group)
	method.genResult(group)
	group.Return()
}

func (method *Method) genBody(group *Group) {
	if len(method.bodyVars) > 0 {
		switch method.requestType {
		case JSON:
			method.genJSONOrXMLBody(group, EncodingJSON)
		case XML:
			method.genJSONOrXMLBody(group, EncodingXML)
		case Form:
			method.genFormBody(group)
		case Multipart:
			group.Var().Id(IdBodyWriter).Op("*").Qual(MultipartPkg, "Writer")
			method.genMultipartBody(group)
		}
	}
}

func (method *Method) setContentType(group *Group) {
	if len(method.bodyVars) > 0 {
		var contentType *Statement
		switch method.requestType {
		case JSON:
			contentType = Lit(headers.MIMEApplicationJSONCharsetUTF8)
		case XML:
			contentType = Lit(headers.MIMEApplicationXMLCharsetUTF8)
		case Form:
			contentType = Lit(headers.MIMEApplicationForm)
		case Multipart:
			contentType = Id(IdBodyWriter).Dot("FormDataContentType").Call()
		}
		group.Id(IdRequest).Dot("Header").
			Dot("Set").Call(Lit(headers.HeaderContentType), contentType)
	}
}

func (method *Method) genRequest(group *Group) {
	group.Id(IdUrl).Op(":=").Qual(StringsPkg, "TrimRight").Call(Id(method.service.self).Dot(FieldBaseUrl), Lit("/")).
		Op("+").
		Lit("/").
		Op("+").
		Qual(StringsPkg, "TrimLeft").Call(Id(IdUri), Lit("/"))

	group.List(Id(IdRequest), Id(IdError)).Op("=").
		Qual(HttpPkg, "NewRequest").Call(Lit(method.httpMethod), Id(IdUrl), Id(IdBody))

	group.If(Id(IdError).Op("!=").Nil()).Block(Return())
}

func (method *Method) addHeader(group *Group) {
	group.For(List(Id(IdHeaderKey), Id(IdHeaderSlice))).Op(":=").Range().Id(method.service.self).Dot(FieldHeader).Block(
		For(List(Id("_"), Id(IdHeaderValue))).Op(":=").Range().Id(IdHeaderSlice).Block(
			Id(IdRequest).Dot("Header").Dot("Add").Call(Id(IdHeaderKey), Id(IdHeaderValue)),
		),
	)

	for _, pattern := range method.headerVars {
		if len(pattern.ids) == 0 {
			group.Id(IdRequest).Dot("Header").Dot("Set").Call(Lit(pattern.key), Lit(pattern.pattern))
		} else if pattern.pattern == StringPlaceholder {
			group.Id(IdRequest).Dot("Header").Dot("Set").Call(Lit(pattern.key), Id(pattern.ids[0]))
		} else {
			group.Id(IdRequest).Dot("Header").
				Dot("Set").Call(Lit(pattern.key),
				Qual("fmt", "Sprintf").Call(Lit(pattern.pattern), List(genIds(pattern.ids)...)))
		}
	}
}

func (method *Method) addCookies(group *Group) {
	group.For(List(Id("_"), Id(IdCookie))).Op(":=").Range().Id(method.service.self).Dot(FieldCookies).Block(
		Id(IdRequest).Dot("AddCookie").Call(Id(IdCookie)),
	)

	for _, pattern := range method.cookieVars {
		if len(pattern.ids) == 0 {
			group.Id(IdRequest).Dot("AddCookie").Call(
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Lit(pattern.pattern),
				}),
			)
		} else if pattern.pattern == StringPlaceholder {
			group.Id(IdRequest).Dot("AddCookie").Call(
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Id(pattern.ids[0]),
				}),
			)
		} else {
			group.Id(IdRequest).Dot("AddCookie").Call(
				Op("&").Qual(HttpPkg, "Cookie").Values(Dict{
					Id("Name"):  Lit(pattern.key),
					Id("Value"): Qual("fmt", "Sprintf").Call(Lit(pattern.pattern), List(genIds(pattern.ids)...)),
				}),
			)
		}
	}
}

func (method *Method) genResult(group *Group) {
	if method.resultType == HttpRequest {
		group.Id(IdResult).Op("=").Id(IdRequest)
	} else {
		group.Var().Id(IdResponse).Op("*").Qual(HttpPkg, "Response")
		group.Id(IdClient).Op(":=").New(Qual(HttpPkg, "Client"))
		group.List(Id(IdResponse), Id(IdError)).Op("=").Id(IdClient).Dot("Do").Call(Id(IdRequest))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())
		switch method.resultType {
		case HttpResponse:
			group.Id(IdResult).Op("=").Id(IdResponse)
		case JSON:
			method.unmarshalResult(group, EncodingJSON)
		case XML:
			method.unmarshalResult(group, EncodingXML)
		case HTML:
			method.unmarshalResult(group, UnHTMLPkg)
		}
	}
}

func (method *Method) unmarshalResult(group *Group, pkg string) {
	group.Var().Id(IdResultData).Index().Byte()
	group.List(Id(IdResultData), Id(IdError)).Op("=").
		Qual(Ioutil, "ReadAll").Call(Id(IdResponse).Dot("Body"))
	group.Defer().Id(IdResponse).Dot("Body").Dot("Close").Call()
	group.If(Id(IdError).Op("!=").Nil()).Block(Return())
	group.Id(IdStatusCode).Op("=").Id(IdResponse).Dot("StatusCode")
	group.Id(IdResult).Op("=").Add(method.newObject(method.signature.Results().At(0).Type().String())).Values()
	group.Id(IdError).Op("=").Qual(pkg, "Unmarshal").Call(Id(IdResultData), Id(IdResult))
	group.If(Id(IdError).Op("!=").Nil()).Block(Return())
}

func (method *Method) newObject(typ string) Code {
	var statement *Statement
	var qual *Statement
	if strings.HasPrefix(typ, "*") {
		statement = Op("&")
		typ = typ[1:]
	}

	if !strings.Contains(typ, ".") {
		qual = Id(typ)
	} else {
		var pkg string
		dot := strings.LastIndex(typ, ".")
		if dot != -1 {
			pkg = typ[:dot]
			typ = typ[dot+1:]
		}
		qual = Qual(pkg, typ)
	}

	if statement == nil {
		statement = qual
	} else {
		statement = statement.Add(qual)
	}
	return statement
}

func (method *Method) genFormBody(group *Group) {
	group.Id(IdDataMap).Op(":=").Make(Qual(NetURL, "Values"))
	for _, bodyVar := range method.bodyVars {
		switch bodyVar.typ {
		case TypeInt:
			fallthrough
		case TypeString:
			if len(bodyVar.ids) == 0 {
				group.Id(IdDataMap).Dot("Add").Call(Lit(bodyVar.key), Lit(bodyVar.pattern))
			} else if bodyVar.pattern == StringPlaceholder {
				group.Id(IdDataMap).Dot("Add").Call(Lit(bodyVar.key), Id(bodyVar.ids[0]))
			} else {
				group.Id(IdDataMap).Dot("Add").Call(Lit(bodyVar.key), Qual("fmt", "Sprintf").Call(Lit(bodyVar.pattern), List(genIds(bodyVar.ids)...)))
			}
		}
	}
	group.Id(IdBody).Op("=").Qual(Bytes, "NewBufferString").Call(Id(IdDataMap).Dot("Encode").Call())
}

func (method *Method) getIOReaderWriter(bodyVar *BodyMeta) func(group *Group) {
	return func(group *Group) {
		group.Var().Id(IdPartWriter).Qual(IO, "Writer")
		group.List(Id(IdPartWriter), Id(IdError)).Op("=").
			Id(IdBodyWriter).Dot("CreateFormField").Call(Lit(bodyVar.key))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())
		group.List(Id("_"), Id(IdError)).Op("=").Qual(IO, "Copy").Call(Id(IdPartWriter), Id(bodyVar.ids[0]))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())
	}
}

func (method *Method) getFileWriter(bodyVar *BodyMeta) func(group *Group) {
	return func(group *Group) {
		group.Var().Id(IdPartWriter).Qual(IO, "Writer")
		group.Var().Id(IdFile).Op("*").Qual(OS, "File")
		group.Var().Id(IdFilePath).String()
		if len(bodyVar.ids) == 0 {
			group.Id(IdFilePath).Op("=").Lit(bodyVar.pattern)
		} else if bodyVar.pattern == StringPlaceholder {
			group.Id(IdFilePath).Op("=").Id(bodyVar.ids[0])
		} else {
			group.Id(IdFilePath).Op("=").Qual("fmt", "Sprintf").Call(Lit(bodyVar.pattern), List(genIds(bodyVar.ids)...))
		}
		group.List(Id(IdFile), Id(IdError)).Op("=").Qual(OS, "Open").Call(Id(IdFilePath))
		group.Defer().Id(IdFile).Dot("Close").Call()
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())

		group.List(Id(IdPartWriter), Id(IdError)).Op("=").
			Id(IdBodyWriter).Dot("CreateFormFile").Call(Lit(bodyVar.key), Id(IdFilePath))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())

		group.List(Id("_"), Id(IdError)).Op("=").Qual(IO, "Copy").Call(Id(IdPartWriter), Id(IdFile))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())
	}
}

func (method *Method) genMultipartBody(group *Group) {
	group.Id(IdBody).Op("=").Qual(Bytes, "NewBufferString").Call(Lit(""))
	group.Id(IdBodyWriter).Op("=").Qual(MultipartPkg, "NewWriter").Call(Id(IdBody))
	for _, bodyVar := range method.bodyVars {
		switch bodyVar.typ {
		case TypeInt:
			fallthrough
		case TypeString:
			if len(bodyVar.ids) == 0 {
				group.Id(IdBodyWriter).Dot("WriteField").Call(Lit(bodyVar.key), Lit(bodyVar.pattern))
			} else if bodyVar.pattern == StringPlaceholder {
				group.Id(IdBodyWriter).Dot("WriteField").Call(Lit(bodyVar.key), Id(bodyVar.ids[0]))
			} else {
				group.Id(IdBodyWriter).Dot("WriteField").Call(Lit(bodyVar.key), Qual("fmt", "Sprintf").Call(Lit(bodyVar.pattern), List(genIds(bodyVar.ids)...)))
			}
		case IOReader:
			group.BlockFunc(method.getIOReaderWriter(bodyVar))
		case TypeFile:
			group.BlockFunc(method.getFileWriter(bodyVar))
		}
	}
	group.Id(IdBodyWriter).Dot("Close").Call()
}

func (method *Method) genJSONOrXMLBody(group *Group, pkg string) {
	if method.singleBody {
		switch method.bodyVars[0].typ {
		case IOReader:
			group.Id(IdBody).Op("=").Qual(Bytes, "NewBufferString").Call(Lit(""))
			group.List(Id("_"), Id(IdError)).Op("=").Qual(IO, "Copy").Call(Id(IdBody), Id(method.bodyVars[0].ids[0]))
			group.If(Id(IdError).Op("!=").Nil()).Block(Return())
		case Other:
			group.Var().Id(IdData).Index().Byte()
			group.List(Id(IdData), Id(IdError)).Op("=").Qual(pkg, "Marshal").Call(Id(method.bodyVars[0].ids[0]))
			group.If(Id(IdError).Op("!=").Nil()).Block(Return())
			group.Id(IdBody).Op("=").Qual(Bytes, "NewBuffer").Call(Id(IdData))
		}
	} else {
		group.Var().Id(IdData).Index().Byte()
		group.Id(IdDataMap).Op(":=").Make(Map(String()).Interface())
		for _, bodyVar := range method.bodyVars {
			switch bodyVar.typ {
			case TypeInt:
				fallthrough
			case TypeString:
				if len(bodyVar.ids) == 0 {
					group.Id(IdDataMap).Index(Lit(bodyVar.key)).Op("=").Lit(bodyVar.pattern)
				} else if bodyVar.pattern == StringPlaceholder {
					group.Id(IdDataMap).Index(Lit(bodyVar.key)).Op("=").Id(bodyVar.ids[0])
				} else {
					group.Id(IdDataMap).Index(Lit(bodyVar.key)).Op("=").Qual("fmt", "Sprintf").Call(Lit(bodyVar.pattern), List(genIds(bodyVar.ids)...))
				}
			case IOReader:
				group.List(Id(IdData), Id(IdError)).Op("=").Qual(Ioutil, "ReadAll").Call(Id(bodyVar.ids[0]))
				group.Id(IdDataMap).Index(Lit(bodyVar.key)).Op("=").String().Values(Id(IdData))
			case Other:
				group.Id(IdDataMap).Index(Lit(bodyVar.key)).Op("=").Id(bodyVar.ids[0])
			}
		}
		group.List(Id(IdData), Id(IdError)).Op("=").Qual(pkg, "Marshal").Call(Id(IdDataMap))
		group.If(Id(IdError).Op("!=").Nil()).Block(Return())
		group.Id(IdBody).Op("=").Qual(Bytes, "NewBuffer").Call(Id(IdData))
	}
}

func (method *Method) resolveMetadata() (err error) {
	err = NewProcessor(method.commentText).Scan(func(ann, key, value string) (err error) {
		switch ann {
		case GetAnn:
			err = method.TrySetMethod(http.MethodGet, value)
		case HeadAnn:
			err = method.TrySetMethod(http.MethodHead, value)
		case PostAnn:
			err = method.TrySetMethod(http.MethodPost, value)
		case PutAnn:
			err = method.TrySetMethod(http.MethodPut, value)
		case PatchAnn:
			err = method.TrySetMethod(http.MethodPatch, value)
		case DeleteAnn:
			err = method.TrySetMethod(http.MethodDelete, value)
		case ConnectAnn:
			err = method.TrySetMethod(http.MethodConnect, value)
		case OptionsAnn:
			err = method.TrySetMethod(http.MethodOptions, value)
		case TraceAnn:
			err = method.TrySetMethod(http.MethodTrace, value)
		case BodyAnn:
			err = method.TrySetBodyType(value)
		case SingleBodyAnn:
			err = method.TrySetSingleBodyType(value)
		case ResultAnn:
			err = method.TrySetResultType(value)
		case ParamAnn:
			err = method.TryAddParam(key, value, TypeString)
		case HeaderAnn:
			err = method.TryAddHeader(key, value)
		case CookieAnn:
			err = method.TryAddCookie(key, value)
		case FileAnn:
			err = method.TryAddParam(key, value, TypeFile)
		}
		return
	})

	if err == nil {
		method.resolveLeftIds()
		err = method.checkSingleBody()
		if err == nil {
			method.resolveRequestType()
			method.resolveUri()
			err = method.resolveResultType()
			if err == nil {
				Log.Debugf(`Final URI: "%s".Format(%v...)`, method.uri.pattern, method.uri.ids)
				Log.Debugf("Final Request Type: %s", method.requestType)
				Log.Debugf("Final Result Type: %s", method.resultType)
			}
		}
	}
	return
}

func (method *Method) resolveUri() {
	if method.uri == nil {
		method.uri, _ = method.genPatternMeta("uri", "/")
	}
}

func (method *Method) checkSingleBody() (err error) {
	if method.singleBody {
		if len(method.bodyVars) != 1 ||
		// if singleBody, the type of single body must be IOReader or Other
			method.bodyVars[0].typ != IOReader && method.bodyVars[0].typ != Other {
			err = errors.New(SingleBodyWithMultiBodyVars)
		}
	}
	return
}

func (method *Method) resolveRequestType() {
	if method.requestType == ZeroStr {
		method.requestType = JSON
	}
}

func (method *Method) resolveResultType() (err error) {
	results := method.signature.Results()
	switch results.Len() {
	case 2:
		// TODO: compare types in a robuster way
		if method.resultType == JSON ||
			method.resultType == XML ||
			method.resultType == HTML ||
			results.At(0).Type().String() != GetType(TypeRequest).String() &&
				results.At(0).Type().String() != GetType(TypeResponse).String() ||
			!types.Identical(results.At(1).Type(), GetType(TypeErr)) {
			err = ConflictAnnotationError(ResultAnn, results)
		} else {
			if results.At(0).Type().String() == GetType(TypeRequest).String() {
				method.resultType = HttpRequest
			}

			if results.At(0).Type().String() == GetType(TypeResponse).String() {
				method.resultType = HttpResponse
			}
		}
	case 3:
		if method.resultType != JSON &&
			method.resultType != XML &&
			method.resultType != HTML &&
			method.resultType != ZeroStr ||
			!types.Identical(results.At(1).Type(), GetType(TypeStatusCode)) ||
			!types.Identical(results.At(2).Type(), GetType(TypeErr)) {
			err = ConflictAnnotationError(ResultAnn, results)
		}
		if err == nil && method.resultType == ZeroStr {
			method.resultType = JSON
		}
	default:
		err = ConflictAnnotationError(ResultAnn, results)
	}
	return
}

func (meta *MethodMeta) resolveLeftIds() {
	for id := range meta.idList {
		if paramMeta, exist := meta.totalIds[id]; exist {
			Log.Debugf("Set Param(%s) <- %s", paramMeta.key, id)
			patternMeta := &PatternMeta{key: paramMeta.key, ids: []string{id}}
			switch paramMeta.typ {
			case TypeString:
				patternMeta.pattern = StringPlaceholder
			case TypeInt:
				patternMeta.pattern = IntPlaceholder
			}
			bodyMeta := &BodyMeta{patternMeta, paramMeta.typ}
			meta.bodyVars = append(meta.bodyVars, bodyMeta)
		} else {
			log.Fatal(IdNotExistError(id))
		}
	}
	return
}

func (meta *MethodMeta) TryAddHeader(key, value string) (err error) {
	var patternMeta *PatternMeta
	patternMeta, err = meta.genPatternMeta(key, value)
	meta.headerVars = append(meta.headerVars, patternMeta)
	return
}

func (meta *MethodMeta) TryAddCookie(key, value string) (err error) {
	var patternMeta *PatternMeta
	patternMeta, err = meta.genPatternMeta(key, value)
	meta.headerVars = append(meta.cookieVars, patternMeta)
	return
}

func (meta *MethodMeta) TryAddParam(key, pattern string, typ ParamType) (err error) {
	patternMeta, err := meta.genPatternMeta(key, pattern)
	if err == nil {
		Log.Debugf("Set Param(%s) %s", patternMeta.key, pattern)
		meta.bodyVars = append(meta.bodyVars, &BodyMeta{patternMeta, typ})
	}
	return
}

func (meta *MethodMeta) genPatternMeta(key, pattern string) (patternMeta *PatternMeta, err error) {
	// TODO: can be empty?
	if key == ZeroStr {
		err = errors.New(PatternKeyMustNotBeEmpty)
	}

	if err == nil {
		patternMeta = &PatternMeta{
			key: key,
			ids: make([]string, 0),
		}
		patterns := IdRe.FindAllString(pattern, -1)
		for _, pattern := range patterns {
			id := getIdFromPattern(pattern)
			if err = meta.checkPattern(id); err != nil {
				break
			}
			meta.idList.deleteKey(id)
			patternMeta.ids = append(patternMeta.ids, id)
		}
		if err == nil {
			patternMeta.pattern = IdRe.ReplaceAllStringFunc(pattern, meta.findAndReplace)
		}
	}
	return
}

func (meta *MethodMeta) TrySetBodyType(value string) (err error) {
	if meta.requestType == ZeroStr {
		if value == JSON || value == XML || value == Form || value == Multipart {
			Log.Debugf("Set Request Body: %s", value)
			meta.requestType = BodyType(value)
		} else {
			err = UnsupportedAnnotationValueError(BodyAnn, value)
		}
	} else {
		err = DuplicatedAnnotationError(BodyAnn + "/" + SingleBodyAnn)
	}
	return
}

func (meta *MethodMeta) TrySetResultType(value string) (err error) {
	if meta.resultType == ZeroStr {
		if value == JSON || value == XML || value == HTML {
			meta.resultType = BodyType(value)
		} else {
			err = UnsupportedAnnotationValueError(ResultAnn, value)
		}
	} else {
		err = DuplicatedAnnotationError(ResultAnn)
	}
	return
}

func (meta *MethodMeta) TrySetSingleBodyType(value string) (err error) {
	if meta.requestType == ZeroStr {
		if value == JSON || value == XML {
			Log.Debugf("Set Request Body: %s(Single)", value)
			meta.requestType = BodyType(value)
			meta.singleBody = true
		} else {
			err = UnsupportedAnnotationValueError(SingleBodyAnn, value)
		}
	} else {
		err = DuplicatedAnnotationError(BodyAnn + "/" + SingleBodyAnn)
	}
	return
}

func (meta *MethodMeta) TrySetMethod(httpMethod, uriPattern string) (err error) {
	if meta.httpMethod != ZeroStr {
		err = DuplicatedHttpMethodError(httpMethod)
	}

	if err == nil {
		_, err = url.Parse(uriPattern)
		if err == nil {
			meta.httpMethod = httpMethod
			meta.uri, err = meta.genPatternMeta("uri", uriPattern)
			if err == nil {
				Log.Debugf("Set Method: %s(%s)", httpMethod, uriPattern)
			}
		}
	}
	return
}

func getIdFromPattern(pattern string) string {
	return strings.TrimRight(strings.TrimLeft(pattern, "{"), "}")
}

func (meta *MethodMeta) findAndReplace(pattern string) (placeholder string) {
	id := getIdFromPattern(pattern)
	paramMeta := meta.totalIds[id]
	switch paramMeta.typ {
	case TypeString:
		placeholder = StringPlaceholder
	case TypeInt:
		placeholder = IntPlaceholder
	default:
		log.Fatal(PatternIdTypeMustBeIntOrStringError(id))
	}
	return
}

func (meta *MethodMeta) checkPattern(id string) (err error) {
	if paramMeta, exist := meta.totalIds[id]; exist {
		if paramMeta.typ != TypeString && paramMeta.typ != TypeInt {
			err = PatternIdTypeMustBeIntOrStringError(id)
		}
	} else {
		err = IdNotExistError(id)
	}
	return
}
