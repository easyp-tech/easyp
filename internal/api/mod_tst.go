package api

import (
	"context"
	"log"

	"github.com/easyp-tech/easyp/internal/mod"
)

// TEMO FUNCTION FOR DEBUG
func modTst() {
	log.Printf("Start")

	result, err := mod.ExecuteCommand(
		context.Background(), "/mnt/ssd_storage/Projects/Hound/easyp/easyp", "ls", "-la",
	)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	log.Printf("result: %v", result)

	cacheDir, err := mod.CreateCacheDir("test")
	if err != nil {
		log.Fatalf("CreateCacheDir: %v", err)
	}
	log.Printf("CacheDir: %s", cacheDir)
}
