package impl

import "errors"

const (
	DuplicatedAnnotation        = "duplicated annotation"
	DuplicatedHttpMethod        = "duplicated http method"
	IdNotExist                  = "id does not exist"
	PathIdTypeMustBeIntOrString = "id in path must be int or string"
)

func DuplicatedAnnotationError(ann string) error {
	return errors.New(DuplicatedAnnotation + ": " + ann)
}

func DuplicatedHttpMethodError(method string) error {
	return errors.New(DuplicatedHttpMethod + ": " + method)
}

func IdNotExistError(id string) error {
	return errors.New(IdNotExist + ": " + id)
}

func PathIdTypeMustBeIntOrStringError(id string) error {
	return errors.New(PathIdTypeMustBeIntOrString + ": " + id)
}
