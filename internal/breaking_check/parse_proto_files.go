package breakingcheck

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/lint"
)

// parsing protofiles
// read services, messages etc

// domain types
type (
	// alias for package name `package` section in protofile.
	PackageName string // TODO: move to domain layer. Use it in import_used linter

	ServiceName string
	MessageName string
)

// collections
type (
	Service struct {
		ProtoFilePath string
		PackageName   PackageName
		*unordered.Service
	}

	Message struct {
		ProtoFilePath string
		PackageName   PackageName
		*unordered.Message
	}

	// TODO: think about struct's name
	Collection struct {
		Services map[ServiceName]Service
		// key message path - for supporting nested messages:
		// message MainMessage {
		// 		message NestedMessage{};
		// };
		// will be: MainMessage.NestedMessage
		Messages    map[string]Message
		MessagesOLD map[MessageName]Message
	}

	// collects proto data collections
	// packageName -> services,messages etc
	ProtoData map[PackageName]*Collection
)

func Collect(protoInfos []lint.ProtoInfo) (ProtoData, error) {
	protoData := make(ProtoData)
	collectedProtoFiles := make(map[string]struct{})

	for _, protoInfo := range protoInfos {
		protoFilePath := protoInfo.Path
		pkgName := PackageName(lint.GetPackageName(protoInfo.Info))

		if _, ok := collectedProtoFiles[protoFilePath]; !ok {
			collectProtoFileInfo(protoData, protoInfo.Info, pkgName, protoFilePath)
			collectedProtoFiles[protoFilePath] = struct{}{}
		}

		// collectes from imports
		for importPath, protoFile := range protoInfo.ProtoFilesFromImport {
			protoFilePath := string(importPath)
			if _, ok := collectedProtoFiles[protoFilePath]; ok {
				continue
			}

			pkgName := lint.GetPackageName(protoFile)
			collectProtoFileInfo(protoData, protoFile, PackageName(pkgName), protoFilePath)
			collectedProtoFiles[protoFilePath] = struct{}{}
		}
	}

	return protoData, nil
}

func collectProtoFileInfo(protoData ProtoData, protoFile *unordered.Proto, pkgName PackageName, protoFilePath string) {
	collection, ok := protoData[pkgName]
	if !ok {
		collection = newCollection()
	}

	// read all services from protoFile
	for _, service := range protoFile.ProtoBody.Services {
		collection.Services[ServiceName(service.ServiceName)] = Service{
			ProtoFilePath: protoFilePath,
			PackageName:   pkgName,
			Service:       service,
		}
	}

	readMessages(collection, "", protoFile.ProtoBody.Messages, protoFilePath, pkgName)
	protoData[pkgName] = collection
}

func readMessages(
	collection *Collection,
	messagePath string,
	messages []*unordered.Message,
	protoFilePath string,
	packageName PackageName,
) {
	for _, message := range messages {
		msg := Message{
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Message:       message,
		}
		newMessagePath := getProtoEntityPath(messagePath, message.MessageName)
		if _, ok := collection.Messages[newMessagePath]; ok {
			panic("ALREADY EXIST") // TODO: return error - check for duplicate
		}
		collection.Messages[newMessagePath] = msg

		readMessages(collection, newMessagePath, message.MessageBody.Messages, protoFilePath, packageName)
	}
}

func getProtoEntityPath(rootPath, name string) string {
	if rootPath == "" {
		return name
	}

	return fmt.Sprintf("%s.%s", rootPath, name)
}

func newCollection() *Collection {
	collection := &Collection{
		Services:    make(map[ServiceName]Service),
		Messages:    make(map[string]Message),
		MessagesOLD: make(map[MessageName]Message),
	}
	return collection
}
