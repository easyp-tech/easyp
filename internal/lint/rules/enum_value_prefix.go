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

// Message implements lint.Rule.
func (e *EnumValuePrefix) Message() string {
	return "enum value prefix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumValuePrefix) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		prefix := pascalToUpperSnake(enum.EnumName)

		for _, enumValue := range enum.EnumBody.EnumFields {
			if !strings.HasPrefix(enumValue.Ident, prefix) {
				res = lint.AppendIssue(
					res,
					e,
					enumValue.Meta.Pos,
					enumValue.Ident,
					enumValue.Comments,
				)
			}
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			prefix := pascalToUpperSnake(enum.EnumName)

			for _, enumValue := range enum.EnumBody.EnumFields {
				if !strings.HasPrefix(enumValue.Ident, prefix) {
					res = lint.AppendIssue(
						res,
						e,
						enumValue.Meta.Pos,
						enumValue.Ident,
						enumValue.Comments,
					)
				}
			}
		}
	}

	return res, nil
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
