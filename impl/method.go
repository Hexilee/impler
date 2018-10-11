package impl

import (
	"go/types"
)

//const (
//	ReadName  = "Read"
//	ErrorName = "Error"
//)

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
		headerIds    []*ParamMeta
		cookieIds    []*ParamMeta
		totalIds     map[string]*ParamMeta
		bodyIds      map[string]*ParamMeta // left ids is bodyIds
		responseIds  []string
		responseType BodyType
		requestType  BodyType
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
)

func NewMethod(srv *Service, rawMethod *types.Func) *Method {
	method := &Method{
		Func:      rawMethod,
		service:   srv,
		signature: rawMethod.Type().(*types.Signature),
		meta: &MethodMeta{
			idList:      make(IdList),
			uriIds:      make([]string, 0),
			headerIds:   make([]*ParamMeta, 0),
			cookieIds:   make([]*ParamMeta, 0),
			totalIds:    make(map[string]*ParamMeta),
			bodyIds:     make(map[string]*ParamMeta),
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

func (method *Method) resolveMetadata() (err error) {
	NewProcessor(method.commentText).Scan(func(ann, key, value string) (err error) {
		switch ann {
		case GetAnn:

		case HeadAnn:
		case PostAnn:
		case PutAnn:
		case PatchAnn:
		case DeleteAnn:
		case ConnectAnn:
		case OptionsAnn:
		case TraceAnn:
		case BodyAnn:
		case ParamAnn:
		case OnlyBodyAnn:
		case HeaderAnn:
		case CookieAnn:
		case FileAnn:
		}
		return
	})
	return
}

func (method *Method) TrySetMethod(httpMethod, uriPattern string) (err error) {

	return
}

//func isSliceOf(sliceType types.Type, kind types.BasicKind) (yes bool) {
//	if slice, ok := sliceType.(*types.Slice); ok {
//		yes = isBasic(slice.Elem(), kind)
//	}
//	return
//}
//
//func isBasic(typ types.Type, kind types.BasicKind) (yes bool) {
//	if basicType, ok := typ.(*types.Basic); ok {
//		if basicType.Kind() == kind {
//			yes = true
//		}
//	}
//	return
//}
//
//func isError(typ types.Type) (yes bool) {
//	if typ, ok := typ.Underlying().(*types.Interface); ok {
//		if typ.NumMethods() == 1 {
//			method := typ.Method(0)
//			if method.Name() == ErrorName {
//				if signature, ok := method.Type().(*types.Signature); ok {
//					params := signature.Params()
//					results := signature.Results()
//					if params
//				}
//			}
//		}
//	}
//	return
//}
