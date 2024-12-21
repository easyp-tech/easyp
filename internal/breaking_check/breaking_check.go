package breakingcheck

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/lint"
)

type BreakingCheck struct {
	against ProtoData
	current ProtoData
}

func (b *BreakingCheck) Check() ([]lint.IssueInfo, error) {
	res := make([]lint.IssueInfo, 0)

	// iterate over packages
	for packageName, collection := range b.against {
		issues := b.checkPackage(packageName, collection)

		res = append(res, issues...)
	}

	return res, nil
}

func (b *BreakingCheck) checkPackage(packageName PackageName, collection *Collection) []lint.IssueInfo {
	res := make([]lint.IssueInfo, 0)

	// iterate over services
	for serviceName, _ := range collection.Services {
		issues := b.checkService(packageName, serviceName)
		res = append(res, issues...)
	}

	for _, againstMessage := range collection.Messages {
		issues := b.checkMessage(againstMessage)
		res = append(res, issues...)
	}

	return res
}

// ===== SERVICE =====

func (b *BreakingCheck) checkService(packageName PackageName, serviceName ServiceName) []lint.IssueInfo {
	res := make([]lint.IssueInfo, 0)

	againstService := b.against[packageName].Services[serviceName]

	currentService, ok := b.current[packageName].Services[serviceName]
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

func (b *BreakingCheck) checkMessage(againstMessage Message) []lint.IssueInfo {
	res := make([]lint.IssueInfo, 0)

	currentMessage, ok := getMessage(b.current, againstMessage.PackageName, againstMessage.MessagePath)
	if !ok {
		// message was deleted
		issue := getMessageDeletedIssue(againstMessage)
		res = append(res, issue)
		return res
	}
	_ = currentMessage

	return res
}

// checkRootMessage check message from RPC (request or response) for breaking check
func (b *BreakingCheck) checkRootMessage(packageName PackageName, messageName MessageName) []lint.IssueInfo {
	res := make([]lint.IssueInfo, 0)

	//messageInfo := lint.InstructionParser{
	//	SourcePkgName: string(packageName),
	//}.Parse(string(messageName))
	//
	//againstMessage, _ := getMessage(b.against, PackageName(messageInfo.PkgName), MessageName(messageInfo.Instruction))
	//
	//currentMessage, ok := getMessage(b.current, PackageName(messageInfo.PkgName), MessageName(messageInfo.Instruction))
	//if !ok {
	//	// message was deleted
	//	issue := getMessageDeletedIssue(againstMessage)
	//	res = append(res, issue)
	//	return res
	//}
	//
	//issues := b.checkMessageOLD(againstMessage, currentMessage)
	//res = append(res, issues...)

	return res
}

func (b *BreakingCheck) checkMessageOLD(againstMessage, currentMessage Message) []lint.IssueInfo {
	res := make([]lint.IssueInfo, 0)

	// check fields
	//for _, againstField := range againstMessage.MessageBody.Fields {
	//	currentField, ok := searchField(currentMessage.MessageBody.Fields, againstField.FieldNumber)
	//	if !ok {
	//		issue := getFieldDeletedIssue(againstMessage, againstField)
	//		res = append(res, issue)
	//		continue
	//	}
	//
	//	// check fields type
	//	if againstField.Type != currentField.Type {
	//		issue := getFieldChangedTypeIssue(againstMessage, againstField, currentField)
	//		res = append(res, issue)
	//		continue
	//	}
	//
	//	// look for type in ProtoData, in nested
	//	// maybe should look for in deps as well
	//	messageInfo := lint.InstructionParser{
	//		SourcePkgName: string(againstMessage.PackageName),
	//	}.Parse(againstField.Type)
	//
	//	// look for in messages
	//	againstMessageType, ok := getMessage(
	//		b.against, PackageName(messageInfo.PkgName), MessageName(messageInfo.Instruction),
	//	)
	//	if ok {
	//		// message type found
	//		currentMessageType, ok := getMessage(
	//			b.current, PackageName(messageInfo.PkgName), MessageName(messageInfo.Instruction),
	//		)
	//		if !ok {
	//			issue := getMessageDeletedIssue(againstMessageType)
	//			res = append(res, issue)
	//			continue
	//		}
	//
	//		issues := b.checkMessageOLD(againstMessageType, currentMessageType)
	//		res = append(res, issues...)
	//		continue
	//	}
	//
	//} // for _, againstField := range againstMessage.MessageBody.Fields

	return res
}

func searchField(source []*parser.Field, number string) (*parser.Field, bool) {
	for _, field := range source {
		if field.FieldNumber == number {
			return field, true
		}
	}

	return nil, false
}

// ===== utils =====
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
