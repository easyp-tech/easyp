package breakingcheck

import (
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
		Services    map[ServiceName]Service
		MessagesOLD map[MessageName]Message
	}

	// collects proto data collections
	// packageName -> services,messages etc
	ProtoData map[PackageName]*Collection
)

func Collect(protoInfos []lint.ProtoInfo) (ProtoData, error) {
	protoData := make(ProtoData)

	for _, protoInfo := range protoInfos {
		pkgName := PackageName(protoInfo.GetPackageName())

		collection, ok := protoData[pkgName]
		if !ok {
			collection = newCollection()
		}

		protoFile := protoInfo.Info

		// read all services from protoFile
		for _, service := range protoFile.ProtoBody.Services {
			collection.Services[ServiceName(service.ServiceName)] = Service{
				ProtoFilePath: protoInfo.Path,
				PackageName:   pkgName,
				Service:       service,
			}
		}

		// read all messages
		for _, message := range protoFile.ProtoBody.Messages {
			collection.MessagesOLD[MessageName(message.MessageName)] = Message{
				ProtoFilePath: protoInfo.Path,
				PackageName:   pkgName,
				Message:       message,
			}
		}

		protoData[PackageName(protoInfo.GetPackageName())] = collection
	}

	return protoData, nil
}

func newCollection() *Collection {
	collection := &Collection{
		Services:    make(map[ServiceName]Service),
		MessagesOLD: make(map[MessageName]Message),
	}
	return collection
}
