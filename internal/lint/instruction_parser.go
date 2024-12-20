package lint

import (
	"strings"
)

// instructionInfo collects info about instruction in proto file
// e.g `google.api.http`:
//
//	`google.api` - package name
//	'http' - instruction name
type InstructionInfo struct {
	PkgName     string
	Instruction string
}

// parseInstruction parse input string and return its package name
// if passed input does not have package -> return pkgName as package name source proto file
type InstructionParser struct {
	SourcePkgName string
}

func (p InstructionParser) Parse(input string) InstructionInfo {
	// check if there is brackets, and extract
	// (google.api.http) -> google.api.http
	// (buf.validate.field).string.uuid -> buf.validate.field
	// or pkg.FieldType -> pkg.FieldType
	iStart := strings.Index(input, "(")
	iEnd := strings.Index(input, ")")
	if iStart != -1 && iEnd != -1 {
		input = input[iStart+1 : iEnd]
	}

	idx := strings.LastIndex(input, ".")
	if idx <= 0 {
		return InstructionInfo{
			PkgName:     p.SourcePkgName,
			Instruction: input,
		}
	}

	return InstructionInfo{
		PkgName:     input[:idx],
		Instruction: input[idx+1:],
	}
}
