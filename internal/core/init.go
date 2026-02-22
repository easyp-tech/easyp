package core

import (
	"bytes"
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
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
)

// InitOptions contains options for the Initialize command.
type InitOptions struct {
	TemplateData InitTemplateData
	Prompter     Prompter
}

// Prompter provides an interface for interactive user interaction.
type Prompter interface {
	// Confirm asks a yes/no question. Returns true if the user answered "yes".
	Confirm(ctx context.Context, message string, defaultValue bool) (bool, error)
}

// Initialize initializes the EasyP configuration.
func (c *Core) Initialize(ctx context.Context, disk DirWalker, opts InitOptions) (err error) {
	cfg := configFromTemplateData(opts.TemplateData)

	// Check for buf config only in root directory (not recursive).
	var migrated bool
	for _, bufConfigName := range []string{"buf.yml", "buf.yaml"} {
		if !disk.Exists(bufConfigName) {
			continue
		}

		migrate, promptErr := opts.Prompter.Confirm(ctx, "Found "+bufConfigName+". Migrate configuration?", true)
		if promptErr != nil {
			return fmt.Errorf("prompter.Confirm: %w", promptErr)
		}
		if !migrate {
			continue
		}

		migrated = true
		if migrateErr := c.migrateFromBUF(ctx, disk, bufConfigName, cfg); migrateErr != nil {
			return fmt.Errorf("migrateFromBUF: %w", migrateErr)
		}

		break
	}

	if !migrated {
		filename := "easyp.yaml"

		// Check for existing config (overwrite protection).
		if disk.Exists(filename) {
			overwrite, promptErr := opts.Prompter.Confirm(ctx, filename+" already exists. Overwrite?", false)
			if promptErr != nil {
				return fmt.Errorf("prompter.Confirm: %w", promptErr)
			}
			if !overwrite {
				return nil
			}
		}

		res, createErr := disk.Create(filename)
		if createErr != nil {
			return fmt.Errorf("disk.Create: %w", createErr)
		}
		defer func() {
			if closeErr := res.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("res.Close: %w", closeErr)
			}
		}()

		// Render to buffer and validate before writing to disk.
		var buf bytes.Buffer
		if renderErr := renderInitConfig(&buf, opts.TemplateData); renderErr != nil {
			return fmt.Errorf("renderInitConfig: %w", renderErr)
		}

		if issues, valErr := config.ValidateRaw(buf.Bytes()); valErr != nil {
			return fmt.Errorf("config.ValidateRaw: %w", valErr)
		} else if config.HasErrors(issues) {
			return fmt.Errorf("generated config has validation errors: %v", issues)
		}

		if _, writeErr := res.Write(buf.Bytes()); writeErr != nil {
			return fmt.Errorf("res.Write: %w", writeErr)
		}
	}

	return nil
}

// configFromTemplateData creates a config.Config from template data (used for buf migration).
func configFromTemplateData(data InitTemplateData) config.Config {
	var allRules []string
	for _, g := range data.LintGroups {
		allRules = append(allRules, g.Rules...)
	}

	return config.Config{
		Lint: config.LintConfig{
			Use:                 allRules,
			AllowCommentIgnores: false,
			EnumZeroValueSuffix: data.EnumZeroValueSuffix,
			ServiceSuffix:       data.ServiceSuffix,
		},
		BreakingCheck: config.BreakingCheck{
			AgainstGitRef: data.AgainstGitRef,
		},
	}
}

func (c *Core) migrateFromBUF(ctx context.Context, disk FS, path string, defaultConfiguration config.Config) (err error) {
	f, err := disk.Open(path)
	if err != nil {
		return fmt.Errorf("disk.Open: %w", err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = fmt.Errorf("disk.Close: %w: original error: %w", closeErr, err)
		}
	}()

	b := BUFConfig{}

	err = yaml.NewDecoder(f).Decode(&b)
	if err != nil {
		return fmt.Errorf("yaml.NewDecoder.Decode: %w", err)
	}

	// Log unsupported fields.
	if len(b.Build.Excludes) > 0 {
		c.logger.Warn(ctx, "buf build.excludes is not supported in easyp, skipping")
	}
	if len(b.Breaking.Use) > 0 {
		c.logger.Warn(ctx, "buf breaking.use granular rules not yet supported, using default breaking check")
	}

	migratedCfg := buildCfgFromBUF(defaultConfiguration, b)

	res, err := disk.Create("easyp.yaml")
	if err != nil {
		return fmt.Errorf("disk.Create: %w", err)
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("res.Close: %w", closeErr)
		}
	}()

	// Encode to buffer and validate before writing to disk.
	var buf bytes.Buffer
	err = yaml.NewEncoder(&buf).Encode(migratedCfg)
	if err != nil {
		return fmt.Errorf("yaml.NewEncoder.Encode: %w", err)
	}

	if issues, valErr := config.ValidateRaw(buf.Bytes()); valErr != nil {
		return fmt.Errorf("config.ValidateRaw: %w", valErr)
	} else if config.HasErrors(issues) {
		return fmt.Errorf("migrated config has validation errors: %v", issues)
	}

	if _, writeErr := res.Write(buf.Bytes()); writeErr != nil {
		return fmt.Errorf("res.Write: %w", writeErr)
	}

	return nil
}

func buildCfgFromBUF(cfg config.Config, bufConfig BUFConfig) config.Config {
	result := config.Config{
		Deps: bufConfig.Deps,
		Lint: config.LintConfig{
			Use:                 bufConfig.Lint.Use,
			Except:              bufConfig.Lint.Except,
			Ignore:              bufConfig.Lint.Ignore,
			IgnoreOnly:          bufConfig.Lint.IgnoreOnly,
			AllowCommentIgnores: bufConfig.Lint.AllowCommentIgnores,
			EnumZeroValueSuffix: bufConfig.Lint.EnumZeroValueSuffix,
			ServiceSuffix:       bufConfig.Lint.ServiceSuffix,
		},
		BreakingCheck: config.BreakingCheck{
			AgainstGitRef: cfg.BreakingCheck.AgainstGitRef,
			Ignore:        bufConfig.Breaking.Ignore,
		},
	}

	// Use defaults from the base config when buf config has empty values.
	if result.Lint.EnumZeroValueSuffix == "" {
		result.Lint.EnumZeroValueSuffix = cfg.Lint.EnumZeroValueSuffix
	}
	if result.Lint.ServiceSuffix == "" {
		result.Lint.ServiceSuffix = cfg.Lint.ServiceSuffix
	}

	return result
}
