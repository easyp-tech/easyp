package version

import (
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/semver"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	develVersion     = "devel"
	unknownVersion   = "unknown"
	develMainVersion = "(devel)"
	revisionPrefix   = 7
)

type buildMetadata struct {
	easypVersion    string
	compilerVersion *pluginpb.Version
}

var (
	// releaseVersion is set by release builds using the Go linker's -X flag.
	releaseVersion string
	metadata       buildMetadata
	metadataOnce   sync.Once
)

// System returns the release version, falling back to embedded build info.
func System() string {
	return getBuildMetadata().easypVersion
}

// CompilerVersion returns compiler metadata used in CodeGeneratorRequest.
func CompilerVersion() *pluginpb.Version {
	return cloneVersion(getBuildMetadata().compilerVersion)
}

func getBuildMetadata() buildMetadata {
	metadataOnce.Do(func() {
		metadata = readBuildMetadata()
	})
	return metadata
}

func readBuildMetadata() buildMetadata {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		bi = nil
	}
	return buildMetadataFromBuildInfo(bi, releaseVersion)
}

func buildMetadataFromBuildInfo(bi *debug.BuildInfo, release string) buildMetadata {
	easypVersion := release
	if easypVersion == "" {
		easypVersion = easypVersionFromBuildInfo(bi)
	}

	return buildMetadata{
		easypVersion:    easypVersion,
		compilerVersion: compilerVersionFromEasyp(easypVersion),
	}
}

func cloneVersion(version *pluginpb.Version) *pluginpb.Version {
	if version == nil {
		// Defensive fallback: in normal flow compilerVersion is always initialized.
		return &pluginpb.Version{
			Major:  proto.Int32(0),
			Minor:  proto.Int32(0),
			Patch:  proto.Int32(0),
			Suffix: proto.String(buildCompilerSuffix(unknownVersion, "")),
		}
	}

	clone := &pluginpb.Version{
		Major: proto.Int32(version.GetMajor()),
		Minor: proto.Int32(version.GetMinor()),
		Patch: proto.Int32(version.GetPatch()),
	}
	if suffix := version.GetSuffix(); suffix != "" {
		clone.Suffix = proto.String(suffix)
	}

	return clone
}

// compilerVersionFromEasyp reports easyp itself as the compiler: its semantic
// version, with the suffix marking easyp and carrying any prerelease or build
// detail, so plugins print e.g. "v1.0.0-easyp".
func compilerVersionFromEasyp(easypVersion string) *pluginpb.Version {
	easypCore, easypSuffix := splitVersionCoreAndSuffix(easypVersion)
	major, minor, patch := semverMajorMinorPatch(easypCore)

	return &pluginpb.Version{
		Major:  proto.Int32(major),
		Minor:  proto.Int32(minor),
		Patch:  proto.Int32(patch),
		Suffix: proto.String(buildCompilerSuffix(easypCore, easypSuffix)),
	}
}

func semverMajorMinorPatch(rawVersion string) (int32, int32, int32) {
	version := strings.TrimSpace(rawVersion)
	if version == "" {
		return 0, 0, 0
	}

	version = strings.TrimPrefix(version, "v")
	version = strings.SplitN(version, "-", 2)[0]
	version = strings.SplitN(version, "+", 2)[0]

	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return 0, 0, 0
	}

	major, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return 0, 0, 0
	}

	minor, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return 0, 0, 0
	}

	patch, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return 0, 0, 0
	}

	return int32(major), int32(minor), int32(patch)
}

// buildCompilerSuffix names easyp and adds what the numbers cannot say: a
// non-release core such as "devel" or a revision, then the version suffix.
func buildCompilerSuffix(easypCore, easypSuffix string) string {
	parts := []string{"easyp"}
	if !semver.IsValid(easypCore) {
		parts = append(parts, sanitizeSuffixToken(easypCore))
	}
	if easypSuffix = sanitizeOptionalToken(easypSuffix); easypSuffix != "" {
		parts = append(parts, easypSuffix)
	}

	return strings.Join(parts, "-")
}

func sanitizeOptionalToken(raw string) string {
	value := sanitizeSuffixToken(raw)
	if value == unknownVersion {
		return ""
	}
	return value
}

func splitVersionCoreAndSuffix(raw string) (string, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return unknownVersion, ""
	}
	if trimmed == develMainVersion {
		return develVersion, ""
	}

	if semver.IsValid(trimmed) {
		core := trimmed
		if build := semver.Build(core); build != "" {
			core = strings.TrimSuffix(core, build)
		}
		if pre := semver.Prerelease(core); pre != "" {
			core = strings.TrimSuffix(core, pre)
		}

		suffix := strings.TrimPrefix(trimmed[len(core):], "-")
		suffix = strings.TrimPrefix(suffix, "+")
		return core, suffix
	}

	parts := strings.SplitN(trimmed, "-", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return trimmed, ""
}

func sanitizeSuffixToken(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return unknownVersion
	}
	if value == develMainVersion {
		value = develVersion
	}

	builder := strings.Builder{}
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-', r == '_', r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return unknownVersion
	}
	return result
}

func easypVersionFromBuildInfo(bi *debug.BuildInfo) string {
	if bi == nil {
		return develVersion
	}

	if mainVersion := strings.TrimSpace(bi.Main.Version); mainVersion != "" && mainVersion != develMainVersion {
		return replaceDirtyMarkers(mainVersion)
	}

	return vcsDerivedVersion(bi)
}

func vcsDerivedVersion(bi *debug.BuildInfo) string {
	if bi == nil {
		return develVersion
	}

	revision := strings.TrimSpace(buildSetting(bi, "vcs.revision"))
	if len(revision) > revisionPrefix {
		revision = revision[:revisionPrefix]
	}
	if revision == "" {
		return develVersion
	}

	if buildSetting(bi, "vcs.modified") == "true" {
		return revision + "-modified"
	}

	return revision
}

func replaceDirtyMarkers(value string) string {
	if value == "" {
		return value
	}

	parts := strings.Split(value, "-")
	for i, part := range parts {
		if strings.EqualFold(part, "dirty") {
			parts[i] = "modified"
		}
	}

	value = strings.Join(parts, "-")
	value = strings.ReplaceAll(value, "+dirty", "+modified")
	value = strings.ReplaceAll(value, "+DIRTY", "+modified")
	value = strings.ReplaceAll(value, ".dirty", ".modified")
	value = strings.ReplaceAll(value, ".DIRTY", ".modified")

	return value
}

func buildSetting(bi *debug.BuildInfo, key string) string {
	for _, s := range bi.Settings {
		if s.Key == key {
			return s.Value
		}
	}

	return ""
}
