package rules_test

// Wave-0 ModuleRule-set guard (D-03, internal/core/lint_model.go): only the
// ten documented cross-file rules may implement core.ModuleRule. This
// membership guard catches an accidental ModuleRule implementation the moment
// it lands in the linked registry; the finishing batch (04-07) strengthens it
// to exact-set equality once all 41 rules are ported.

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

// moduleRuleNames is the D-03 ten-name set as registry names.
var moduleRuleNames = map[string]bool{
	"DIRECTORY_SAME_PACKAGE":           true,
	"PACKAGE_SAME_DIRECTORY":           true,
	"PACKAGE_SAME_CSHARP_NAMESPACE":    true,
	"PACKAGE_SAME_GO_PACKAGE":          true,
	"PACKAGE_SAME_JAVA_MULTIPLE_FILES": true,
	"PACKAGE_SAME_JAVA_PACKAGE":        true,
	"PACKAGE_SAME_PHP_NAMESPACE":       true,
	"PACKAGE_SAME_RUBY_PACKAGE":        true,
	"PACKAGE_SAME_SWIFT_PREFIX":        true,
	"RPC_REQUEST_RESPONSE_UNIQUE":      true,
}

func TestModuleRuleSet(t *testing.T) {
	t.Parallel()

	// The linked registry grows batch by batch, so each registry name is
	// probed individually: a not-yet-ported name fails construction here and
	// is skipped — that failure belongs to the batch's own requirePorted RED
	// guard, never to this green-today guard. Every ported rule that
	// implements core.ModuleRule must report a D-03 set name.
	for _, name := range rules.AllRuleNames() {
		got, _, err := rules.NewLinked(config.LintConfig{Use: []string{name}})
		if err != nil {
			// Not ported yet — allowed while the port is in flight.
			continue
		}
		for _, rule := range got {
			if _, ok := rule.(core.ModuleRule); !ok {
				continue
			}
			registryName := core.LinkedRuleName(rule)
			require.True(t, moduleRuleNames[registryName],
				"rule %s (from %s) implements core.ModuleRule but is not in the D-03 ten-name set",
				registryName, name)
		}
	}
}
