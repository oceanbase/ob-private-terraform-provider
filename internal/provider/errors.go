package provider

import (
	"errors"

	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

// errorAsNotFound wraps errors.As for reusable 404 detection across resources
func errorAsNotFound(err error, target **ocpclient.NotFoundError) bool {
	return errors.As(err, target)
}
