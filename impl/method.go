package impl

import (
	"go/types"
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
		httpMethod   string
		pathPattern  string
		pathIds      map[string]*ParamMeta
		queryIds     map[string]*ParamMeta
		bodyIds      map[string]*ParamMeta
		responseIds  map[string]*ParamMeta
		headerIds    map[string]string
		cookieIds    map[string]string
		responseType BodyType
		requestType  BodyType
	}

	ParamMeta struct {
		Key  string
		Type string
	}
)

func NewMethod(srv *Service, rawMethod *types.Func) *Method {
	return &Method{
		Func:      rawMethod,
		service:   srv,
		signature: rawMethod.Type().(*types.Signature),
		meta: &MethodMeta{
			pathIds:     make(map[string]*ParamMeta),
			queryIds:    make(map[string]*ParamMeta),
			bodyIds:     make(map[string]*ParamMeta),
			responseIds: make(map[string]*ParamMeta),
			headerIds:   make(map[string]string),
			cookieIds:   make(map[string]string),
		},
	}
}

func (method *Method) resolveMetadata() (err error) {
	return
}