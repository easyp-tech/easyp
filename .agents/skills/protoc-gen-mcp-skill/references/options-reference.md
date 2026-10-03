# MCP Proto Options Reference

Complete reference for all `mcp.options.v1` protobuf extension options.

Import in your `.proto` files:
```proto
import "mcp/options/v1/options.proto";
```

## ServiceOptions

Applied via `option (mcp.options.v1.service) = { ... };` inside a `service` block.

| Field | Type | Description |
|---|---|---|
| `namespace` | `string` | Prefix for all generated tool names (e.g., `weather` → `weather_GetForecast`) |
| `description` | `string` | Overrides the service description inferred from proto comments |
| `icons` | `repeated Icon` | Default icon metadata for all tools in this service |

```proto
service WeatherAPI {
  option (mcp.options.v1.service) = {
    namespace: "weather"
    description: "Weather tools exposed as MCP tools."
    icons: [{
      src: "https://example.com/weather.png"
      mime_type: "image/png"
    }]
  };
}
```

## MethodOptions

Applied via `option (mcp.options.v1.method) = { ... };` inside an `rpc` block.

| Field | Type | Description |
|---|---|---|
| `name` | `string` | Override the RPC segment of the tool name |
| `title` | `string` | Human-readable tool title |
| `description` | `string` | Override description from proto comments |
| `hidden` | `bool` | Suppress tool generation for this RPC entirely |
| `annotations` | `ToolAnnotations` | Agent hints (see below) |
| `icons` | `repeated Icon` | Per-tool icons, overrides service default |
| `execution` | `ExecutionOptions` | Execution behavior (e.g., `task_support`) |

```proto
rpc Forecast(GetForecastRequest) returns (GetForecastResponse) {
  option (mcp.options.v1.method) = {
    name: "GetForecast"
    title: "Get forecast"
    description: "Fetch the forecast for a city."
    annotations: {
      read_only_hint: true
      idempotent_hint: true
    }
  };
}
```

### ToolAnnotations

| Field | Type | Description |
|---|---|---|
| `read_only_hint` | `bool` | Tool only reads data, no side effects |
| `destructive_hint` | `bool` | Tool may delete or irreversibly modify data |
| `idempotent_hint` | `bool` | Repeated calls with same input produce same result |
| `open_world_hint` | `bool` | Tool interacts with external systems |

### Icon

| Field | Type | Description |
|---|---|---|
| `src` | `string` | URI to the icon resource |
| `mime_type` | `string` | MIME type (e.g., `image/png`, `image/svg+xml`) |

### ExecutionOptions

| Field | Type | Description |
|---|---|---|
| `task_support` | `TaskSupport` | `TASK_SUPPORT_UNSPECIFIED` or `TASK_SUPPORT_OPTIONAL` |

## FieldOptions

Applied via `[(mcp.options.v1.field) = { ... }]` on a message field.

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Override description from proto comments |
| `examples` | `repeated ExampleValue` | Typed example values for the schema |
| `default_value` | `ExampleValue` | Explicit default value |
| `pattern` | `string` | Regex pattern for string fields |
| `format` | `string` | JSON Schema format (e.g., `email`, `date-time`, `uri`) |
| `min_length` | `uint32` | Minimum string length |
| `max_length` | `uint32` | Maximum string length |
| `minimum` | `float` | Minimum numeric value (inclusive) |
| `maximum` | `float` | Maximum numeric value (inclusive) |
| `exclusive_minimum` | `float` | Minimum numeric value (exclusive) |
| `exclusive_maximum` | `float` | Maximum numeric value (exclusive) |
| `multiple_of` | `float` | Number must be a multiple of this value |
| `min_items` | `uint32` | Minimum array length |
| `max_items` | `uint32` | Maximum array length |
| `unique_items` | `bool` | Array items must be unique |
| `read_only` | `bool` | Mark field as read-only |

```proto
string city = 1 [(mcp.options.v1.field) = {
  description: "City name."
  examples: [{ string_value: "Paris" }, { string_value: "London" }]
  min_length: 1
  max_length: 100
  pattern: "^[A-Z]"
}];

int32 count = 2 [(mcp.options.v1.field) = {
  default_value: { number_value: 10 }
  minimum: 1
  maximum: 1000
}];

repeated string labels = 5 [(mcp.options.v1.field) = {
  min_items: 1
  max_items: 50
  unique_items: true
}];
```

### ExampleValue

Typed example values used in `examples` and `default_value`:

| Oneof Field | Type | Usage |
|---|---|---|
| `string_value` | `string` | `{ string_value: "Paris" }` |
| `number_value` | `double` | `{ number_value: 42.5 }` |
| `integer_value` | `int64` | `{ integer_value: 10 }` |
| `bool_value` | `bool` | `{ bool_value: true }` |
| `array_value` | `ArrayValue` | `{ array_value: { items: [{ string_value: "a" }] } }` |
| `object_value` | `ObjectValue` | `{ object_value: { fields: [{ key: "k" value: { string_value: "v" } }] } }` |
| `null_value` | `NullValue` | `{ null_value: NULL_VALUE }` |

## MessageOptions

Applied via `option (mcp.options.v1.message) = { ... };` inside a `message` block.

| Field | Type | Description |
|---|---|---|
| `title` | `string` | Human-readable message title |
| `description` | `string` | Message description |
| `examples` | `repeated ExampleValue` | Example message payloads |

## OneofOptions

Applied via `option (mcp.options.v1.oneof) = { ... };` inside a `oneof` block.

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Description of the oneof group |
| `required` | `bool` | If true, exactly one variant must be set |

```proto
oneof selector {
  option (mcp.options.v1.oneof) = {
    description: "Select how to find the city."
    required: true
  };
  string city_alias = 41;
  int64 city_id = 42;
}
```

## EnumOptions

Applied via `option (mcp.options.v1.enum) = { ... };` inside an `enum` block.

| Field | Type | Description |
|---|---|---|
| `title` | `string` | Human-readable enum title |
| `description` | `string` | Enum description |

## EnumValueOptions

Applied via `[(mcp.options.v1.enum_value) = { ... }]` on an enum value.

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Description for this enum value |
| `hidden` | `bool` | Hide this value from the schema (commonly used for sentinel zero-values) |

```proto
enum ForecastMode {
  option (mcp.options.v1.enum) = {
    title: "Forecast Mode"
    description: "Scope of the forecast."
  };

  FORECAST_MODE_NONE = 0 [(mcp.options.v1.enum_value) = { hidden: true }];
  FORECAST_MODE_DAILY = 1;
  FORECAST_MODE_HOURLY = 2;
}
```

## Comment-Based Metadata

Proto comments also contribute to generated metadata:

- Plain comment lines become descriptions
- `Example: ...` adds a single schema example
- `Examples: ... | ...` adds multiple schema examples (pipe-separated)

Field options (`mcp.options.v1.field`) take precedence when both comments and
options define the same metadata.
