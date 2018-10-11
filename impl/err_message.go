package impl

import "errors"

const (
	DuplicatedAnnotation = "duplicated annotation"
	DuplicatedHttpMethod = "duplicated http method"
)

func DuplicatedAnnotationError(ann string) error {
	return errors.New(DuplicatedAnnotation + ": " + ann)
}

func DuplicatedHttpMethodError(method string) error {
	return errors.New(DuplicatedHttpMethod + ": " + method)
}
