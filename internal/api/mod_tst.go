package api

import (
	"context"
	"log"

	"github.com/easyp-tech/easyp/internal/mod/commands"
	dirs2 "github.com/easyp-tech/easyp/internal/mod/dirs"
)

// TEMO FUNCTION FOR DEBUG
func modTst() {
	log.Printf("Start")

	module := "github.com/googleapis/googleapis"

	dirs := dirs2.New("/tmp/tmp.zYICI6g0Nv/cache")
	cmds := commands.New(dirs)

	err := cmds.Get(context.Background(), module)
	if err != nil {
		log.Fatalf("getCommand.Get: %v", err)
	}
}
