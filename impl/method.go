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
	}
)
