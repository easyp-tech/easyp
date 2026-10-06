package gitmodules

import v1 "github.com/easyp-tech/easyp/internal/config/v1"

type v1ModuleCheckout struct {
	dir      string
	module   v1.Module
	commit   string
	snapshot string
}
