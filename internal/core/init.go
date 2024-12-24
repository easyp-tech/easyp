package core

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	wfs "github.com/easyp-tech/easyp/internal/fs"
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

// Initialize initializes the EasyP configuration.
func (i *Core) Initialize(ctx context.Context, disk wfs.DirWalker, defaultLinters []string) error {
	config := defaultConfig(defaultLinters)

	var migrated bool
	err := disk.WalkDir(func(path string, disk wfs.FS, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		}

		defaultConfiguration := defaultConfig(defaultLinters)
		if filepath.Base(path) == "buf.yml" || filepath.Base(path) == "buf.yaml" {
			migrated = true
			err = migrateFromBUF(disk, path, defaultConfiguration)
			if err != nil {
				return fmt.Errorf("migrateFromBUF: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("fs.WalkDir: %w", err)
	}

	if !migrated {
		filename := "easyp.yaml"
		res, err := disk.Create(filename)
		if err != nil {
			return fmt.Errorf("disk.Create: %w", err)
		}

		err = yaml.NewEncoder(res).Encode(config)
		if err != nil {
			return fmt.Errorf("yaml.NewEncoder.Encode: %w", err)
		}
	}

	return nil
}

func defaultConfig(defaultLinters []string) EasyPConfig {
	return EasyPConfig{
		Version: "v1alpha",
		Lint: Lint{
			Use:                                  defaultLinters,
			AllowCommentIgnores:                  false,
			EnumZeroValueSuffix:                  "_NONE",
			RPCAllowSameRequestResponse:          false,
			RPCAllowGoogleProtobufEmptyRequests:  false,
			RPCAllowGoogleProtobufEmptyResponses: false,
			ServiceSuffix:                        "API",
		},
	}
}

func migrateFromBUF(disk wfs.FS, path string, defaultConfiguration EasyPConfig) error {
	f, err := disk.Open(path)
	if err != nil {
		return fmt.Errorf("disk.Open: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			// TODO: Handle error
		}
	}()

	b := BUFConfig{}

	err = yaml.NewDecoder(f).Decode(&b)
	if err != nil {
		return fmt.Errorf("yaml.NewDecoder.Decode: %w", err)
	}

	config := buildCfgFromBUF(defaultConfiguration, b)

	dir := filepath.Dir(path)

	filename := filepath.Join(dir, "easyp.yaml")
	res, err := disk.Create(filename)
	if err != nil {
		return fmt.Errorf("disk.Create: %w", err)
	}

	err = yaml.NewEncoder(res).Encode(config)
	if err != nil {
		return fmt.Errorf("yaml.NewEncoder.Encode: %w", err)
	}

	return nil
}

func buildCfgFromBUF(cfg EasyPConfig, bufConfig BUFConfig) EasyPConfig {
	return EasyPConfig{
		Version: cfg.Version,
		Deps:    nil,
		Lint: Lint{
			Use:                                  bufConfig.Lint.Use,
			Except:                               bufConfig.Lint.Except,
			Ignore:                               bufConfig.Lint.Ignore,
			IgnoreOnly:                           bufConfig.Lint.IgnoreOnly,
			AllowCommentIgnores:                  bufConfig.Lint.AllowCommentIgnores,
			EnumZeroValueSuffix:                  bufConfig.Lint.EnumZeroValueSuffix,
			RPCAllowSameRequestResponse:          bufConfig.Lint.RPCAllowSameRequestResponse,
			RPCAllowGoogleProtobufEmptyRequests:  bufConfig.Lint.RPCAllowGoogleProtobufEmptyRequests,
			RPCAllowGoogleProtobufEmptyResponses: bufConfig.Lint.RPCAllowGoogleProtobufEmptyResponses,
			ServiceSuffix:                        bufConfig.Lint.ServiceSuffix,
		},
	}
}
