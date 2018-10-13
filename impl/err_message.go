package impl

import (
	"errors"
	"fmt"
)

const (
	DuplicatedAnnotation           = "duplicated annotation"
	DuplicatedHttpMethod           = "duplicated http method"
	IdNotExist                     = "id does not exist"
	PatternIdTypeMustBeIntOrString = "id in pattern must be int or string"
	PatternKeyMustNotBeEmpty       = "key of pattern must not be empty"
	UnsupportedAnnotationValue     = "annotation value is unsupported"
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

func PatternIdTypeMustBeIntOrStringError(id string) error {
	return errors.New(PatternIdTypeMustBeIntOrString + ": " + id)
}

func UnsupportedAnnotationValueError(ann, value string) error {
	return errors.New(UnsupportedAnnotationValue + fmt.Sprintf(": %s %s", ann, value))
}
