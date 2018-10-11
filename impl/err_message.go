package impl

import "errors"

const (
	DuplicatedAnnotation = "duplicated annotation"
)

func DuplicatedAnnotationError(ann string) error {
	return errors.New(DuplicatedAnnotation + ": " + ann)
}
