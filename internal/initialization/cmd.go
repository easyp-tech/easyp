package initialization

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"gopkg.in/yaml.v3"

	configPkg "github.com/easyp-tech/easyp/internal/config"
)

// Initialize initializes the EasyP configuration.
func (i *Init) Initialize(ctx context.Context, disk FS, defaultLinters []string) error {
	config := defaultConfig(defaultLinters)

	var migrated bool
	err := fs.WalkDir(disk, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			return nil
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
		filename := configPkg.DefaultConfigFileName
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

func migrateFromBUF(disk FS, path string, defaultConfiguration EasyPConfig) error {
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

	filename := filepath.Join(dir, configPkg.DefaultConfigFileName)
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
