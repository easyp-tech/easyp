package core

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

const breakingCheckRuleName = "BREAKING_CHECK"

type BreakingChecker struct {
	against ProtoData
	current ProtoData
}

func (b *BreakingChecker) Check() ([]IssueInfo, error) {
	res := make([]IssueInfo, 0)

	// iterate over packages
	for packageName, collection := range b.against {
		issues := b.checkPackage(packageName, collection)

		res = append(res, issues...)
	}

	return res, nil
}

func (b *BreakingChecker) checkPackage(packageName PackageName, collection *Collection) []IssueInfo {
	res := make([]IssueInfo, 0)

	for _, againstImport := range collection.Imports {
		issues := b.checkImports(againstImport)
		res = append(res, issues...)
	}

	for _, againstService := range collection.Services {
		issues := b.checkService(againstService)
		res = append(res, issues...)
	}

	for _, againstMessage := range collection.Messages {
		issues := b.checkMessage(againstMessage)
		res = append(res, issues...)
	}

	for _, againstOneOf := range collection.OneOfs {
		issues := b.checkOneOf(againstOneOf)
		res = append(res, issues...)
	}

	for _, againstEnum := range collection.Enums {
		issues := b.checkEnum(againstEnum)
		res = append(res, issues...)
	}

	return res
}

// ===== IMPORTS =====

func (b *BreakingChecker) checkImports(againstImport Import) []IssueInfo {
	res := make([]IssueInfo, 0)

	_, ok := getImport(b.current, againstImport.PackageName, ImportPath(againstImport.Location))
	if !ok {
		issue := getImportDeletedIssue(againstImport)
		res = append(res, issue)
		return res
	}

	return res
}

// ===== SERVICE =====

func (b *BreakingChecker) checkService(againstService Service) []IssueInfo {
	res := make([]IssueInfo, 0)

	currentService, ok := getService(b.current, againstService.PackageName, againstService.ServiceName)
	if !ok {
		// service was deleted
		issue := getServiceDeletedIssue(againstService)
		res = append(res, issue)
		return res
	}

	// check RPCs
	for _, againstRPC := range againstService.ServiceBody.RPCs {
		// rpc was deleted
		currentRPC, ok := searchRPC(currentService.ServiceBody.RPCs, againstRPC.RPCName)
		if !ok {
			issue := getRPCDeletedIssue(againstService, againstRPC)
			res = append(res, issue)
			continue
		}

		// check RPCs

		// check request
		// TODO: check stream

		if againstRPC.RPCRequest.MessageType != currentRPC.RPCRequest.MessageType {
			issue := getRPCRequestChangedTypeIssue(againstService, currentService, againstRPC, currentRPC)
			res = append(res, issue)
		}

		// TODO: check stream
		// check response
		if againstRPC.RPCResponse.MessageType != currentRPC.RPCResponse.MessageType {
			issue := getRPCResponseChangedTypeIssue(againstService, currentService, againstRPC, currentRPC)
			res = append(res, issue)
		}

		// check messages in RPC
	}

	return res
}

func searchRPC(source []*parser.RPC, name string) (*parser.RPC, bool) {
	for _, rpc := range source {
		if rpc.RPCName == name {
			return rpc, true
		}
	}

	return nil, false
}

// ===== MESSAGE =====

func (b *BreakingChecker) checkMessage(againstMessage Message) []IssueInfo {
	res := make([]IssueInfo, 0)

	currentMessage, ok := getMessage(b.current, againstMessage.PackageName, againstMessage.MessagePath)
	if !ok {
		// message was deleted
		issue := getMessageDeletedIssue(againstMessage)
		res = append(res, issue)
		return res
	}

	// check fields
	for _, againstField := range againstMessage.MessageBody.Fields {
		currentField, ok := searchField(currentMessage.MessageBody.Fields, againstField.FieldNumber)
		if !ok {
			issue := getFieldDeletedIssue(againstMessage, againstField)
			res = append(res, issue)
			continue
		}

		if againstField.Type != currentField.Type {
			issue := getFieldChangedTypeIssue(againstMessage, againstField, currentField)
			res = append(res, issue)
			continue
		}

		if !againstField.IsOptional && currentField.IsOptional {
			issue := getFieldBecameOptional(againstMessage, againstField)
			res = append(res, issue)
		}
		if againstField.IsOptional && !currentField.IsOptional {
			issue := getFieldBecameNotOptional(againstMessage, againstField)
			res = append(res, issue)
		}
	}

	return res
}

// ===== OneOf =====

func (b *BreakingChecker) checkOneOf(againstOneOf OneOf) []IssueInfo {
	res := make([]IssueInfo, 0)

	currentOneOf, ok := getOneOf(b.current, againstOneOf.PackageName, againstOneOf.OneOfPath)
	if !ok {
		issue := getOneOfDeletedIssue(againstOneOf)
		res = append(res, issue)
		return res
	}

	// check fields
	for _, againstField := range againstOneOf.OneofFields {
		currentField, ok := searchOneOfField(currentOneOf.OneofFields, againstField.FieldNumber)
		if !ok {
			issue := getOneOfFieldDeletedIssue(againstOneOf, againstField)
			res = append(res, issue)
			continue
		}

		if againstField.Type != currentField.Type {
			issue := getOneOfFieldChangedTypeIssue(againstOneOf, againstField, currentField)
			res = append(res, issue)
			continue
		}
	}

	return res
}

// ===== ENUM =====

func (b *BreakingChecker) checkEnum(againstEnum Enum) []IssueInfo {
	res := make([]IssueInfo, 0)

	currentEnum, ok := getEnum(b.current, againstEnum.PackageName, againstEnum.EnumPath)
	if !ok {
		issue := getEnumDeletedIssue(againstEnum)
		res = append(res, issue)
		return res
	}

	for _, againstField := range againstEnum.EnumBody.EnumFields {
		currentField, ok := searchEnumField(currentEnum.EnumBody.EnumFields, againstField.Number)
		if !ok {
			issue := getEnumFieldDeletedIssue(againstEnum, againstField)
			res = append(res, issue)
			continue
		}

		if againstField.Ident != currentField.Ident {
			issue := getEnumFieldRenamedIssue(againstEnum, againstField, currentField)
			res = append(res, issue)
			continue
		}
	}

	return res
}

// ===== utils =====

func getImport(source ProtoData, packageName PackageName, importPath ImportPath) (Import, bool) {
	collection, ok := source[packageName]
	if !ok {
		return Import{}, false
	}

	imp, ok := collection.Imports[importPath]
	if !ok {
		return Import{}, false
	}

	return imp, true
}

func getService(source ProtoData, packageName PackageName, serviceName string) (Service, bool) {
	collection, ok := source[packageName]
	if !ok {
		return Service{}, false
	}

	service, ok := collection.Services[serviceName]
	if !ok {
		return Service{}, false
	}

	return service, true
}

func getMessage(source ProtoData, packageName PackageName, messagePath string) (Message, bool) {
	collection, ok := source[packageName]
	if !ok {
		return Message{}, false
	}

	message, ok := collection.Messages[messagePath]
	if !ok {
		return Message{}, false
	}

	return message, true
}

func getOneOf(source ProtoData, packageName PackageName, oneOfPath string) (OneOf, bool) {
	collection, ok := source[packageName]
	if !ok {
		return OneOf{}, false
	}

	oneOf, ok := collection.OneOfs[oneOfPath]
	if !ok {
		return OneOf{}, false
	}

	return oneOf, true
}

func getEnum(source ProtoData, packageName PackageName, enumPath string) (Enum, bool) {
	collection, ok := source[packageName]
	if !ok {
		return Enum{}, false
	}

	enum, ok := collection.Enums[enumPath]
	if !ok {
		return Enum{}, false
	}

	return enum, true
}

func searchField(source []*parser.Field, number string) (*parser.Field, bool) {
	for _, field := range source {
		if field.FieldNumber == number {
			return field, true
		}
	}

	return nil, false
}

func searchOneOfField(source []*parser.OneofField, number string) (*parser.OneofField, bool) {
	for _, field := range source {
		if field.FieldNumber == number {
			return field, true
		}
	}

	return nil, false
}

func searchEnumField(source []*parser.EnumField, number string) (*parser.EnumField, bool) {
	for _, field := range source {
		if field.Number == number {
			return field, true
		}
	}

	return nil, false
}

// issues

func buildBreakingCheckIssue(path, message string, pos meta.Position) IssueInfo {
	issue := Issue{
		Position:   pos,
		SourceName: "",
		Message:    message,
		RuleName:   breakingCheckRuleName,
	}
	return IssueInfo{
		Issue: issue,
		Path:  path,
	}
}

func getImportDeletedIssue(againstImport Import) IssueInfo {
	message := fmt.Sprintf("Previously import \"%s\" was deleted.\n", againstImport.Location)
	return buildBreakingCheckIssue(againstImport.ProtoFilePath, message, againstImport.Meta.Pos)
}

func getServiceDeletedIssue(againstService Service) IssueInfo {
	message := fmt.Sprintf(
		"Previously present service \"%s\" was deleted from file.", againstService.ServiceName,
	)
	return buildBreakingCheckIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos)
}

func getRPCDeletedIssue(againstService Service, againstRPC *parser.RPC) IssueInfo {
	message := fmt.Sprintf(
		"Previously present RPC \"%s\" on service \"%s\" was deleted.",
		againstRPC.RPCName, againstService.ServiceName,
	)
	return buildBreakingCheckIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos)
}

func getRPCRequestChangedTypeIssue(
	againstService, current Service, againstRPC, currentRPC *parser.RPC,
) IssueInfo {
	againstParser := InstructionParser{SourcePkgName: againstService.PackageName}
	currentParser := InstructionParser{SourcePkgName: current.PackageName}

	message := fmt.Sprintf(
		"RPC \"%s\" on service \"%s\" changed request type "+
			"from \"%s\" to \"%s\".",
		againstRPC.RPCName, againstService.ServiceName,
		againstParser.Parse(againstRPC.RPCRequest.MessageType).GetFullName(),
		currentParser.Parse(currentRPC.RPCRequest.MessageType).GetFullName(),
	)
	return buildBreakingCheckIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos)
}

func getRPCResponseChangedTypeIssue(
	againstService, current Service, againstRPC, currentRPC *parser.RPC,
) IssueInfo {
	againstParser := InstructionParser{SourcePkgName: againstService.PackageName}
	currentParser := InstructionParser{SourcePkgName: current.PackageName}

	message := fmt.Sprintf(
		"RPC \"%s\" on service \"%s\" changed response type "+
			"from \"%s\" to \"%s\".",
		againstRPC.RPCName, againstService.ServiceName,
		againstParser.Parse(againstRPC.RPCResponse.MessageType).GetFullName(),
		currentParser.Parse(currentRPC.RPCResponse.MessageType).GetFullName(),
	)
	return buildBreakingCheckIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos)
}

func getMessageDeletedIssue(againstMessage Message) IssueInfo {
	message := fmt.Sprintf(
		"Previously present message \"%s\" was deleted from file.\n", againstMessage.MessagePath,
	)
	return buildBreakingCheckIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos)
}

func getFieldDeletedIssue(againstMessage Message, againstField *parser.Field) IssueInfo {
	message := fmt.Sprintf("Previously present field \"%s\" with name \"%s\" "+
		"on message \"%s\" was deleted.",
		againstField.FieldNumber, againstField.FieldName, againstMessage.MessagePath,
	)
	return buildBreakingCheckIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos)
}

func getFieldChangedTypeIssue(againstMessage Message, againstField, currentField *parser.Field) IssueInfo {
	message := fmt.Sprintf("Field \"%s\" with name \"%s\" "+
		"on message \"%s\" changed type from \"%s\" to \"%s\".",
		againstField.FieldNumber, againstField.FieldName, againstMessage.MessageName,
		againstField.Type, currentField.Type,
	)
	return buildBreakingCheckIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos)
}

func getFieldBecameOptional(againstMessage Message, againstField *parser.Field) IssueInfo {
	message := fmt.Sprintf("Field \"%s\" with name \"%s\" "+
		"on message \"%s\" became optional",
		againstField.FieldNumber, againstField.FieldName, againstMessage.MessageName,
	)
	return buildBreakingCheckIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos)
}

func getFieldBecameNotOptional(againstMessage Message, againstField *parser.Field) IssueInfo {
	message := fmt.Sprintf("Field \"%s\" with name \"%s\" "+
		"on message \"%s\" became not optional",
		againstField.FieldNumber, againstField.FieldName, againstMessage.MessageName,
	)
	return buildBreakingCheckIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos)
}

func getOneOfDeletedIssue(againstOneOf OneOf) IssueInfo {
	message := fmt.Sprintf("Previously present oneof \"%s\" was deleted.",
		againstOneOf.OneOfPath,
	)
	return buildBreakingCheckIssue(againstOneOf.ProtoFilePath, message, againstOneOf.Meta.Pos)
}

func getOneOfFieldDeletedIssue(againstOneOf OneOf, againstField *parser.OneofField) IssueInfo {
	message := fmt.Sprintf("Previously present field \"%s\" with name \"%s\" "+
		"on OneOf \"%s\" was deleted.",
		againstField.FieldNumber, againstField.FieldName, againstOneOf.OneOfPath,
	)
	return buildBreakingCheckIssue(againstOneOf.ProtoFilePath, message, againstField.Meta.Pos)
}

func getOneOfFieldChangedTypeIssue(
	againstOneOf OneOf, againstOneOfField, currentOneOfField *parser.OneofField,
) IssueInfo {
	message := fmt.Sprintf("Field \"%s\" with name \"%s\" "+
		"on OneOf \"%s\" changed type from \"%s\" to \"%s\".",
		againstOneOfField.FieldNumber, againstOneOfField.FieldName, againstOneOf.OneOfPath,
		againstOneOfField.Type, currentOneOfField.Type,
	)
	return buildBreakingCheckIssue(againstOneOf.ProtoFilePath, message, againstOneOfField.Meta.Pos)
}

func getEnumDeletedIssue(againstEnum Enum) IssueInfo {
	message := fmt.Sprintf("Previously present enum \"%s\" was deleted from file.",
		againstEnum.EnumPath,
	)
	return buildBreakingCheckIssue(againstEnum.ProtoFilePath, message, againstEnum.Meta.Pos)
}

func getEnumFieldDeletedIssue(againstEnum Enum, againstField *parser.EnumField) IssueInfo {
	message := fmt.Sprintf("Previously present enum value \"%s\" on enum \"%s\" was deleted.",
		againstField.Number, againstEnum.EnumPath,
	)
	return buildBreakingCheckIssue(againstEnum.ProtoFilePath, message, againstEnum.Meta.Pos)
}

func getEnumFieldRenamedIssue(againstEnum Enum, againstField, currentField *parser.EnumField) IssueInfo {
	message := fmt.Sprintf("Enum value \"%s\" on enum \"%s\" changed "+
		"name from \"%s\" to \"%s\".",
		againstField.Number, againstEnum.EnumPath,
		againstField.Ident, currentField.Ident,
	)
	return buildBreakingCheckIssue(againstEnum.ProtoFilePath, message, againstEnum.Meta.Pos)
}
