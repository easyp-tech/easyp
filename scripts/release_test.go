package scripts

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type releaseConfig struct {
	Builds []struct {
		ID      string   `yaml:"id"`
		Ldflags []string `yaml:"ldflags"`
	} `yaml:"builds"`
	Release struct {
		Prerelease string `yaml:"prerelease"`
		MakeLatest string `yaml:"make_latest"`
	} `yaml:"release"`
	Dockers []struct {
		Tags  []string `yaml:"tags"`
		Flags []string `yaml:"flags"`
	} `yaml:"dockers_v2"`
	Brews []struct {
		SkipUpload string `yaml:"skip_upload"`
	} `yaml:"brews"`
}

type releaseTemplateData struct {
	Tag        string
	Version    string
	Prerelease string
	IsSnapshot bool
}

func TestReleaseChannels(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("../.goreleaser.yaml")
	require.NoError(t, err)
	var config releaseConfig
	err = yaml.Unmarshal(contents, &config)
	require.NoError(t, err)
	require.Equal(t, "true", config.Release.Prerelease)
	require.Len(t, config.Builds, 2)
	require.Len(t, config.Dockers, 1)
	require.Empty(t, config.Brews)

	tests := []struct {
		name       string
		data       releaseTemplateData
		wantLatest string
		wantTags   []string
		wantBuild  string
	}{
		{
			name:       "accidental stable version cannot update release defaults",
			data:       releaseTemplateData{Tag: "v1.0.0", Version: "1.0.0"},
			wantLatest: "false",
			wantTags:   []string{"v1.0.0"},
			wantBuild:  "v1.0.0",
		},
		{
			name:       "nightly requires an explicit tag",
			data:       releaseTemplateData{Tag: "v1.0.0-nightly.20261005.1", Version: "1.0.0-nightly.20261005.1", Prerelease: "nightly.20261005.1"},
			wantLatest: "false",
			wantTags:   []string{"v1.0.0-nightly.20261005.1"},
			wantBuild:  "v1.0.0-nightly.20261005.1",
		},
		{
			name:       "other prereleases cannot overwrite defaults",
			data:       releaseTemplateData{Tag: "v1.0.0-rc.1", Version: "1.0.0-rc.1", Prerelease: "rc.1"},
			wantLatest: "false",
			wantTags:   []string{"v1.0.0-rc.1"},
			wantBuild:  "v1.0.0-rc.1",
		},
		{
			name:       "snapshot keeps its build identity",
			data:       releaseTemplateData{Tag: "v0.17.0", Version: "SNAPSHOT-1234567", IsSnapshot: true},
			wantLatest: "false",
			wantTags:   []string{"v0.17.0"},
			wantBuild:  "SNAPSHOT-1234567",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.wantLatest, renderReleaseTemplate(t, config.Release.MakeLatest, tt.data))
			var tags []string
			for _, raw := range config.Dockers[0].Tags {
				tag := renderReleaseTemplate(t, raw, tt.data)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
			require.Equal(t, tt.wantTags, tags)
			for _, build := range config.Builds {
				flags := renderReleaseTemplate(t, strings.Join(build.Ldflags, " "), tt.data)
				var versions []string
				fields := strings.Fields(flags)
				for index, field := range fields {
					var assignment string
					switch {
					case field == "-X":
						require.Less(t, index+1, len(fields))
						assignment = fields[index+1]
					case strings.HasPrefix(field, "-X="):
						assignment = strings.TrimPrefix(field, "-X=")
					default:
						continue
					}
					value, ok := strings.CutPrefix(assignment, "github.com/easyp-tech/easyp/internal/version.releaseVersion=")
					if ok {
						versions = append(versions, value)
					}
				}
				require.Equal(t, []string{tt.wantBuild}, versions, build.ID)
			}
			var dockerVersions []string
			for index, raw := range config.Dockers[0].Flags {
				flag := renderReleaseTemplate(t, raw, tt.data)
				var assignment string
				switch {
				case flag == "--build-arg":
					require.Less(t, index+1, len(config.Dockers[0].Flags))
					assignment = renderReleaseTemplate(t, config.Dockers[0].Flags[index+1], tt.data)
				case strings.HasPrefix(flag, "--build-arg="):
					assignment = strings.TrimPrefix(flag, "--build-arg=")
				default:
					continue
				}
				value, ok := strings.CutPrefix(assignment, "RELEASE_VERSION=")
				if ok {
					dockerVersions = append(dockerVersions, value)
				}
			}
			require.Equal(t, []string{tt.wantBuild}, dockerVersions)
		})
	}
}

func renderReleaseTemplate(t *testing.T, raw string, data releaseTemplateData) string {
	t.Helper()

	tmpl, err := template.New("release").Option("missingkey=error").Parse(raw)
	require.NoError(t, err)
	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	require.NoError(t, err)
	return output.String()
}

func TestReleaseTagValidation(t *testing.T) {
	t.Parallel()

	script, err := filepath.Abs("validate-release-tag.sh")
	require.NoError(t, err)
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))
		return strings.TrimSpace(string(output))
	}
	git("init", "-b", "v1.0")
	git("-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "commit", "--allow-empty", "-m", "v1 nightly source")
	git("update-ref", "refs/remotes/origin/v1.0", "HEAD")
	git("branch", "unrelated")
	git("checkout", "-b", "main")
	git("-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "commit", "--allow-empty", "-m", "v1 source on main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	git("tag", "v1.0.0-nightly.20261005.4")
	git("checkout", "unrelated")
	git("-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "commit", "--allow-empty", "-m", "outside v1.0")
	git("tag", "v1.0.0-nightly.20261005.2")
	git("checkout", "v1.0")

	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{name: "v1 stable cannot be published", tag: "v1.0.0", want: "Unsupported release tag"},
		{name: "v0 stable uses its original release pipeline", tag: "v0.17.1", want: "Unsupported release tag"},
		{name: "nightly on v1.0", tag: "v1.0.0-nightly.20261005.1", want: "nightly\n", wantOK: true},
		{name: "nightly on main after merge", tag: "v1.0.0-nightly.20261005.4", want: "nightly\n", wantOK: true},
		{name: "nightly outside release branches", tag: "v1.0.0-nightly.20261005.2", want: "is not in origin/v1.0 or origin/main history"},
		{name: "rc", tag: "v1.0.0-rc.1", want: "Unsupported release tag"},
		{name: "other v1 nightly base", tag: "v1.1.0-nightly.20261005.1", want: "Unsupported release tag"},
		{name: "date missing", tag: "v1.0.0-nightly", want: "Unsupported release tag"},
		{name: "leading zero sequence", tag: "v1.0.0-nightly.20261005.01", want: "Unsupported release tag"},
		{name: "zero sequence", tag: "v1.0.0-nightly.20261005.0", want: "Unsupported release tag"},
		{name: "leading zero semver", tag: "v01.0.0", want: "Unsupported release tag"},
		{name: "nonexistent tag", tag: "v1.0.0-nightly.20261005.3", want: "does not resolve to a commit"},
	}
	for _, tt := range tests {
		if tt.tag != "v1.0.0-nightly.20261005.2" && tt.tag != "v1.0.0-nightly.20261005.3" && tt.tag != "v1.0.0-nightly.20261005.4" {
			git("tag", tt.tag)
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("bash", script, tt.tag)
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			if tt.wantOK {
				require.NoError(t, err, string(output))
				require.Equal(t, tt.want, string(output))
				return
			}
			require.Error(t, err)
			require.Contains(t, string(output), tt.want)
		})
	}

	branchTests := []struct {
		name string
		keep string
		tag  string
	}{
		{name: "main only after deleting v1.0", keep: "main", tag: "v1.0.0-nightly.20261005.4"},
		{name: "v1.0 only", keep: "v1.0", tag: "v1.0.0-nightly.20261005.1"},
		{name: "no source branches", tag: "v1.0.0-nightly.20261005.1"},
	}
	for _, tt := range branchTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clone := t.TempDir()
			cmd := exec.Command("git", "clone", "--no-hardlinks", "--local", root, clone)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, string(output))
			for _, branch := range []string{"main", "v1.0"} {
				if branch == tt.keep {
					continue
				}
				cmd = exec.Command("git", "update-ref", "-d", "refs/remotes/origin/"+branch)
				cmd.Dir = clone
				output, err = cmd.CombinedOutput()
				require.NoError(t, err, string(output))
			}
			cmd = exec.Command("bash", script, tt.tag)
			cmd.Dir = clone
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if tt.keep == "" {
				require.Error(t, err)
				require.Empty(t, stdout.String())
				require.Contains(t, stderr.String(), "is not in origin/v1.0 or origin/main history")
				return
			}
			require.NoError(t, err, stderr.String())
			require.Equal(t, "nightly\n", stdout.String())
			require.Empty(t, stderr.String())
		})
	}
}

func TestReleaseWorkflow(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("../.github/workflows/release.yml")
	require.NoError(t, err)
	var workflow struct {
		On   map[string]yaml.Node `yaml:"on"`
		Jobs map[string]struct {
			Steps []struct {
				Uses string            `yaml:"uses"`
				Run  string            `yaml:"run"`
				With map[string]string `yaml:"with"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	err = yaml.Unmarshal(contents, &workflow)
	require.NoError(t, err)
	require.Len(t, workflow.On, 1, "this workflow must publish only on explicit tag pushes")
	push, ok := workflow.On["push"]
	require.True(t, ok)
	var trigger struct {
		Tags []string `yaml:"tags"`
	}
	err = push.Decode(&trigger)
	require.NoError(t, err)
	require.Equal(t, []string{"v1.0.0-nightly.*"}, trigger.Tags)
	job, ok := workflow.Jobs["release"]
	require.True(t, ok)
	guard := -1
	checkedOut := false
	releaseAction := false
	for index, step := range job.Steps {
		if strings.HasPrefix(step.Uses, "actions/checkout@") {
			require.Equal(t, "0", step.With["fetch-depth"])
			checkedOut = true
		}
		if strings.Contains(step.Run, "scripts/validate-release-tag.sh") {
			require.True(t, checkedOut)
			require.Contains(t, step.Run, `"$GITHUB_REF_NAME"`)
			guard = index
		}
		if strings.HasPrefix(step.Uses, "docker/login-action@") || strings.HasPrefix(step.Uses, "goreleaser/goreleaser-action@") {
			require.GreaterOrEqual(t, guard, 0, "release must validate the tag before accessing publishing credentials")
			require.Less(t, guard, index)
		}
		if strings.HasPrefix(step.Uses, "goreleaser/goreleaser-action@") {
			require.Equal(t, "v2.18.2", step.With["version"])
			releaseAction = true
		}
	}
	require.True(t, releaseAction)
}
