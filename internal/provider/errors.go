package provider

import (
	"errors"

	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

// errorAsNotFound 包装 errors.As，便于在所有 resource 中复用 404 判定
func errorAsNotFound(err error, target **ocpclient.NotFoundError) bool {
	return errors.As(err, target)
}
