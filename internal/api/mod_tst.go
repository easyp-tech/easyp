package api

import (
	"context"
	"log"

	"github.com/easyp-tech/easyp/internal/mod/repo"
)

// TEMO FUNCTION FOR DEBUG
func modTst() {
	log.Printf("Start")

	result, err := repo.ExecuteCommand(
		context.Background(), "/mnt/ssd_storage/Projects/Hound/easyp/easyp", "vsvdfasdf", "-la",
	)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}

	log.Printf("result: %v", result)
}
