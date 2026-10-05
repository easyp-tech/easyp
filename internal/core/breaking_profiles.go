package core

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/wellknownimports"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

type breakingGraph struct {
	files    map[string]protoreflect.FileDescriptor
	targets  map[string]bool
	paths    map[string]string
	messages map[protoreflect.FullName]protoreflect.MessageDescriptor
	enums    map[protoreflect.FullName]protoreflect.EnumDescriptor
	services map[protoreflect.FullName]protoreflect.ServiceDescriptor
	methods  map[protoreflect.FullName]protoreflect.MethodDescriptor
	ext      map[protoreflect.FullName]protoreflect.FieldDescriptor
}

type breakingProfiles struct {
	file, source, wireJSON, wire bool
}

func selectedBreakingProfiles(categories []string) (breakingProfiles, error) {
	if err := config.ValidateBreakingCategories(categories); err != nil {
		return breakingProfiles{}, err
	}
	selected := breakingProfiles{}
	for _, category := range categories {
		switch category {
		case "FILE":
			selected.file, selected.source, selected.wireJSON, selected.wire = true, true, true, true
		case "PACKAGE":
			selected.source, selected.wireJSON, selected.wire = true, true, true
		case "WIRE_JSON":
			selected.wireJSON, selected.wire = true, true
		case "WIRE":
			selected.wire = true
		}
	}
	return selected, nil
}

func (c *Core) compareBreakingProfiles(
	ctx context.Context, current, against DirWalker, baseline *Core, profiles breakingProfiles,
) ([]IssueInfo, error) {
	currentGraph, err := c.compileBreakingGraph(ctx, current)
	if err != nil {
		return nil, fmt.Errorf("compile current descriptors: %w", err)
	}
	againstGraph, err := baseline.compileBreakingGraph(ctx, against)
	if err != nil {
		return nil, fmt.Errorf("compile baseline descriptors: %w", err)
	}
	if c.breakingCheckConfig.IgnoreUnstable {
		filterUnstableBreakingGraph(currentGraph)
		filterUnstableBreakingGraph(againstGraph)
	}
	return compareBreakingGraphs(againstGraph, currentGraph, profiles), nil
}

func (c *Core) compileBreakingGraph(ctx context.Context, walker DirWalker) (*breakingGraph, error) {
	importRoots, ignoreDirs := c.importRoots, c.breakingCheckConfig.IgnoreDirs
	rooted, ok := walker.(interface{ RootPath() string })
	if !ok {
		return nil, fmt.Errorf("breaking source does not expose its root path")
	}
	root := rooted.RootPath()
	targets := make([]string, 0)
	targetPaths := make(map[string]string)
	physical := make(map[string]bool)
	err := walker.WalkDir(func(path string, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filepath.Ext(path) != ".proto" || path_helpers.IsIgnoredPath(path, ignoreDirs) {
			return nil
		}
		canonicalPath, err := canonicalBreakingPath(root, path, importRoots)
		if err != nil {
			return err
		}
		physicalPath := filepath.Join(root, filepath.FromSlash(path))
		resolvedPath, err := filepath.EvalSymlinks(physicalPath)
		if err == nil {
			physicalPath = resolvedPath
		}
		physicalPath = filepath.Clean(physicalPath)
		if physical[physicalPath] {
			return nil
		}
		physical[physicalPath] = true
		targets = append(targets, canonicalPath)
		targetPaths[canonicalPath] = filepath.ToSlash(path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir: %w", err)
	}
	slices.Sort(targets)
	if len(targets) == 0 {
		return newBreakingGraph(), nil
	}
	imports := uniquePhysicalRoots(importRoots)
	if len(imports) == 0 {
		imports = []string{root}
	}
	compiler := protocompile.Compiler{
		Resolver:       wellknownimports.WithStandardImports(&protocompile.SourceResolver{ImportPaths: imports, Accessor: c.openSourceFile}),
		SourceInfoMode: protocompile.SourceInfoStandard,
	}
	compiled, err := compiler.Compile(ctx, targets...)
	if err != nil {
		return nil, fmt.Errorf("Compile: %w", err)
	}
	return buildBreakingGraph(compiled, targetPaths), nil
}

func canonicalBreakingPath(root, path string, importRoots []string) (string, error) {
	physicalPath := filepath.Join(root, filepath.FromSlash(path))
	for _, importRoot := range importRoots {
		relative, err := filepath.Rel(importRoot, physicalPath)
		if err == nil && filepath.IsLocal(relative) {
			return filepath.ToSlash(relative), nil
		}
	}
	return filepath.ToSlash(path), nil
}

func uniquePhysicalRoots(roots []string) []string {
	result := make([]string, 0, len(roots))
	seen := make(map[string]bool)
	for _, root := range roots {
		resolved, err := filepath.EvalSymlinks(root)
		if err == nil {
			root = resolved
		}
		root = filepath.Clean(root)
		if !seen[root] {
			seen[root] = true
			result = append(result, root)
		}
	}
	return result
}

func newBreakingGraph() *breakingGraph {
	return &breakingGraph{
		files: make(map[string]protoreflect.FileDescriptor), targets: make(map[string]bool), paths: make(map[string]string),
		messages: make(map[protoreflect.FullName]protoreflect.MessageDescriptor),
		enums:    make(map[protoreflect.FullName]protoreflect.EnumDescriptor),
		services: make(map[protoreflect.FullName]protoreflect.ServiceDescriptor),
		methods:  make(map[protoreflect.FullName]protoreflect.MethodDescriptor),
		ext:      make(map[protoreflect.FullName]protoreflect.FieldDescriptor),
	}
}

func buildBreakingGraph(files linker.Files, targetPaths map[string]string) *breakingGraph {
	graph := newBreakingGraph()
	for name := range targetPaths {
		graph.targets[name] = true
	}
	seen := make(map[protoreflect.FileDescriptor]bool)
	var visit func(protoreflect.FileDescriptor)
	visit = func(file protoreflect.FileDescriptor) {
		if seen[file] {
			return
		}
		seen[file] = true
		graph.files[string(file.Path())] = file
		if path, ok := targetPaths[string(file.Path())]; ok {
			graph.paths[string(file.Path())] = path
		} else {
			graph.paths[string(file.Path())] = filepath.ToSlash(file.Path())
		}
		for i := range file.Imports().Len() {
			visit(file.Imports().Get(i))
		}
		for i := range file.Messages().Len() {
			addBreakingMessage(graph, file.Messages().Get(i))
		}
		for i := range file.Enums().Len() {
			graph.enums[file.Enums().Get(i).FullName()] = file.Enums().Get(i)
		}
		for i := range file.Services().Len() {
			service := file.Services().Get(i)
			graph.services[service.FullName()] = service
			for j := range service.Methods().Len() {
				method := service.Methods().Get(j)
				graph.methods[method.FullName()] = method
			}
		}
		for i := range file.Extensions().Len() {
			graph.ext[file.Extensions().Get(i).FullName()] = file.Extensions().Get(i)
		}
	}
	for _, file := range files {
		visit(file)
	}
	return graph
}

func addBreakingMessage(graph *breakingGraph, message protoreflect.MessageDescriptor) {
	graph.messages[message.FullName()] = message
	for i := range message.Extensions().Len() {
		extension := message.Extensions().Get(i)
		graph.ext[extension.FullName()] = extension
	}
	for i := range message.Enums().Len() {
		enum := message.Enums().Get(i)
		graph.enums[enum.FullName()] = enum
	}
	for i := range message.Messages().Len() {
		addBreakingMessage(graph, message.Messages().Get(i))
	}
}

func filterUnstableBreakingGraph(graph *breakingGraph) {
	unstable := func(d protoreflect.Descriptor) bool { return isUnstablePackage(PackageName(d.ParentFile().Package())) }
	for name, d := range graph.messages {
		if unstable(d) {
			delete(graph.messages, name)
		}
	}
	for name, d := range graph.enums {
		if unstable(d) {
			delete(graph.enums, name)
		}
	}
	for name, d := range graph.services {
		if unstable(d) {
			delete(graph.services, name)
		}
	}
	for name, d := range graph.methods {
		if unstable(d) {
			delete(graph.methods, name)
		}
	}
	for name, d := range graph.ext {
		if unstable(d) {
			delete(graph.ext, name)
		}
	}
	for name, d := range graph.files {
		if isUnstablePackage(PackageName(d.Package())) {
			delete(graph.files, name)
			delete(graph.paths, name)
		}
	}
}

func compareBreakingGraphs(against, current *breakingGraph, profile breakingProfiles) []IssueInfo {
	var issues []IssueInfo
	appendIssue := func(descriptor protoreflect.Descriptor, rule, message string) {
		if descriptor == nil {
			return
		}
		file := descriptor.ParentFile()
		location := file.SourceLocations().ByDescriptor(descriptor)
		path := against.paths[string(file.Path())]
		if current.files[string(file.Path())] == file {
			path = current.paths[string(file.Path())]
		}
		if path == "" {
			path = filepath.ToSlash(file.Path())
		}
		issues = append(issues, IssueInfo{
			Path: path,
			Issue: Issue{RuleName: rule, Message: message, Position: meta.Position{
				Filename: path, Line: location.StartLine + 1, Column: location.StartColumn + 1,
			}},
		})
	}

	if profile.file {
		for name, oldFile := range against.files {
			if _, ok := current.files[name]; !ok && isTargetBreakingFile(against, name) {
				appendIssue(oldFile, "FILE_NO_DELETE", fmt.Sprintf("File %q was deleted.", name))
			}
		}
	}
	compareBreakingPackages(against, current, profile, appendIssue)
	compareBreakingMessages(against, current, profile, appendIssue)
	compareBreakingEnums(against, current, profile, appendIssue)
	compareBreakingServices(against, current, profile, appendIssue)
	compareBreakingExtensions(against, current, profile, appendIssue)
	compareBreakingFileOptions(against, current, profile, appendIssue)
	return deduplicateBreakingIssues(issues)
}

func isTargetBreakingFile(graph *breakingGraph, name string) bool {
	_, isTarget := graph.targets[name]
	return isTarget
}

func compareBreakingPackages(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	if !profile.source {
		return
	}
	packages := make(map[protoreflect.FullName]bool)
	for _, file := range against.files {
		if file.Package() != "" {
			packages[file.Package()] = true
		}
	}
	for pkg := range packages {
		if !hasPackage(current, pkg) {
			var source protoreflect.Descriptor
			for _, file := range against.files {
				if file.Package() == pkg {
					if source == nil || file.Path() < source.ParentFile().Path() {
						source = file
					}
				}
			}
			report(source, "PACKAGE_NO_DELETE", fmt.Sprintf("Package %q was deleted.", pkg))
		}
	}
}

func hasPackage(graph *breakingGraph, pkg protoreflect.FullName) bool {
	for _, file := range graph.files {
		if file.Package() == pkg {
			return true
		}
	}
	return false
}

func compareBreakingMessages(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	for name, oldMessage := range against.messages {
		newMessage := current.messages[name]
		if newMessage == nil {
			if profile.source && !oldMessage.IsMapEntry() {
				report(oldMessage, "MESSAGE_NO_DELETE", fmt.Sprintf("Message %q was deleted.", name))
			}
			continue
		}
		if profile.file && oldMessage.ParentFile().Path() != newMessage.ParentFile().Path() {
			report(oldMessage, "MESSAGE_MOVED", fmt.Sprintf("Message %q was moved from %q to %q.", name, against.paths[oldMessage.ParentFile().Path()], current.paths[newMessage.ParentFile().Path()]))
		}
		compareMessageReserved(oldMessage, newMessage, report)
		compareExtensionRanges(oldMessage, newMessage, profile, report)
		compareRequiredFields(oldMessage, newMessage, report)
		compareMessageOptions(oldMessage, newMessage, profile, report)
		if jsonProfileEnabled(profile) && !sameJSONFormat(oldMessage, newMessage) {
			report(oldMessage, "MESSAGE_SAME_JSON_FORMAT", fmt.Sprintf("Message %q changed JSON format support.", name))
		}
		compareBreakingFields(oldMessage, newMessage, profile, report)
		compareOneofs(oldMessage, newMessage, profile, report)
	}
}

func compareBreakingFields(oldMessage, newMessage protoreflect.MessageDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	for i := range oldMessage.Fields().Len() {
		oldField := oldMessage.Fields().Get(i)
		newField := newMessage.Fields().ByNumber(oldField.Number())
		if newField == nil {
			if profile.source {
				report(oldField, "FIELD_NO_DELETE", fmt.Sprintf("Field %q (%d) on %q was deleted.", oldField.Name(), oldField.Number(), oldMessage.FullName()))
			} else if profile.wire || profile.wireJSON {
				if !newMessage.ReservedRanges().Has(oldField.Number()) {
					report(oldField, "FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED", fmt.Sprintf("Deleted field %q number %d on %q must be reserved.", oldField.Name(), oldField.Number(), oldMessage.FullName()))
				}
				if profile.wireJSON && !newMessage.ReservedNames().Has(oldField.Name()) {
					report(oldField, "FIELD_NO_DELETE_UNLESS_NAME_RESERVED", fmt.Sprintf("Deleted field %q on %q must have its name reserved for JSON compatibility.", oldField.Name(), oldMessage.FullName()))
				}
			}
			continue
		}
		compareField(oldField, newField, profile, report)
	}
}

func compareField(oldField, newField protoreflect.FieldDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	fieldName := fmt.Sprintf("Field %q (%d)", oldField.Name(), oldField.Number())
	if profile.source {
		if oldField.Name() != newField.Name() {
			report(oldField, "FIELD_SAME_NAME", fmt.Sprintf("%s on %q changed name to %q.", fieldName, oldField.ContainingMessage().FullName(), newField.Name()))
		}
		if oldField.Kind() != newField.Kind() || !sameResolvedType(oldField, newField) {
			report(oldField, "FIELD_SAME_TYPE", fmt.Sprintf("%s on %q changed type from %s to %s.", fieldName, oldField.ContainingMessage().FullName(), fieldTypeName(oldField), fieldTypeName(newField)))
		}
		if fieldCardinality(oldField) != fieldCardinality(newField) {
			report(oldField, "FIELD_SAME_CARDINALITY", fmt.Sprintf("%s on %q changed cardinality from %s to %s.", fieldName, oldField.ContainingMessage().FullName(), fieldCardinality(oldField), fieldCardinality(newField)))
		}
		if oldField.Name() == newField.Name() && oldField.JSONName() != newField.JSONName() {
			report(oldField, "FIELD_SAME_JSON_NAME", fmt.Sprintf("%s changed JSON name from %q to %q.", fieldName, oldField.JSONName(), newField.JSONName()))
		}
		if generatedOneof(oldField) != generatedOneof(newField) {
			report(oldField, "FIELD_SAME_ONEOF", fmt.Sprintf("%s changed real oneof membership.", fieldName))
		}
		compareFieldCodegenOptions(oldField, newField, report)
	} else if profile.wireJSON || profile.wire {
		if profile.wireJSON && oldField.Name() != newField.Name() {
			report(oldField, "FIELD_SAME_NAME", fmt.Sprintf("%s changed name from %q to %q.", fieldName, oldField.Name(), newField.Name()))
		}
		if profile.wireJSON && oldField.Name() == newField.Name() && oldField.JSONName() != newField.JSONName() {
			report(oldField, "FIELD_SAME_JSON_NAME", fmt.Sprintf("%s changed JSON name from %q to %q.", fieldName, oldField.JSONName(), newField.JSONName()))
		}
		if !wireFieldTypesCompatible(oldField, newField, profile.wireJSON) {
			rule := "FIELD_WIRE_COMPATIBLE_TYPE"
			if profile.wireJSON {
				rule = "FIELD_WIRE_JSON_COMPATIBLE_TYPE"
			}
			report(oldField, rule, fmt.Sprintf("%s on %q changed incompatibly from %s to %s.", fieldName, oldField.ContainingMessage().FullName(), fieldTypeName(oldField), fieldTypeName(newField)))
		}
		oldCardinality, newCardinality := fieldCardinality(oldField), fieldCardinality(newField)
		if oldCardinality != newCardinality && !wireCompatibleCardinality(oldField, newField, profile.wireJSON) {
			rule := "FIELD_WIRE_COMPATIBLE_CARDINALITY"
			if profile.wireJSON {
				rule = "FIELD_WIRE_JSON_COMPATIBLE_CARDINALITY"
			}
			report(oldField, rule, fmt.Sprintf("%s on %q changed cardinality from %s to %s.", fieldName, oldField.ContainingMessage().FullName(), oldCardinality, newCardinality))
		}
		if generatedOneof(oldField) != generatedOneof(newField) {
			report(oldField, "FIELD_SAME_ONEOF", fmt.Sprintf("%s changed real oneof membership.", fieldName))
		}
	}
	if !sameFieldDefault(oldField, newField) {
		report(oldField, "FIELD_SAME_DEFAULT", fmt.Sprintf("%s default changed.", fieldName))
	}
}

func wireFieldTypesCompatible(oldField, newField protoreflect.FieldDescriptor, json bool) bool {
	return wireCompatibleType(oldField, newField, json, make(map[typePair]bool))
}

func sameResolvedType(oldField, newField protoreflect.FieldDescriptor) bool {
	if oldField.Kind() != newField.Kind() {
		return false
	}
	switch oldField.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return oldField.Message().FullName() == newField.Message().FullName()
	case protoreflect.EnumKind:
		return oldField.Enum().FullName() == newField.Enum().FullName()
	default:
		return true
	}
}

func fieldTypeName(field protoreflect.FieldDescriptor) string {
	switch field.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return string(field.Message().FullName())
	case protoreflect.EnumKind:
		return string(field.Enum().FullName())
	default:
		return field.Kind().String()
	}
}

func fieldCardinality(field protoreflect.FieldDescriptor) string {
	if field.IsMap() {
		return "map"
	}
	switch field.Cardinality() {
	case protoreflect.Required:
		return "required"
	case protoreflect.Repeated:
		return "repeated"
	}
	if field.HasPresence() {
		return "optional_explicit"
	}
	return "optional_implicit"
}

func generatedOneof(field protoreflect.FieldDescriptor) protoreflect.FullName {
	oneof := field.ContainingOneof()
	if oneof == nil || oneof.IsSynthetic() {
		return ""
	}
	return oneof.FullName()
}

type typePair struct {
	oldName protoreflect.FullName
	newName protoreflect.FullName
}

func wireCompatibleType(oldField, newField protoreflect.FieldDescriptor, json bool, seen map[typePair]bool) bool {
	oldKind, newKind := oldField.Kind(), newField.Kind()
	if oldKind == newKind {
		switch oldKind {
		case protoreflect.EnumKind:
			if oldField.Enum().FullName() == newField.Enum().FullName() {
				return true
			}
			return compatibleEnumType(oldField.Enum(), newField.Enum())
		case protoreflect.MessageKind, protoreflect.GroupKind:
			if oldField.Message().FullName() == newField.Message().FullName() {
				return true
			}
			return wireCompatibleMessages(oldField.Message(), newField.Message(), json, seen)
		default:
			return true
		}
	}
	if oldKind == protoreflect.StringKind && newKind == protoreflect.BytesKind {
		return !json
	}
	if json {
		return sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Int32Kind, protoreflect.Uint32Kind}) ||
			sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Int64Kind, protoreflect.Uint64Kind}) ||
			sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind}) ||
			sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind})
	}
	return sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.BoolKind}) ||
		sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Sint32Kind, protoreflect.Sint64Kind}) ||
		sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind}) ||
		sameScalarFamily(oldKind, newKind, []protoreflect.Kind{protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind})
}

func sameScalarFamily(oldKind, newKind protoreflect.Kind, kinds []protoreflect.Kind) bool {
	return slices.Contains(kinds, oldKind) && slices.Contains(kinds, newKind)
}

func compatibleEnumType(oldEnum, newEnum protoreflect.EnumDescriptor) bool {
	if oldEnum.Name() != newEnum.Name() {
		return false
	}
	newPairs := enumPairs(newEnum)
	for pair := range enumPairs(oldEnum) {
		if !newPairs[pair] {
			return false
		}
	}
	return true
}

type enumPair struct {
	name   protoreflect.Name
	number protoreflect.EnumNumber
}

func enumPairs(enum protoreflect.EnumDescriptor) map[enumPair]bool {
	pairs := make(map[enumPair]bool)
	for i := range enum.Values().Len() {
		value := enum.Values().Get(i)
		pairs[enumPair{name: value.Name(), number: value.Number()}] = true
	}
	return pairs
}

func wireCompatibleMessages(oldMessage, newMessage protoreflect.MessageDescriptor, json bool, seen map[typePair]bool) bool {
	pair := typePair{oldName: oldMessage.FullName(), newName: newMessage.FullName()}
	if seen[pair] {
		return true
	}
	seen[pair] = true
	if !rangesPreserved(oldMessage.ReservedRanges(), newMessage.ReservedRanges()) {
		return false
	}
	for i := range oldMessage.ReservedNames().Len() {
		if !newMessage.ReservedNames().Has(oldMessage.ReservedNames().Get(i)) {
			return false
		}
	}
	if oldMessage.Options().(*descriptorpb.MessageOptions).GetMessageSetWireFormat() != newMessage.Options().(*descriptorpb.MessageOptions).GetMessageSetWireFormat() {
		return false
	}
	if json && !sameJSONFormat(oldMessage, newMessage) {
		return false
	}
	for i := range newMessage.Fields().Len() {
		next := newMessage.Fields().Get(i)
		old := oldMessage.Fields().ByNumber(next.Number())
		if next.Cardinality() == protoreflect.Required && (old == nil || old.Cardinality() != protoreflect.Required) {
			return false
		}
	}
	for i := range oldMessage.Fields().Len() {
		old := oldMessage.Fields().Get(i)
		next := newMessage.Fields().ByNumber(old.Number())
		if next == nil {
			if old.Cardinality() == protoreflect.Required || !newMessage.ReservedRanges().Has(old.Number()) {
				return false
			}
			if json && !newMessage.ReservedNames().Has(old.Name()) {
				return false
			}
			continue
		}
		if !wireCompatibleType(old, next, json, seen) || !wireCompatibleCardinality(old, next, json) || !sameFieldDefault(old, next) {
			return false
		}
		if json && (old.Name() != next.Name() || old.JSONName() != next.JSONName()) {
			return false
		}
		oldOne, newOne := old.ContainingOneof(), next.ContainingOneof()
		oldReal, newReal := oldOne != nil && !oldOne.IsSynthetic(), newOne != nil && !newOne.IsSynthetic()
		if oldReal != newReal || (oldReal && oldOne.Name() != newOne.Name()) {
			return false
		}
	}
	return true
}

func wireCompatibleCardinality(oldField, newField protoreflect.FieldDescriptor, json bool) bool {
	oldCardinality, newCardinality := fieldCardinality(oldField), fieldCardinality(newField)
	if oldCardinality == newCardinality || optionalPresenceCardinality(oldCardinality, newCardinality) {
		return true
	}
	if oldCardinality == "required" || newCardinality == "required" {
		return false
	}
	if !json && ((oldCardinality == "map" && newCardinality == "repeated") || (oldCardinality == "repeated" && newCardinality == "map")) {
		if oldField.Kind() != protoreflect.MessageKind || newField.Kind() != protoreflect.MessageKind {
			return false
		}
		return wireCompatibleMessages(oldField.Message(), newField.Message(), false, make(map[typePair]bool))
	}
	return false
}

func optionalPresenceCardinality(oldCardinality, newCardinality string) bool {
	return strings.HasPrefix(oldCardinality, "optional_") && strings.HasPrefix(newCardinality, "optional_")
}

func compareBreakingEnums(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	for name, oldEnum := range against.enums {
		newEnum := current.enums[name]
		if newEnum == nil {
			if profile.source {
				report(oldEnum, "ENUM_NO_DELETE", fmt.Sprintf("Enum %q was deleted.", name))
			}
			continue
		}
		if profile.file && oldEnum.ParentFile().Path() != newEnum.ParentFile().Path() {
			report(oldEnum, "ENUM_MOVED", fmt.Sprintf("Enum %q was moved from %q to %q.", name, against.paths[oldEnum.ParentFile().Path()], current.paths[newEnum.ParentFile().Path()]))
		}
		compareEnumReserved(oldEnum, newEnum, report)
		if profile.source && oldEnum.IsClosed() != newEnum.IsClosed() {
			report(oldEnum, "ENUM_SAME_TYPE", fmt.Sprintf("Enum %q changed between open and closed semantics.", name))
		}
		if jsonProfileEnabled(profile) && !sameJSONFormat(oldEnum, newEnum) {
			rule := "ENUM_SAME_JSON_FORMAT"
			report(oldEnum, rule, fmt.Sprintf("Enum %q changed JSON format support.", name))
		}
		oldValues := enumPairs(oldEnum)
		newValues := enumPairs(newEnum)
		for oldPair := range oldValues {
			if newValues[oldPair] {
				continue
			}
			matchingNumber := enumValueAtNumber(newEnum, oldPair.number)
			if matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.source || matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.wireJSON {
				report(oldEnum, "ENUM_VALUE_SAME_NAME", fmt.Sprintf("Enum value %q at number %d in %q changed name to %q.", oldPair.name, oldPair.number, name, matchingNumber.Name()))
				continue
			}
			if profile.source {
				report(oldEnum, "ENUM_VALUE_NO_DELETE", fmt.Sprintf("Enum value %q = %d in %q was deleted.", oldPair.name, oldPair.number, name))
			} else if profile.wire || profile.wireJSON {
				if matchingNumber == nil && !newEnum.ReservedRanges().Has(oldPair.number) {
					report(oldEnum, "ENUM_VALUE_NO_DELETE_UNLESS_NUMBER_RESERVED", fmt.Sprintf("Deleted enum number %d in %q must be reserved.", oldPair.number, name))
				}
				if profile.wireJSON && !newEnum.ReservedNames().Has(oldPair.name) {
					report(oldEnum, "ENUM_VALUE_NO_DELETE_UNLESS_NAME_RESERVED", fmt.Sprintf("Deleted enum name %q in %q must be reserved for JSON compatibility.", oldPair.name, name))
				}
			}
		}
	}
}

func enumValueAtNumber(enum protoreflect.EnumDescriptor, number protoreflect.EnumNumber) protoreflect.EnumValueDescriptor {
	for i := range enum.Values().Len() {
		value := enum.Values().Get(i)
		if value.Number() == number {
			return value
		}
	}
	return nil
}

func compareBreakingServices(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	for name, oldService := range against.services {
		newService := current.services[name]
		if newService == nil {
			report(oldService, "SERVICE_NO_DELETE", fmt.Sprintf("Service %q was deleted.", name))
			continue
		}
		if profile.file && oldService.ParentFile().Path() != newService.ParentFile().Path() {
			report(oldService, "SERVICE_MOVED", fmt.Sprintf("Service %q was moved from %q to %q.", name, against.paths[oldService.ParentFile().Path()], current.paths[newService.ParentFile().Path()]))
		}
		for i := range oldService.Methods().Len() {
			oldMethod := oldService.Methods().Get(i)
			newMethod := newService.Methods().ByName(oldMethod.Name())
			if newMethod == nil {
				report(oldMethod, "RPC_NO_DELETE", fmt.Sprintf("RPC %q on service %q was deleted.", oldMethod.Name(), name))
				continue
			}
			if oldMethod.Input().FullName() != newMethod.Input().FullName() {
				report(oldMethod, "RPC_SAME_REQUEST_TYPE", fmt.Sprintf("RPC %q request type changed from %q to %q.", oldMethod.Name(), oldMethod.Input().FullName(), newMethod.Input().FullName()))
			}
			if oldMethod.Output().FullName() != newMethod.Output().FullName() {
				report(oldMethod, "RPC_SAME_RESPONSE_TYPE", fmt.Sprintf("RPC %q response type changed from %q to %q.", oldMethod.Name(), oldMethod.Output().FullName(), newMethod.Output().FullName()))
			}
			if oldMethod.IsStreamingClient() != newMethod.IsStreamingClient() {
				report(oldMethod, "RPC_SAME_CLIENT_STREAMING", fmt.Sprintf("RPC %q changed client streaming behavior.", oldMethod.Name()))
			}
			if oldMethod.IsStreamingServer() != newMethod.IsStreamingServer() {
				report(oldMethod, "RPC_SAME_SERVER_STREAMING", fmt.Sprintf("RPC %q changed server streaming behavior.", oldMethod.Name()))
			}
			oldOptions := oldMethod.Options().(*descriptorpb.MethodOptions)
			newOptions := newMethod.Options().(*descriptorpb.MethodOptions)
			if oldOptions.GetIdempotencyLevel() != newOptions.GetIdempotencyLevel() {
				report(oldMethod, "RPC_SAME_IDEMPOTENCY_LEVEL", fmt.Sprintf("RPC %q changed idempotency level.", oldMethod.Name()))
			}
		}
	}
}

func compareBreakingExtensions(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	for name, old := range against.ext {
		next := current.ext[name]
		if next == nil {
			if profile.source {
				report(old, "EXTENSION_NO_DELETE", fmt.Sprintf("Extension %q was deleted.", name))
			}
			continue
		}
		if old.Number() != next.Number() {
			report(old, "EXTENSION_SAME_NUMBER", fmt.Sprintf("Extension %q changed number from %d to %d.", name, old.Number(), next.Number()))
		}
		if old.ContainingMessage().FullName() != next.ContainingMessage().FullName() {
			report(old, "EXTENSION_SAME_EXTENDEE", fmt.Sprintf("Extension %q changed the extended message.", name))
		}
		if profile.file && old.ParentFile().Path() != next.ParentFile().Path() {
			report(old, "EXTENSION_MOVED", fmt.Sprintf("Extension %q was moved between files.", name))
		}
		compareField(old, next, profile, report)
	}
}

func compareBreakingFileOptions(against, current *breakingGraph, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	seen := make(map[[2]string]bool)
	pair := func(oldFile, newFile protoreflect.FileDescriptor) {
		key := [2]string{oldFile.Path(), newFile.Path()}
		if seen[key] {
			return
		}
		seen[key] = true
		compareBreakingFilePair(oldFile, newFile, profile, report)
	}
	for name, old := range against.files {
		if next := current.files[name]; next != nil {
			pair(old, next)
		}
	}
	if !profile.source {
		return
	}
	for name, old := range against.messages {
		if next := current.messages[name]; next != nil {
			pair(old.ParentFile(), next.ParentFile())
		}
	}
	for name, old := range against.enums {
		if next := current.enums[name]; next != nil {
			pair(old.ParentFile(), next.ParentFile())
		}
	}
	for name, old := range against.services {
		if next := current.services[name]; next != nil {
			pair(old.ParentFile(), next.ParentFile())
		}
	}
	for name, old := range against.ext {
		if next := current.ext[name]; next != nil {
			pair(old.ParentFile(), next.ParentFile())
		}
	}
}

func compareBreakingFilePair(oldFile, newFile protoreflect.FileDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	name := oldFile.Path()
	if oldFile.Package() != newFile.Package() {
		report(oldFile, "FILE_SAME_PACKAGE", fmt.Sprintf("File %q changed package from %q to %q.", name, oldFile.Package(), newFile.Package()))
	}
	if !profile.source {
		return
	}
	oldOptions := oldFile.Options().(*descriptorpb.FileOptions)
	newOptions := newFile.Options().(*descriptorpb.FileOptions)
	compareOption(oldFile, oldOptions.GetGoPackage(), newOptions.GetGoPackage(), "FILE_SAME_GO_PACKAGE", "go_package", report)
	compareOption(oldFile, oldOptions.GetJavaPackage(), newOptions.GetJavaPackage(), "FILE_SAME_JAVA_PACKAGE", "java_package", report)
	compareOption(oldFile, oldOptions.GetJavaOuterClassname(), newOptions.GetJavaOuterClassname(), "FILE_SAME_JAVA_OUTER_CLASSNAME", "java_outer_classname", report)
	compareOption(oldFile, oldOptions.GetJavaMultipleFiles(), newOptions.GetJavaMultipleFiles(), "FILE_SAME_JAVA_MULTIPLE_FILES", "java_multiple_files", report)
	compareOption(oldFile, oldOptions.GetJavaStringCheckUtf8(), newOptions.GetJavaStringCheckUtf8(), "FILE_SAME_JAVA_STRING_CHECK_UTF8", "java_string_check_utf8", report)
	compareOption(oldFile, oldOptions.GetCsharpNamespace(), newOptions.GetCsharpNamespace(), "FILE_SAME_CSHARP_NAMESPACE", "csharp_namespace", report)
	compareOption(oldFile, oldOptions.GetObjcClassPrefix(), newOptions.GetObjcClassPrefix(), "FILE_SAME_OBJC_CLASS_PREFIX", "objc_class_prefix", report)
	compareOption(oldFile, oldOptions.GetOptimizeFor(), newOptions.GetOptimizeFor(), "FILE_SAME_OPTIMIZE_FOR", "optimize_for", report)
	compareOption(oldFile, oldOptions.GetPhpNamespace(), newOptions.GetPhpNamespace(), "FILE_SAME_PHP_NAMESPACE", "php_namespace", report)
	compareOption(oldFile, oldOptions.GetPhpClassPrefix(), newOptions.GetPhpClassPrefix(), "FILE_SAME_PHP_CLASS_PREFIX", "php_class_prefix", report)
	compareOption(oldFile, oldOptions.GetPhpMetadataNamespace(), newOptions.GetPhpMetadataNamespace(), "FILE_SAME_PHP_METADATA_NAMESPACE", "php_metadata_namespace", report)
	compareOption(oldFile, oldOptions.GetRubyPackage(), newOptions.GetRubyPackage(), "FILE_SAME_RUBY_PACKAGE", "ruby_package", report)
	compareOption(oldFile, oldOptions.GetSwiftPrefix(), newOptions.GetSwiftPrefix(), "FILE_SAME_SWIFT_PREFIX", "swift_prefix", report)
	compareOption(oldFile, oldOptions.GetCcEnableArenas(), newOptions.GetCcEnableArenas(), "FILE_SAME_CC_ENABLE_ARENAS", "cc_enable_arenas", report)
	compareOption(oldFile, oldOptions.GetCcGenericServices(), newOptions.GetCcGenericServices(), "FILE_SAME_CC_GENERIC_SERVICES", "cc_generic_services", report)
	compareOption(oldFile, oldOptions.GetJavaGenericServices(), newOptions.GetJavaGenericServices(), "FILE_SAME_JAVA_GENERIC_SERVICES", "java_generic_services", report)
	compareOption(oldFile, oldOptions.GetPyGenericServices(), newOptions.GetPyGenericServices(), "FILE_SAME_PY_GENERIC_SERVICES", "py_generic_services", report)
	if oldFile.Syntax() != newFile.Syntax() {
		report(oldFile, "FILE_SAME_SYNTAX", fmt.Sprintf("File %q changed syntax from %s to %s.", name, oldFile.Syntax(), newFile.Syntax()))
	}
	oldProto := protoutil.ProtoFromFileDescriptor(oldFile)
	newProto := protoutil.ProtoFromFileDescriptor(newFile)
	if oldProto.GetEdition() != newProto.GetEdition() {
		report(oldFile, "FILE_SAME_SYNTAX", fmt.Sprintf("File %q changed edition from %s to %s.", name, oldProto.GetEdition(), newProto.GetEdition()))
	}
}

func compareOption[T comparable](descriptor protoreflect.Descriptor, oldValue, newValue T, rule, option string, report func(protoreflect.Descriptor, string, string)) {
	if oldValue != newValue {
		report(descriptor, rule, fmt.Sprintf("File option %s changed.", option))
	}
}

func compareFieldCodegenOptions(oldField, newField protoreflect.FieldDescriptor, report func(protoreflect.Descriptor, string, string)) {
	oldOptions := oldField.Options().(*descriptorpb.FieldOptions)
	newOptions := newField.Options().(*descriptorpb.FieldOptions)
	if oldOptions.GetJstype() != newOptions.GetJstype() {
		report(oldField, "FIELD_SAME_JSTYPE", fmt.Sprintf("Field %q changed jstype option.", oldField.FullName()))
	}
	if effectiveCPPString(oldField) != effectiveCPPString(newField) {
		report(oldField, "FIELD_SAME_CPP_STRING_TYPE", fmt.Sprintf("Field %q changed C++ string type option.", oldField.FullName()))
	}
	if oldField.Kind() == protoreflect.StringKind && newField.Kind() == protoreflect.StringKind && effectiveUTF8(oldField) != effectiveUTF8(newField) {
		report(oldField, "FIELD_SAME_UTF8_VALIDATION", fmt.Sprintf("Field %q changed UTF-8 validation behavior.", oldField.FullName()))
	}
	if oldField.Kind() == protoreflect.StringKind && newField.Kind() == protoreflect.StringKind && effectiveJavaUTF8(oldField) != effectiveJavaUTF8(newField) {
		report(oldField, "FIELD_SAME_JAVA_UTF8_VALIDATION", fmt.Sprintf("Field %q changed Java UTF-8 validation behavior.", oldField.FullName()))
	}
}

func effectiveUTF8(field protoreflect.FieldDescriptor) descriptorpb.FeatureSet_Utf8Validation {
	for current := protoreflect.Descriptor(field); current != nil; current = current.Parent() {
		features := descriptorFeatures(current)
		if features != nil && features.Utf8Validation != nil {
			return features.GetUtf8Validation()
		}
	}
	if field.Syntax() == protoreflect.Proto3 || field.Syntax() == protoreflect.Editions {
		return descriptorpb.FeatureSet_VERIFY
	}
	return descriptorpb.FeatureSet_NONE
}

func effectiveJavaUTF8(field protoreflect.FieldDescriptor) bool {
	if value, ok := languageFeature(field, "pb.java", "utf8_validation"); ok {
		if value == "VERIFY" {
			return true
		}
		if value == "NONE" {
			return false
		}
	}
	return effectiveUTF8(field) == descriptorpb.FeatureSet_VERIFY || field.ParentFile().Options().(*descriptorpb.FileOptions).GetJavaStringCheckUtf8()
}

func fieldFeatures(field protoreflect.FieldDescriptor) *descriptorpb.FeatureSet {
	return field.Options().(*descriptorpb.FieldOptions).GetFeatures()
}

func compareMessageOptions(oldMessage, newMessage protoreflect.MessageDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	if oldMessage.Options().(*descriptorpb.MessageOptions).GetMessageSetWireFormat() != newMessage.Options().(*descriptorpb.MessageOptions).GetMessageSetWireFormat() {
		report(oldMessage, "MESSAGE_SAME_MESSAGE_SET_WIRE_FORMAT", fmt.Sprintf("Message %q changed message-set encoding.", oldMessage.FullName()))
	}
	if !profile.source {
		return
	}
	oldOptions := oldMessage.Options().(*descriptorpb.MessageOptions)
	newOptions := newMessage.Options().(*descriptorpb.MessageOptions)
	if !oldOptions.GetNoStandardDescriptorAccessor() && newOptions.GetNoStandardDescriptorAccessor() {
		report(oldMessage, "MESSAGE_NO_REMOVE_STANDARD_DESCRIPTOR_ACCESSOR", fmt.Sprintf("Message %q disabled standard descriptor accessors.", oldMessage.FullName()))
	}
}

func compareOneofs(oldMessage, newMessage protoreflect.MessageDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	if !profile.source {
		return
	}
	for i := range oldMessage.Oneofs().Len() {
		oldOneof := oldMessage.Oneofs().Get(i)
		if oldOneof.IsSynthetic() {
			continue
		}
		newOneof := newMessage.Oneofs().ByName(oldOneof.Name())
		if newOneof == nil || newOneof.IsSynthetic() {
			report(oldOneof, "ONEOF_NO_DELETE", fmt.Sprintf("Oneof %q on %q was deleted.", oldOneof.Name(), oldMessage.FullName()))
			continue
		}
		if profile.file && oldOneof.ParentFile().Path() != newOneof.ParentFile().Path() {
			report(oldOneof, "ONEOF_MOVED", fmt.Sprintf("Oneof %q on %q was moved from %q to %q.", oldOneof.Name(), oldMessage.FullName(), oldOneof.ParentFile().Path(), newOneof.ParentFile().Path()))
		}
	}
}

func compareRequiredFields(oldMessage, newMessage protoreflect.MessageDescriptor, report func(protoreflect.Descriptor, string, string)) {
	oldRequired, newRequired := oldMessage.RequiredNumbers(), newMessage.RequiredNumbers()
	for i := range oldRequired.Len() {
		number := oldRequired.Get(i)
		if !newRequired.Has(number) {
			report(oldMessage, "MESSAGE_SAME_REQUIRED_FIELDS", fmt.Sprintf("Message %q removed required field number %d.", oldMessage.FullName(), number))
		}
	}
	for i := range newRequired.Len() {
		number := newRequired.Get(i)
		if !oldRequired.Has(number) {
			field := newMessage.Fields().ByNumber(number)
			report(field, "MESSAGE_SAME_REQUIRED_FIELDS", fmt.Sprintf("Message %q added required field number %d.", oldMessage.FullName(), number))
		}
	}
}

func compareMessageReserved(oldMessage, newMessage protoreflect.MessageDescriptor, report func(protoreflect.Descriptor, string, string)) {
	for i := range oldMessage.ReservedNames().Len() {
		name := oldMessage.ReservedNames().Get(i)
		if !newMessage.ReservedNames().Has(name) {
			report(oldMessage, "RESERVED_MESSAGE_NO_DELETE", fmt.Sprintf("Message %q released reserved field name %q.", oldMessage.FullName(), name))
		}
	}
	if !rangesPreserved(oldMessage.ReservedRanges(), newMessage.ReservedRanges()) {
		report(oldMessage, "RESERVED_MESSAGE_NO_DELETE", fmt.Sprintf("Message %q released a reserved field number range.", oldMessage.FullName()))
	}
}

func compareEnumReserved(oldEnum, newEnum protoreflect.EnumDescriptor, report func(protoreflect.Descriptor, string, string)) {
	for i := range oldEnum.ReservedNames().Len() {
		name := oldEnum.ReservedNames().Get(i)
		if !newEnum.ReservedNames().Has(name) {
			report(oldEnum, "RESERVED_ENUM_NO_DELETE", fmt.Sprintf("Enum %q released reserved value name %q.", oldEnum.FullName(), name))
		}
	}
	if !enumRangesPreserved(oldEnum.ReservedRanges(), newEnum.ReservedRanges()) {
		report(oldEnum, "RESERVED_ENUM_NO_DELETE", fmt.Sprintf("Enum %q released a reserved number range.", oldEnum.FullName()))
	}
}

func compareExtensionRanges(oldMessage, newMessage protoreflect.MessageDescriptor, profile breakingProfiles, report func(protoreflect.Descriptor, string, string)) {
	if !profile.source {
		return
	}
	if !rangesPreserved(oldMessage.ExtensionRanges(), newMessage.ExtensionRanges()) {
		report(oldMessage, "EXTENSION_MESSAGE_NO_DELETE", fmt.Sprintf("Message %q removed or narrowed an extension range.", oldMessage.FullName()))
	}
}

func rangesPreserved(oldRanges, newRanges protoreflect.FieldRanges) bool {
	oldIntervals := make([][2]int64, oldRanges.Len())
	newIntervals := make([][2]int64, newRanges.Len())
	for i := range oldIntervals {
		r := oldRanges.Get(i)
		oldIntervals[i] = [2]int64{int64(r[0]), int64(r[1])}
	}
	for i := range newIntervals {
		r := newRanges.Get(i)
		newIntervals[i] = [2]int64{int64(r[0]), int64(r[1])}
	}
	return intervalsPreserved(oldIntervals, newIntervals)
}

func enumRangesPreserved(oldRanges, newRanges protoreflect.EnumRanges) bool {
	oldIntervals := make([][2]int64, oldRanges.Len())
	newIntervals := make([][2]int64, newRanges.Len())
	for i := range oldIntervals {
		r := oldRanges.Get(i)
		oldIntervals[i] = [2]int64{int64(r[0]), int64(r[1]) + 1}
	}
	for i := range newIntervals {
		r := newRanges.Get(i)
		newIntervals[i] = [2]int64{int64(r[0]), int64(r[1]) + 1}
	}
	return intervalsPreserved(oldIntervals, newIntervals)
}

func jsonProfileEnabled(profile breakingProfiles) bool {
	return profile.source || profile.wireJSON
}

func sameJSONFormat(oldDescriptor, newDescriptor protoreflect.Descriptor) bool {
	return effectiveJSONFormat(oldDescriptor) != descriptorpb.FeatureSet_ALLOW || effectiveJSONFormat(newDescriptor) == descriptorpb.FeatureSet_ALLOW
}

func effectiveJSONFormat(descriptor protoreflect.Descriptor) descriptorpb.FeatureSet_JsonFormat {
	for current := descriptor; current != nil; current = current.Parent() {
		var features *descriptorpb.FeatureSet
		switch typed := current.(type) {
		case protoreflect.FileDescriptor:
			features = typed.Options().(*descriptorpb.FileOptions).GetFeatures()
		case protoreflect.MessageDescriptor:
			features = typed.Options().(*descriptorpb.MessageOptions).GetFeatures()
		case protoreflect.EnumDescriptor:
			features = typed.Options().(*descriptorpb.EnumOptions).GetFeatures()
		}
		if features != nil && features.JsonFormat != nil {
			return features.GetJsonFormat()
		}
		if file, ok := current.(protoreflect.FileDescriptor); ok {
			if file.Syntax() == protoreflect.Proto3 || file.Syntax() == protoreflect.Editions {
				return descriptorpb.FeatureSet_ALLOW
			}
			return descriptorpb.FeatureSet_LEGACY_BEST_EFFORT
		}
	}
	return descriptorpb.FeatureSet_JSON_FORMAT_UNKNOWN
}

func deduplicateBreakingIssues(issues []IssueInfo) []IssueInfo {
	seen := make(map[string]bool, len(issues))
	result := make([]IssueInfo, 0, len(issues))
	for _, issue := range issues {
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%d", issue.Path, issue.RuleName, issue.Message, issue.Position.Line, issue.Position.Column)
		if !seen[key] {
			seen[key] = true
			result = append(result, issue)
		}
	}
	slices.SortFunc(result, func(left, right IssueInfo) int {
		if order := strings.Compare(left.Path, right.Path); order != 0 {
			return order
		}
		if left.Position.Line != right.Position.Line {
			return left.Position.Line - right.Position.Line
		}
		if left.Position.Column != right.Position.Column {
			return left.Position.Column - right.Position.Column
		}
		if order := strings.Compare(left.RuleName, right.RuleName); order != 0 {
			return order
		}
		return strings.Compare(left.Message, right.Message)
	})
	return result
}

// intervalsPreserved checks interval coverage, not individual field numbers.
// int64 half-open endpoints also represent an enum's inclusive MaxInt32 safely.
func intervalsPreserved(oldRanges, newRanges [][2]int64) bool {
	slices.SortFunc(newRanges, func(a, b [2]int64) int {
		if a[0] < b[0] {
			return -1
		}
		if a[0] > b[0] {
			return 1
		}
		return 0
	})
	for _, old := range oldRanges {
		end := old[0]
		for _, next := range newRanges {
			if next[1] <= end {
				continue
			}
			if next[0] > end {
				break
			}
			end = next[1]
			if end >= old[1] {
				break
			}
		}
		if end < old[1] {
			return false
		}
	}
	return true
}

func sameFieldDefault(oldField, newField protoreflect.FieldDescriptor) bool {
	if !oldField.HasDefault() && !newField.HasDefault() {
		if oldField.Kind() == protoreflect.EnumKind && newField.Kind() == protoreflect.EnumKind {
			if oldField.Cardinality() == protoreflect.Repeated || newField.Cardinality() == protoreflect.Repeated {
				return true
			}
			// Linked descriptors expose DefaultEnumValue only for an explicit
			// option. The implicit enum default is its first declared value.
			return oldField.Enum().Values().Get(0).Number() == newField.Enum().Values().Get(0).Number()
		}
		return true
	}
	normalize := func(field protoreflect.FieldDescriptor) string {
		value := field.Default()
		switch field.Kind() {
		case protoreflect.BytesKind:
			return string(value.Bytes())
		case protoreflect.StringKind:
			return value.String()
		case protoreflect.BoolKind:
			if value.Bool() {
				return "1"
			}
			return "0"
		default:
			return fmt.Sprint(value.Interface())
		}
	}
	return normalize(oldField) == normalize(newField)
}

func descriptorFeatures(d protoreflect.Descriptor) *descriptorpb.FeatureSet {
	switch d := d.(type) {
	case protoreflect.FieldDescriptor:
		return fieldFeatures(d)
	case protoreflect.MessageDescriptor:
		return d.Options().(*descriptorpb.MessageOptions).GetFeatures()
	case protoreflect.EnumDescriptor:
		return d.Options().(*descriptorpb.EnumOptions).GetFeatures()
	case protoreflect.FileDescriptor:
		return d.Options().(*descriptorpb.FileOptions).GetFeatures()
	default:
		return nil
	}
}

func languageFeature(d protoreflect.Descriptor, extension, name string) (string, bool) {
	for current := d; current != nil; current = current.Parent() {
		features := descriptorFeatures(current)
		if features == nil {
			continue
		}
		var value string
		found := false
		features.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if string(field.FullName()) != extension || field.Kind() != protoreflect.MessageKind {
				return true
			}
			message := v.Message()
			member := message.Descriptor().Fields().ByName(protoreflect.Name(name))
			if member == nil || !message.Has(member) || member.Kind() != protoreflect.EnumKind {
				return true
			}
			enum := member.Enum().Values().ByNumber(message.Get(member).Enum())
			if enum != nil {
				value = string(enum.Name())
				found = true
			}
			return !found
		})
		if found {
			return value, true
		}
	}
	return "", false
}

func effectiveCPPString(field protoreflect.FieldDescriptor) string {
	if value, ok := languageFeature(field, "pb.cpp", "string_type"); ok {
		return value
	}
	options := field.Options().(*descriptorpb.FieldOptions)
	switch options.GetCtype() {
	case descriptorpb.FieldOptions_CORD:
		return "CORD"
	case descriptorpb.FieldOptions_STRING_PIECE:
		return "VIEW"
	default:
		return "STRING"
	}
}
