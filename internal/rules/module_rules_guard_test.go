package rules_test

// Wave-0 ModuleRule-set guard (D-03, internal/core/lint_model.go), strengthened
// by the finishing batch 04-07 to EXACT-SET equality: with the registry
// complete (44/44 rules), exactly the eleven documented cross-file rules may
// implement core.ModuleRule — no more, no fewer. PackageNoImportCycle arrived
// with Phase 6 (Phase 06-02).

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

// moduleRuleNames is the D-03 eleven-name set as registry names.
var moduleRuleNames = map[string]bool{
	"DIRECTORY_SAME_PACKAGE":           true,
	"PACKAGE_NO_IMPORT_CYCLE":          true,
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

	// The registry is complete, so the whole construction must succeed: a
	// construction failure here is a registry gap, not a port-in-flight skip.
	got, _, err := rules.New(config.LintConfig{Use: rules.AllRuleNames()})
	require.NoError(t, err, "the linked registry is complete — every rule must construct")

	var gotNames []string
	for _, rule := range got {
		if _, ok := rule.(core.ModuleRule); ok {
			gotNames = append(gotNames, core.GetRuleName(rule))
		}
	}

	wantNames := make([]string, 0, len(moduleRuleNames))
	for name := range moduleRuleNames {
		wantNames = append(wantNames, name)
	}

	// Exact set: every ModuleRule the registry holds must be one of the D-03
	// eleven, and every D-03 eleven must implement ModuleRule — no more, no fewer.
	require.ElementsMatch(t, wantNames, gotNames,
		"the ModuleRule set must equal the D-03 eleven-name set exactly")
}
