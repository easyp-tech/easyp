package rules

import (
	"fmt"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"regexp"
)

var _ core.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase is a rule for checking name of enum for pascal case.
type EnumPascalCase struct{}

// Validate implements Rule.
func (c *EnumPascalCase) Validate(svc *unordered.Proto) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, enum := range svc.ProtoBody.Enums {
		if !pascalCase.MatchString(enum.EnumName) {
			res = append(res, &Error{
				Err: fmt.Errorf("%d:%d:%s: %w", enum.Meta.Pos.Line, enum.Meta.Pos.Column, enum.EnumName, core.ErrEnumPascalCase),
			})
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}
