package breakingcheck

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	// main dir. in could be current dir
	//rootProjectDir = "."
	rootProjectDir = "/mnt/ssd_storage/Projects/Hound/easyp/proto-experiments"
	// path where porot files are stored
	//path = "no_deps"
	path = "."
	//path = "iam"

	//
	againstBranchName = "master"
)

func EntryPoint() {
	ctx := context.Background()
	slog.Info("Entry point breaking check")

	resultDit := filepath.Join(rootProjectDir, path)

	// read against proto files and current
	againstProtoFiles, err := ReadAgainstProtoFiles(ctx, againstBranchName, rootProjectDir, path)
	//againtProtoFiles, err := ReadAgainstProtoFiles(ctx, againstBranchName, rootProjectDir, ".")
	if err != nil {
		panic(fmt.Sprintf("failed to read against proto files: %v", err))
	}
	_ = againstProtoFiles

	currentProtoFiles, err := ReadCurrentProtoFiles(ctx, resultDit)
	if err != nil {
		panic(fmt.Sprintf("failed to read current proto files: %v", err))
	}
	_ = currentProtoFiles

	againstProtoData, err := Collect(againstProtoFiles)
	if err != nil {
		panic(fmt.Sprintf("failed to collect against proto files: %v", err))
	}
	_ = againstProtoData

	currentProtoData, err := Collect(currentProtoFiles)
	if err != nil {
		panic(fmt.Sprintf("failed to collect current proto files: %v", err))
	}
	_ = currentProtoData

	breakingCheck := &BreakingCheck{
		against: againstProtoData,
		current: currentProtoData,
	}
	issues, err := breakingCheck.Check()
	if err != nil {
		panic(fmt.Sprintf("failed to check against proto files: %v", err))
	}
	if err := printIssues(
		os.Stdout,
		issues,
	); err != nil {
		panic(fmt.Errorf("printLintErrors: %w", err))
	}

}
