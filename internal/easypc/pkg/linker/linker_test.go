package linker_test

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

func parseTestFile(t *testing.T, path string, content string) *parser.ASTFile {
	t.Helper()
	src := []byte(content)
	ts := lexer.NewTokenStream(len(src)/8+16, len(src)/32+16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed for %s: %v", path, err)
	}
	p := parser.NewParser(src, ts, path)
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed for %s: %v", path, err)
	}
	return ast
}

func TestLinkerVisibilityAndResolution(t *testing.T) {
	// File 1: defines an enum
	f1 := parseTestFile(t, "common/enums.proto", `
syntax = "proto3";
package common;

enum Status {
    STATUS_UNSPECIFIED = 0;
    STATUS_ACTIVE = 1;
}
`)

	// File 2: publicly re-exports File 1
	f2 := parseTestFile(t, "common/all.proto", `
syntax = "proto3";
package common;

import public "common/enums.proto";
`)

	// File 3: imports File 2 (transitive public visibility to File 1)
	f3 := parseTestFile(t, "service/app.proto", `
syntax = "proto3";
package service;

import "common/all.proto";

message Account {
    common.Status status = 1;
}
`)

	// File 4: does NOT import anything
	f4 := parseTestFile(t, "other/isolated.proto", `
syntax = "proto3";
package other;

message Isolated {
    common.Status status = 1;
}
`)

	files := []*parser.ASTFile{f1, f2, f3, f4}
	fileMap := map[string]*parser.ASTFile{
		"common/enums.proto":    f1,
		"common/all.proto":      f2,
		"service/app.proto":     f3,
		"other/isolated.proto":  f4,
	}

	table := linker.NewSymbolTable(files, fileMap)
	table.ResolveTypes()

	f1ID := table.FileIDs["common/enums.proto"]
	f2ID := table.FileIDs["common/all.proto"]
	f3ID := table.FileIDs["service/app.proto"]
	f4ID := table.FileIDs["other/isolated.proto"]

	// Test visibility bitsets
	if !table.Vis.IsVisible(f3ID, f1ID) {
		t.Errorf("expected f3 (service/app.proto) to see f1 (common/enums.proto) via transitive public import")
	}
	if !table.Vis.IsVisible(f3ID, f2ID) {
		t.Errorf("expected f3 to see f2")
	}
	if table.Vis.IsVisible(f4ID, f1ID) {
		t.Errorf("expected f4 to NOT see f1")
	}

	// Verify f3 Account.status field was resolved to TypeEnum
	if len(f3.Messages) != 1 {
		t.Fatalf("expected 1 message in f3, got %d", len(f3.Messages))
	}
	accMsg := &f3.Messages[0]
	if accMsg.Fields.Count != 1 {
		t.Fatalf("expected 1 field in Account, got %d", accMsg.Fields.Count)
	}
	statusFld := &f3.Fields[f3.ChildIndices[accMsg.Fields.Start]]
	if statusFld.TypeKind != parser.TypeEnum {
		t.Errorf("expected status field to have TypeEnum, got %v", statusFld.TypeKind)
	}
	if statusFld.TypeName.String(f3.Src) != ".common.Status" {
		t.Errorf("expected typeName .common.Status, got %q", statusFld.TypeName.String(f3.Src))
	}
}

func TestCardImageFillResolution(t *testing.T) {
	fImage := parseTestFile(t, "google/actions/sdk/v2/conversation/prompt/content/image.proto", `
syntax = "proto3";
package google.actions.sdk.v2.conversation;

message Image {
  enum ImageFill {
    UNSPECIFIED = 0;
    GRAY = 1;
  }
  string url = 1;
}
`)

	fCard := parseTestFile(t, "google/actions/sdk/v2/conversation/prompt/content/card.proto", `
syntax = "proto3";
package google.actions.sdk.v2.conversation;

import "google/actions/sdk/v2/conversation/prompt/content/image.proto";

message Card {
  Image.ImageFill image_fill = 5;
}
`)

	files := []*parser.ASTFile{fImage, fCard}
	fileMap := map[string]*parser.ASTFile{
		"google/actions/sdk/v2/conversation/prompt/content/image.proto": fImage,
		"google/actions/sdk/v2/conversation/prompt/content/card.proto":  fCard,
	}

	table := linker.NewSymbolTable(files, fileMap)
	table.ResolveTypes()

	cardMsg := &fCard.Messages[0]
	fld := &fCard.Fields[fCard.ChildIndices[cardMsg.Fields.Start]]
	t.Logf("fld.TypeKind = %v, fld.TypeName = %q", fld.TypeKind, fld.TypeName.String(fCard.Src))
	if fld.TypeKind != parser.TypeEnum {
		t.Errorf("expected TypeEnum, got %v", fld.TypeKind)
	}
	if fld.TypeName.String(fCard.Src) != ".google.actions.sdk.v2.conversation.Image.ImageFill" {
		t.Errorf("expected .google.actions.sdk.v2.conversation.Image.ImageFill, got %q", fld.TypeName.String(fCard.Src))
	}
}
