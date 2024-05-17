package rules

import (
	"strings"
	"unicode"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumValuePrefix)(nil)

// EnumValuePrefix this rule requires that all enum value names are prefixed with the enum name.
type EnumValuePrefix struct {
}

// Validate implements lint.Rule.
func (e EnumValuePrefix) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		prefix := pascalToUpperSnake(enum.EnumName)

		for _, enumValue := range enum.EnumBody.EnumFields {
			if !strings.HasPrefix(enumValue.Ident, prefix) {
				res = append(res, BuildError(enumValue.Meta.Pos, enumValue.Ident, ErrEnumValuePrefix))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}

func pascalToUpperSnake(s string) string {
	var result string

	for _, char := range s {
		if unicode.IsUpper(char) {
			if len(result) > 0 {
				result += "_"
			}
			result += string(char)
		} else {
			result += string(unicode.ToUpper(char))
		}
	}

	return result
}
