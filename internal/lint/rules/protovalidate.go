package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ProtoValidate)(nil)

// ProtoValidate this rule requires that all protovalidate constraints specified are valid.
type ProtoValidate struct{}

//This rule requires that all protovalidate constraints specified are valid.
//
//For a buf.validate.field to be valid, it must ensure:
//
//skipped is the only field if set.
//At most one of required and ignore_empty are set.
//required isn't set if the field belongs to a oneof.
//Neither required nor ignore_empty is set if the field is an extension.
//Its CEL constraints are valid.
//Its type specific rules, such as (buf.validate.field).int32, are valid.
//For a buf.validate.message to be valid, it must ensure:
//
//disabled is the only field if set.
//Its CEL constraints are valid.
//For a set of CEL constraints on a message or field to be valid, each constraint must:
//
//Have a CEL expression that compiles successfully and evaluates to a string or boolean. These are the only two types that the protovalidate runtime allows, and it's a runtime error for a CEL expression to evaluate to another type.
//Have a non-empty message if the CEL expression evaluates to a boolean value. This message is used by the protovalidate runtime report validation failure.
//Have an empty message if the CEL expression evaluates to a string value. The validation failure message in this case is the value this CEL expression evaluates to, while message won't be used in any way.
//Have a non-empty id, consisting of only alphanumeric characters, _, - and .. The id must be unique within the buf.validate.message or buf.validate.field it's specified on. A unique id is useful for debugging and locating the CEL constraint that fails, and can be used as a key for i18n.
//For a set of rules specified on a field, such as (buf.validate.field).int32, to be valid, it must additionally:
//
//Have a type compatible with the type it validates: (buf.validate.field).int32 rules can only be set on a field of type int32 or google.protobuf.Int32Value. A type mismatch causes a runtime error.
//Permit some value: setting contains: "foo" and not_contains: "foo" isn't valid because it rejects all values.
//Have no obviously redundant rules. For example, it's redundant to set lt: 5 and const: 3.
//Numeric rules, timestamp rules and duration rules
//The field to validate must match the rules type or its corresponding wrapper type (if any).
//If a lower bound (gt or gte) and an upper bound(lt or lte) are both specified, they must not be equal. If they are both inclusive (gte and lte), they must be replaced by const. Otherwise, all values are invalid.
//Durations and timestamps defined in options, such as (buf.validate.field).timestamp.lt, must be valid.
//If the rule is timestamp:
//within must be a positive duration.
//lt_now and gt_now must not both be specified.
//String rules
//The field to validate must be string or google.protobuf.StringValue.
//If len is specified, min_len or max_len must not be specified. If both are specified, min_len must be lower than max_len.
//If len_bytes is specified, min_bytes or max_bytes must not be specified. If both are specified, min_bytes must be lower than max_bytes.
//If min_len and max_bytes are both defined, min_len must be less than or equal to max_bytes. It's impossible for a string to have 3 or more UTF-8 characters while having less than 2 bytes.
//If min_bytes and max_len are both defined, min_bytes must be less than or equal to 4 times max_len. It's impossible for a string to have 2 or less UTF-8 characters while having 9 or more bytes, since each UTF-8 character takes at most 4 bytes.
//If prefix, suffix, or contains is specified, its length must not exceed max_len and max_bytes. Otherwise, all values are invalid.
//Any value of prefix, suffix and contains must not contain, or be a substring of not_contains, if they're both specified.
//If strict is set to false, well_known_regex must also be specified.
//If pattern is specified, is must be a valid regular expression in RE2 syntax.
//Bytes rules
//The field to validate must be bytes or google.protobuf.BytesValue.
//If len is specified, min_len or max_len must not be specified. If both are specified, min_len must be lower than max_len.
//If any of prefix, suffix and contains is specified, its length must not exceed max_len. Otherwise, all values are invalid.
//If pattern is specified, is must be a valid regular expression in RE2 syntax.
//Map rules
//The field to validate must be a map.
//min_pairs must not be higher than max_pairs.
//The set of rules in keys must be valid and compatible with the map field's key type.
//The set of rules in values must be valid and compatible with the map field's value type.
//Repeated rules
//The field to validate must have label repeated.
//min_items must not be higher than max_items.
//The set of rules in items must be compatible with the field's type.
//If unique is set to true, the field must be a scalar or a wrapper type.

// Validate checks that all protovalidate constraints specified are valid.
func (p *ProtoValidate) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	//TODO implement me
	panic("implement me")
}

// Message implements lint.Rule.
func (p *ProtoValidate) Message() string {
	panic("implement me")
}
