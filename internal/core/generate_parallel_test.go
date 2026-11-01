package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
)

type parallelPluginExecutor struct {
	execute func(context.Context, pluginexecutor.Info, *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
}

func (e parallelPluginExecutor) Execute(ctx context.Context, info pluginexecutor.Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	return e.execute(ctx, info, request)
}

func (parallelPluginExecutor) GetName() string { return "parallelPluginExecutor" }

func TestPluginsOverlapAndApplyResponsesInConfigurationOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		started := make(chan string, 3)
		release := map[string]chan struct{}{
			"producer": make(chan struct{}),
			"first":    make(chan struct{}),
			"second":   make(chan struct{}),
		}
		responses := map[string]*pluginpb.CodeGeneratorResponse{
			"producer": {File: []*pluginpb.CodeGeneratorResponse_File{{
				Name: proto.String("out.txt"), Content: proto.String("base\n// @@protoc_insertion_point(items)\n"),
			}}},
			"first": {File: []*pluginpb.CodeGeneratorResponse_File{{
				Name: proto.String("out.txt"), InsertionPoint: proto.String("items"), Content: proto.String("first\n"),
			}}},
			"second": {File: []*pluginpb.CodeGeneratorResponse_File{{
				Name: proto.String("out.txt"), InsertionPoint: proto.String("items"), Content: proto.String("second\n"),
			}}},
		}
		executor := parallelPluginExecutor{execute: func(_ context.Context, info pluginexecutor.Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
			// Every executor may mutate its request, including nested descriptors.
			request.FileToGenerate[0] = info.Source
			request.ProtoFile[0].Name = proto.String(info.Source)
			request.Parameter = proto.String(info.Source)
			started <- info.Source
			<-release[info.Source]
			return responses[info.Source], nil
		}}
		app := testCoreWithPlugins([]Plugin{
			{Source: PluginSource{Name: "producer"}},
			{Source: PluginSource{Remote: "first"}},
			{Source: PluginSource{Name: "second", Command: []string{"custom-plugin"}}},
		}, executor)
		plan := &GenerationPlan{core: app, root: root, files: []string{"input.proto"},
			descriptors: []*descriptorpb.FileDescriptorProto{{Name: proto.String("input.proto")}},
		}
		bucket := NewGenerateBucket()
		done := make(chan error, 1)
		go func() { done <- plan.ExecuteInto(t.Context(), bucket) }()
		synctest.Wait()
		assert.Len(t, started, 3, "independent plugin calls must overlap")
		// Complete responses in reverse order and let each worker finish.
		close(release["second"])
		synctest.Wait()
		close(release["first"])
		synctest.Wait()
		close(release["producer"])
		require.NoError(t, <-done)
		output, ok := bucket.GetFile(t.Context(), filepath.Join(root, "out.txt"))
		require.True(t, ok)
		assert.Equal(t, "base\nfirst\nsecond\n// @@protoc_insertion_point(items)", string(output.Data()))
		assert.Equal(t, []string{"input.proto"}, plan.files)
		assert.Equal(t, "input.proto", plan.descriptors[0].GetName())
	})
}

func TestParallelPluginErrorsFollowConfigurationOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	firstErr := errors.New("first plugin failed")
	laterErr := errors.New("later plugin failed")
	synctest.Test(t, func(t *testing.T) {
		laterFinished := make(chan struct{})
		executor := parallelPluginExecutor{execute: func(_ context.Context, info pluginexecutor.Info, _ *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
			if info.Source == "first" {
				<-laterFinished
				return nil, firstErr
			}
			close(laterFinished)
			return nil, laterErr
		}}
		plan := &GenerationPlan{core: testCoreWithPlugins([]Plugin{
			{Source: PluginSource{Name: "first"}}, {Source: PluginSource{Name: "later"}},
		}, executor), root: root}
		require.ErrorIs(t, plan.Execute(t.Context()), firstErr)
	})
}

func TestEarlierInvalidOutputPrecedesLaterExecutionError(t *testing.T) {
	t.Parallel()
	executor := parallelPluginExecutor{execute: func(_ context.Context, info pluginexecutor.Info, _ *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
		if info.Source == "invalid-output" {
			return &pluginpb.CodeGeneratorResponse{File: []*pluginpb.CodeGeneratorResponse_File{{Name: proto.String("../escape")}}}, nil
		}
		return nil, errors.New("later execution error")
	}}
	plan := &GenerationPlan{core: testCoreWithPlugins([]Plugin{
		{Source: PluginSource{Name: "invalid-output"}}, {Source: PluginSource{Name: "later"}},
	}, executor), root: t.TempDir()}
	require.ErrorContains(t, plan.Execute(t.Context()), "invalid output path")
}

func TestParallelPluginFailureDoesNotWriteSuccessfulOutputs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		response *pluginpb.CodeGeneratorResponse
		message  string
	}{
		{name: "protocol error", response: &pluginpb.CodeGeneratorResponse{Error: proto.String("generation failed")}, message: "generation failed"},
		{name: "missing response", message: "returned no response"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			executor := parallelPluginExecutor{execute: func(_ context.Context, info pluginexecutor.Info, _ *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
				if info.Source == "failed" {
					return tt.response, nil
				}
				return &pluginpb.CodeGeneratorResponse{File: []*pluginpb.CodeGeneratorResponse_File{{
					Name: proto.String("out.txt"), Content: proto.String("successful output"),
				}}}, nil
			}}
			root := t.TempDir()
			plan := &GenerationPlan{core: testCoreWithPlugins([]Plugin{
				{Source: PluginSource{Name: "successful"}}, {Source: PluginSource{Name: "failed"}},
			}, executor), root: root}
			require.ErrorContains(t, plan.Execute(t.Context()), tt.message)
			_, err := os.Stat(filepath.Join(root, "out.txt"))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestParallelPluginsAreBoundedAndCancellationJoins(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		release := make(chan struct{})
		var started atomic.Int64
		executor := parallelPluginExecutor{execute: func(_ context.Context, _ pluginexecutor.Info, _ *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
			started.Add(1)
			<-release
			return &pluginpb.CodeGeneratorResponse{File: []*pluginpb.CodeGeneratorResponse_File{{Name: proto.String("out.txt"), Content: proto.String("data")}}}, nil
		}}
		plugins := make([]Plugin, 12)
		for index := range plugins {
			plugins[index].Source.Name = "plugin"
		}
		plan := &GenerationPlan{core: testCoreWithPlugins(plugins, executor), root: root}
		done := make(chan error, 1)
		go func() { done <- plan.Execute(ctx) }()
		synctest.Wait()
		assert.EqualValues(t, 4, started.Load())
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Error("returned before active plugin calls finished")
		default:
		}
		close(release)
		require.ErrorIs(t, <-done, context.Canceled)
		assert.EqualValues(t, 4, started.Load())
		_, err := os.Stat(filepath.Join(root, "out.txt"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
