package easypconfig

import invjsonschema "github.com/invopop/jsonschema"

type configSchemaRoot struct {
	Version  string                `json:"version,omitempty"`
	Lint     *configSchemaLint     `json:"lint,omitempty"`
	Deps     []string              `json:"deps,omitempty"`
	Generate *configSchemaGenerate `json:"generate,omitempty"`
	Breaking *configSchemaBreaking `json:"breaking,omitempty"`
}

type configSchemaLint struct {
	Use                 []string            `json:"use,omitempty"`
	EnumZeroValueSuffix string              `json:"enum_zero_value_suffix,omitempty"`
	ServiceSuffix       string              `json:"service_suffix,omitempty"`
	Ignore              []string            `json:"ignore,omitempty"`
	Except              []string            `json:"except,omitempty"`
	AllowCommentIgnores bool                `json:"allow_comment_ignores,omitempty"`
	IgnoreOnly          map[string][]string `json:"ignore_only,omitempty"`
}

type configSchemaGenerate struct {
	Inputs  []configSchemaInput  `json:"inputs"`
	Plugins []configSchemaPlugin `json:"plugins"`
	Managed *configSchemaManaged `json:"managed,omitempty"`
}

func (configSchemaGenerate) JSONSchemaExtend(schema *invjsonschema.Schema) {
	setMinItems(schema, "inputs", 1)
	setMinItems(schema, "plugins", 1)
}

type configSchemaInput struct {
	Directory configSchemaInputDirectory `json:"directory,omitempty"`
	GitRepo   *configSchemaInputGitRepo  `json:"git_repo,omitempty"`
}

func (configSchemaInput) JSONSchemaExtend(schema *invjsonschema.Schema) {
	schema.OneOf = []*invjsonschema.Schema{
		{Required: []string{"directory"}},
		{Required: []string{"git_repo"}},
	}
}

type configSchemaInputDirectory struct{}

func (configSchemaInputDirectory) JSONSchema() *invjsonschema.Schema {
	reflector := &invjsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	objectSchema := reflector.Reflect(configSchemaInputDirectoryObject{})
	objectSchema.Version = ""
	objectSchema.ID = ""
	objectSchema.Definitions = nil
	objectSchema.Title = ""

	return &invjsonschema.Schema{
		OneOf: []*invjsonschema.Schema{
			{Type: "string"},
			objectSchema,
		},
	}
}

type configSchemaInputDirectoryObject struct {
	Path string `json:"path"`
	Root string `json:"root,omitempty"`
}

type configSchemaInputGitRepo struct {
	URL          string `json:"url"`
	SubDirectory string `json:"sub_directory,omitempty"`
	Root         string `json:"root,omitempty"`
}

type configSchemaPlugin struct {
	Name        string                 `json:"name,omitempty"`
	Remote      string                 `json:"remote,omitempty"`
	Path        string                 `json:"path,omitempty"`
	Command     []string               `json:"command,omitempty"`
	Out         string                 `json:"out"`
	Opts        configSchemaPluginOpts `json:"opts,omitempty"`
	WithImports bool                   `json:"with_imports,omitempty"`
}

func (configSchemaPlugin) JSONSchemaExtend(schema *invjsonschema.Schema) {
	schema.OneOf = []*invjsonschema.Schema{
		{Required: []string{"name"}},
		{Required: []string{"remote"}},
		{Required: []string{"path"}},
		{Required: []string{"command"}},
	}
}

type configSchemaPluginOpts map[string]any

func (configSchemaPluginOpts) JSONSchema() *invjsonschema.Schema {
	return &invjsonschema.Schema{
		Type: "object",
		AdditionalProperties: &invjsonschema.Schema{
			OneOf: []*invjsonschema.Schema{
				{Type: "string"},
				{
					Type:  "array",
					Items: &invjsonschema.Schema{Type: "string"},
				},
			},
		},
	}
}

type configSchemaManaged struct {
	Enabled  bool                              `json:"enabled,omitempty"`
	Disable  []configSchemaManagedDisableRule  `json:"disable,omitempty"`
	Override []configSchemaManagedOverrideRule `json:"override,omitempty"`
}

type configSchemaManagedDisableRule struct {
	Module      string `json:"module,omitempty"`
	Path        string `json:"path,omitempty"`
	FileOption  string `json:"file_option,omitempty"`
	FieldOption string `json:"field_option,omitempty"`
	Field       string `json:"field,omitempty"`
}

func (configSchemaManagedDisableRule) JSONSchemaExtend(schema *invjsonschema.Schema) {
	schema.AnyOf = []*invjsonschema.Schema{
		{Required: []string{"module"}},
		{Required: []string{"path"}},
		{Required: []string{"file_option"}},
		{Required: []string{"field_option"}},
		{Required: []string{"field"}},
	}
	schema.Not = &invjsonschema.Schema{Required: []string{"file_option", "field_option"}}
	schema.DependentRequired = map[string][]string{
		"field": {"field_option"},
	}
}

type configSchemaManagedOverrideRule struct {
	FileOption  string `json:"file_option,omitempty"`
	FieldOption string `json:"field_option,omitempty"`
	Value       any    `json:"value"`
	Module      string `json:"module,omitempty"`
	Path        string `json:"path,omitempty"`
	Field       string `json:"field,omitempty"`
}

func (configSchemaManagedOverrideRule) JSONSchemaExtend(schema *invjsonschema.Schema) {
	schema.AnyOf = []*invjsonschema.Schema{
		{Required: []string{"file_option"}},
		{Required: []string{"field_option"}},
	}
	schema.Not = &invjsonschema.Schema{Required: []string{"file_option", "field_option"}}
	schema.DependentRequired = map[string][]string{
		"field": {"field_option"},
	}
}

type configSchemaBreaking struct {
	Ignore        []string `json:"ignore,omitempty"`
	AgainstGitRef string   `json:"against_git_ref,omitempty"`
}

func setMinItems(schema *invjsonschema.Schema, fieldName string, min uint64) {
	if schema == nil || schema.Properties == nil {
		return
	}

	itemSchema, ok := schema.Properties.Get(fieldName)
	if !ok || itemSchema == nil {
		return
	}

	itemSchema.MinItems = &min
}
