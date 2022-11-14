package validate

import (
	"fmt"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/internal/entities"
)

func BucketOperation(ops []entities.Operation) error {
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
