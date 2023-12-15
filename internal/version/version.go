package version

import "runtime/debug"

// System returns application version based on build info.
func System() string {
	bi, ver := buildVersion()
	switch {
	case bi == nil:
		return "(unknown)"
	case ver == "" || bi.Main.Version != "(devel)":
		return bi.Main.Version
	default:
		return ver
	}
}

func buildVersion() (bi *debug.BuildInfo, ver string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, ""
	}
	const revisionPrefix = 7
	revision := buildSetting(bi, "vcs.revision")
	modified := buildSetting(bi, "vcs.modified")
	time := buildSetting(bi, "vcs.time")
	if revision == "" {
		return bi, time
	} else if len(revision) > revisionPrefix {
		revision = revision[:revisionPrefix]
	}
	if modified != "false" {
		revision += "-modified"
	}
	if time == "" {
		return bi, revision
	}

	return bi, revision + " " + time
}

func buildSetting(bi *debug.BuildInfo, key string) string {
	for _, s := range bi.Settings {
		if s.Key == key {
			return s.Value
		}
	}

	return ""
}
