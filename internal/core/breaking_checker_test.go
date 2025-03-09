package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"go.redsock.ru/protopack/internal/fs/fs"
)

const (
	originalDir  = "../../testdata/breaking_check/original"
	brokenDir    = "../../testdata/breaking_check/broken"
	notBrokenDir = "../../testdata/breaking_check/not_broken"
)

func TestCore_BreakingCheck(t *testing.T) {
	t.Parallel()

	c := &Core{}

	originalProtoData := readProtoData(t, c, originalDir)

	tests := map[string]struct {
		path       string
		wantIssues []IssueInfo
	}{
		"not_broken": {
			path:       notBrokenDir,
			wantIssues: nil,
		},
		"broken": {
			path: brokenDir,
			wantIssues: []IssueInfo{
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   39,
							Line:     5,
							Column:   1,
						},
						SourceName: "",
						Message:    "Previously import \"\"messages.proto\"\" was deleted.\n",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   39,
							Line:     5,
							Column:   1,
						},
						SourceName: "",
						Message:    "Previously present field \"1\" with name \"field_1\" on message \"RPC1Request\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "messages.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   65,
							Line:     7,
							Column:   1,
						},
						SourceName: "",
						Message:    "Previously present enum value \"2\" on enum \"SomeEnum\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   649,
							Line:     45,
							Column:   1,
						},
						SourceName: "",
						Message:    "Previously present RPC \"RPC2\" on service \"Service\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   498,
							Line:     35,
							Column:   1,
						},
						SourceName: "",
						Message:    "Previously present field \"2\" with name \"password\" on message \"AuthInfo\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   283,
							Line:     22,
							Column:   3,
						},
						SourceName: "",
						Message:    "Previously present field \"1\" with name \"rpc2_response_nested_field\" on message \"RPC2Response.RPC2ResponseNested\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
				{
					Issue: Issue{
						Position: meta.Position{
							Filename: "",
							Offset:   463,
							Line:     31,
							Column:   5,
						},
						SourceName: "",
						Message:    "Previously present field \"4\" with name \"rrr\" on OneOf \"RPC2Response.login\" was deleted.",
						RuleName:   breakingCheckRuleName,
					},
					Path: "services.proto",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoData := readProtoData(t, c, tc.path)

			breakingChecker := &BreakingChecker{
				against: originalProtoData,
				current: protoData,
			}

			issues, err := breakingChecker.Check()
			require.NoError(t, err)

			if tc.wantIssues == nil {
				require.Empty(t, issues)
				return
			}

			require.ElementsMatch(t, tc.wantIssues, issues)
		})
	}
}

func readProtoData(t *testing.T, c *Core, path string) ProtoData {
	fsWalker := fs.NewFSWalker(path, ".")
	protoInfo, err := c.readProtoFiles(context.Background(), fsWalker)
	require.NoError(t, err)

	protoData, err := collect(protoInfo)
	require.NoError(t, err)
	return protoData
}
