package api

import (
	"context"
	"log"

	"github.com/easyp-tech/easyp/internal/mod"
)

// TEMO FUNCTION FOR DEBUG
func modTst() {
	log.Printf("Start")

	module := "github.com/googleapis/googleapis"

	getCommand := &mod.GetCommand{}

	err := getCommand.Get(context.Background(), module)
	if err != nil {
		log.Fatalf("getCommand.Get: %v", err)
	}
}
