package breakingcheck

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
)

// buildError creates an Issue.
// TODO: almost the same as lint.buildError
func buildIssue(path, message string, pos meta.Position, sourceName string) lint.IssueInfo {
	issue := lint.Issue{
		Position: pos,
		//SourceName: sourceName,
		SourceName: "",
		Message:    message,
		RuleName:   "BREAKING_CHECK",
	}
	return lint.IssueInfo{
		Issue: issue,
		Path:  path,
	}
}

func getServiceDeletedIssue(againstService Service) lint.IssueInfo {
	message := fmt.Sprintf(
		"Previously present service \"%s\" was deleted from file.", againstService.ServiceName,
	)
	return buildIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos, againstService.ServiceName)
}

func getRPCDeletedIssue(againstService Service, againstRPC *parser.RPC) lint.IssueInfo {
	message := fmt.Sprintf(
		"Previously present RPC \"%s\" on service \"%s\" was deleted.",
		againstRPC.RPCName, againstService.ServiceName,
	)
	return buildIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos, againstService.ServiceName)
}

func getRPCRequestChangedTypeIssue(
	againstService, current Service, againstRPC, currentRPC *parser.RPC,
) lint.IssueInfo {
	againstParser := lint.InstructionParser{SourcePkgName: string(againstService.PackageName)}
	currentParser := lint.InstructionParser{SourcePkgName: string(current.PackageName)}

	message := fmt.Sprintf(
		"RPC \"%s\" on service \"%s\" changed request type "+
			"from \"%s\" to \"%s\".",
		againstRPC.RPCName, againstService.ServiceName,
		againstParser.Parse(againstRPC.RPCRequest.MessageType).GetFullName(),
		currentParser.Parse(currentRPC.RPCRequest.MessageType).GetFullName(),
	)
	return buildIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos, againstService.ServiceName)
}

func getRPCResponseChangedTypeIssue(
	againstService, current Service, againstRPC, currentRPC *parser.RPC,
) lint.IssueInfo {
	againstParser := lint.InstructionParser{SourcePkgName: string(againstService.PackageName)}
	currentParser := lint.InstructionParser{SourcePkgName: string(current.PackageName)}

	message := fmt.Sprintf(
		"RPC \"%s\" on service \"%s\" changed response type "+
			"from \"%s\" to \"%s\".",
		againstRPC.RPCName, againstService.ServiceName,
		againstParser.Parse(againstRPC.RPCResponse.MessageType).GetFullName(),
		currentParser.Parse(currentRPC.RPCResponse.MessageType).GetFullName(),
	)
	return buildIssue(againstService.ProtoFilePath, message, againstService.Meta.Pos, againstService.ServiceName)
}

func getMessageDeletedIssue(againstMessage Message) lint.IssueInfo {
	message := fmt.Sprintf(
		"Previously present message \"%s\" was deleted from file.\n", againstMessage.MessagePath,
	)
	return buildIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos, againstMessage.MessageName)
}

func getFieldDeletedIssue(againstMessage Message, againstField *parser.Field) lint.IssueInfo {
	message := fmt.Sprintf("Previously present field \"%s\" with name \"%s\" "+
		"on message \"%s\" was deleted.",
		againstField.FieldNumber, againstField.FieldName, againstMessage.MessagePath,
	)
	return buildIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos, againstMessage.MessageName)
}

func getFieldChangedTypeIssue(againstMessage Message, againstField, currentField *parser.Field) lint.IssueInfo {
	message := fmt.Sprintf("Field \"%s\" with name \"%s\" "+
		"on message \"%s\" changed type from \"%s\" to \"%s\".",
		againstField.FieldName, againstField.FieldName, againstMessage.MessageName,
		againstField.Type, currentField.Type,
	)
	return buildIssue(againstMessage.ProtoFilePath, message, againstMessage.Meta.Pos, againstMessage.MessageName)
}
