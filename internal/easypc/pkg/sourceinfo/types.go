package sourceinfo

// Location represents a source code location in a .proto file matching descriptorpb.SourceCodeInfo_Location.
type Location struct {
	Path                    []int32
	Span                    [4]int32
	SpanLen                 uint8 // 3 for single-line, 4 for multi-line
	LeadingComments         string
	TrailingComments        string
	LeadingDetachedComments []string
}

// FileSourceInfo holds the generated source code locations for an ASTFile.
type FileSourceInfo struct {
	Locations []Location
}
