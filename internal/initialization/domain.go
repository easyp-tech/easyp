package initialization

import (
	"io/fs"
	"os"
)

type (
	// BUFConfig is the configuration for the buf tool.
	BUFConfig struct {
		Version  string   `yaml:"version"`
		Deps     []string `yaml:"deps"`
		Build    Build    `yaml:"build"`
		Lint     Lint     `yaml:"lint"`
		Breaking Breaking `yaml:"breaking"`
	}

	// Build is the configuration for the build section of the buf tool.
	Build struct {
		Excludes []string `yaml:"excludes"`
	}

	// Lint is the configuration for the lint section of the buf tool.
	Lint struct {
		Use                                  []string            `yaml:"use"`
		Except                               []string            `yaml:"except"`
		Ignore                               []string            `yaml:"ignore"`
		IgnoreOnly                           map[string][]string `yaml:"ignore_only"`
		AllowCommentIgnores                  bool                `yaml:"allow_comment_ignores"`
		EnumZeroValueSuffix                  string              `yaml:"enum_zero_value_suffix"`
		RPCAllowSameRequestResponse          bool                `yaml:"rpc_allow_same_request_response"`
		RPCAllowGoogleProtobufEmptyRequests  bool                `yaml:"rpc_allow_google_protobuf_empty_requests"`
		RPCAllowGoogleProtobufEmptyResponses bool                `yaml:"rpc_allow_google_protobuf_empty_responses"`
		ServiceSuffix                        string              `yaml:"service_suffix"`
	}

	// Breaking is the configuration for the breaking section of the buf tool.
	Breaking struct {
		Use                    []string            `yaml:"use"`
		Except                 []string            `yaml:"except"`
		Ignore                 []string            `yaml:"ignore"`
		IgnoreOnly             map[string][]string `yaml:"ignore_only"`
		IgnoreUnstablePackages bool                `yaml:"ignore_unstable_packages"`
	}

	// Migrate contains original configuration for the migration.
	Migrate struct {
		BUF *BUFConfig
		//Protoool *ProtooolConfig TODO
	}

	// EasyPConfig is the configuration for EasyP.
	EasyPConfig struct {
		Version  string   `yaml:"version"`
		Deps     []string `yaml:"deps"`
		Build    Build    `yaml:"build"`
		Lint     Lint     `yaml:"lint"`
		Breaking Breaking `yaml:"breaking"`
	}

	// FS is the interface for the file system.
	FS interface {
		fs.FS
		// Create creates a file.
		Create(name string) (*os.File, error)
	}
)
