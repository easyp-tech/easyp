package api

import (
	"context"
	"log"

	modPkg "github.com/easyp-tech/easyp/internal/package_manager/mod"
	dirs2 "github.com/easyp-tech/easyp/internal/package_manager/services/dirs"
)

// TEMO FUNCTION FOR DEBUG
func modTst() {
	log.Printf("Start")

	module := "github.com/googleapis/googleapis"
	// module := "github.com/googleapis/googleapis@0e50601ea3d1f828a90d2ddbd52920fcafd461fd"
	// module := "github.com/googleapis/googleapis@common-protos-1_3_1"
	// module := "github.com/googleapis/googleapis@0e50601ea3d1f828a90d2ddbd52920fcafd461fd111"

	// module := "github.com/bufbuild/protovalidate"
	// module := "github.com/bufbuild/protovalidate@v0.3.1"
	// module := "github.com/bufbuild/protovalidate@tools/v0.3.1"

	dirs := dirs2.New("/tmp/tmp.zYICI6g0Nv/cache")
	mod := modPkg.New(dirs)

	err := mod.Get(context.Background(), module)
	if err != nil {
		log.Fatalf("getCommand.Get: %v", err)
	}
}
