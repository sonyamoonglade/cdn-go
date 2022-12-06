package validate

import (
	"errors"
	"fmt"
	"strings"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/internal/entities"

	"github.com/go-playground/validator/v10"
)

var v validator.Validate

func init() {
	v = *validator.New()
}

func BucketOperation(ops []*entities.Operation) error {
	for _, op := range ops {
		if op.Name != cdn_go.OperationGet && op.Name != cdn_go.OperationPost && op.Name != cdn_go.OperationDelete {
			return fmt.Errorf("validation error: invalid operation %s", op.Name)
		}
		if op.Type != cdn_go.OperationTypePrivate && op.Type != cdn_go.OperationTypePublic {
			return fmt.Errorf("validation error: invalid type %s", op.Type)
		}
	}

	return nil
}

func ValidateRequiredFields(dto interface{}) error {
	err := v.Struct(dto)
	if err == nil {
		return nil
	}
	var buff string = "validation error: "
	errs := err.(validator.ValidationErrors)
	for _, e := range errs {
		buff += fmt.Sprintf("field '%s' is missing in request body", strings.ToLower(e.Field()))
	}
	return errors.New(buff)
}
