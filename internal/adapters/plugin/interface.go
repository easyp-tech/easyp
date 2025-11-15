package plugin

import (
	"context"

	"google.golang.org/protobuf/types/pluginpb"
)

type Executor interface {
	Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
	GetName() string
}
