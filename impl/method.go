package impl

import (
	"errors"
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
		uriPattern  string
		uriIds      []string
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
			uriIds:      make([]string, 0),
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
	if types.Identical(GetType(TypeIOReader), param.Type()) {
		meta.typ = IOReader
	}
	return
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
	for i := 0; i < results.Len(); i++ {
		result := results.At(i)
		resultList = append(resultList, getQual(result.Type().String()))
	}

	file.Func().
		Params(Id(service.self).Qual(service.pkg, service.implName)).
		Id(method.Name()).
		Params(paramList...).Params(resultList...).
		Block(
			Return(),
		)
	return
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
			var meta *PatternMeta
			meta, err = method.genPatternMeta(key, value)
			method.headerVars = append(method.headerVars, meta)
		case CookieAnn:
			var meta *PatternMeta
			meta, err = method.genPatternMeta(key, value)
			method.cookieVars = append(method.cookieVars, meta)
		case FileAnn:
			err = method.TryAddParam(key, value, TypeFile)
		}
		return
	})

	if err == nil {
		method.resolveLeftIds()
		if method.singleBody && len(method.bodyVars) != 1 {
			err = errors.New(SingleBodyWithMultiBodyVars)
		}
		if err == nil {
			err = method.resolveResultType()
			if err == nil && method.requestType == ZeroStr {
				method.requestType = JSON
			}
		}
	}
	return
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
			if types.Identical(results.At(0).Type(), GetType(TypeRequest)) {
				method.resultType = HttpRequest
			}

			if types.Identical(results.At(0).Type(), GetType(TypeResponse)) {
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
		return
	}
	return
}

func (meta *MethodMeta) TryAddParam(key, pattern string, typ ParamType) (err error) {
	patternMeta, err := meta.genPatternMeta(key, pattern)
	if err == nil {
		meta.bodyVars = append(meta.bodyVars, &BodyMeta{patternMeta, typ})
	}
	return
}

func (meta *MethodMeta) genPatternMeta(key, pattern string) (patternMeta *PatternMeta, err error) {
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
			patterns := IdRe.FindAllString(uriPattern, -1)
			for _, pattern := range patterns {
				id := getIdFromPattern(pattern)
				if err = meta.checkPattern(id); err != nil {
					break
				}
				meta.idList.deleteKey(id)
				meta.uriIds = append(meta.uriIds, id)
			}
			if err == nil {
				meta.uriPattern = IdRe.ReplaceAllStringFunc(uriPattern, meta.findAndReplace)
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
