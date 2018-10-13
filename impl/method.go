package impl

import (
	"go/types"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	IdRegexp = `\{[a-zA-Z_][0-9a-zA-Z_]*\}`
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
		meta        *MethodMeta
	}

	MethodMeta struct {
		idList       IdList // delete when a id is used; to get left ids
		httpMethod   string
		uriPattern   string
		uriIds       []string
		headerVars   []*PatternMeta
		cookieVars   []*PatternMeta
		bodyVars     []*BodyMeta
		totalIds     map[string]*ParamMeta
		responseIds  []string
		responseType BodyType
		requestType  BodyType
		singleBody   bool // json || xml
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
		PatternMeta
		typ ParamType
	}
)

func NewMethod(srv *Service, rawMethod *types.Func) *Method {
	method := &Method{
		Func:      rawMethod,
		service:   srv,
		signature: rawMethod.Type().(*types.Signature),
		meta: &MethodMeta{
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
		method.meta.idList.addKey(param.Name())
		method.meta.totalIds[param.Name()] = NewParamMeta(param)
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

// TODO: complete *Method.resolveMetadata
func (method *Method) resolveMetadata() (err error) {
	NewProcessor(method.commentText).Scan(func(ann, key, value string) (err error) {
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
		case ParamAnn:
		case HeaderAnn:
		case CookieAnn:
		case FileAnn:
		}
		return
	})

	// TODO: resolve left idList
	// TODO: check SingleBody and length of bodyVars
	// TODO: check response type
	return
}

func (method *Method) TrySetBodyType(value string) (err error) {
	if method.meta.requestType == ZeroStr {
		if value == JSON || value == XML || value == Form || value == Multipart {
			method.meta.requestType = BodyType(value)
		} else {
			err = UnsupportedAnnotationValueError(BodyAnn, value)
		}
	} else {
		err = DuplicatedAnnotationError(BodyAnn + "/" + SingleBodyAnn)
	}
	return
}

func (method *Method) TrySetSingleBodyType(value string) (err error) {
	if method.meta.requestType == ZeroStr {
		if value == JSON || value == XML {
			method.meta.requestType = BodyType(value)
			method.meta.singleBody = true
		} else {
			err = UnsupportedAnnotationValueError(SingleBodyAnn, value)
		}
	} else {
		err = DuplicatedAnnotationError(BodyAnn + "/" + SingleBodyAnn)
	}
	return
}

func (method *Method) TrySetMethod(httpMethod, uriPattern string) (err error) {
	if method.meta.httpMethod != ZeroStr {
		err = DuplicatedHttpMethodError(httpMethod)
	}

	if err == nil {
		_, err = url.Parse(uriPattern)
		if err == nil {
			method.meta.httpMethod = httpMethod
			patterns := IdRe.FindAllString(uriPattern, -1)
			for _, pattern := range patterns {
				id := getIdFromPattern(pattern)
				if err = method.checkPath(id); err != nil {
					break
				}
				method.meta.idList.deleteKey(id)
				method.meta.uriIds = append(method.meta.uriIds, id)
			}
			if err == nil {
				method.meta.uriPattern = IdRe.ReplaceAllStringFunc(uriPattern, method.findAndReplace)
			}
		}
	}
	return
}

func getIdFromPattern(pattern string) string {
	return strings.TrimRight(strings.TrimLeft(pattern, "{"), "}")
}

func (method *Method) findAndReplace(pattern string) (placeholder string) {
	id := getIdFromPattern(pattern)
	paramMeta := method.meta.totalIds[id]
	switch paramMeta.typ {
	case TypeString:
		placeholder = "%s"
	case TypeInt:
		placeholder = "%d"
	default:
		log.Fatal(PathIdTypeMustBeIntOrStringError(id))
	}
	return
}

func (method *Method) checkPath(id string) (err error) {
	if meta, exist := method.meta.totalIds[id]; exist {
		if meta.typ != TypeString && meta.typ != TypeInt {
			err = PathIdTypeMustBeIntOrStringError(id)
		}
	} else {
		err = IdNotExistError(id)
	}
	return
}
