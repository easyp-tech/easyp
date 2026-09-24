package modules

import (
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// v1RequirementLine keeps the source spelling so edits preserve comments and
// whitespace outside the version or indirect marker being changed.
type v1RequirementLine struct {
	index      int
	body       string
	comment    string
	hasComment bool
	module     string
	version    string
}

func parseV1RequirementLines(lines []string) []v1RequirementLine {
	var requirements []v1RequirementLine
	block := ""
	for i, line := range lines {
		body, comment, hasComment := splitV1ManifestComment(line)
		trimmed := strings.TrimSpace(body)
		if strings.HasSuffix(trimmed, "(") {
			block = strings.TrimSpace(strings.TrimSuffix(trimmed, "("))
			continue
		}
		if trimmed == ")" {
			block = ""
			continue
		}
		if block != "" && block != "require" {
			continue
		}
		fields := strings.Fields(body)
		if block == "" {
			if len(fields) == 0 || fields[0] != "require" {
				continue
			}
			fields = fields[1:]
		}
		if len(fields) < 1 || len(fields) > 2 {
			continue
		}
		entry := v1RequirementLine{index: i, body: body, comment: comment, hasComment: hasComment, module: fields[0]}
		if len(fields) == 2 {
			entry.version = fields[1]
		}
		requirements = append(requirements, entry)
	}
	return requirements
}

func (line v1RequirementLine) withVersion(version string) v1RequirementLine {
	if version == line.version {
		return line
	}
	if version == "" {
		at := strings.LastIndex(line.body, line.version)
		line.body = strings.TrimRight(line.body[:at], " \t") + line.body[at+len(line.version):]
	} else if line.version == "" {
		at := strings.LastIndex(line.body, line.module) + len(line.module)
		line.body = line.body[:at] + " " + version + line.body[at:]
	} else {
		at := strings.LastIndex(line.body, line.version)
		line.body = line.body[:at] + version + line.body[at+len(line.version):]
	}
	line.version = version
	return line
}

func (line v1RequirementLine) indirect() bool {
	words := strings.Fields(line.comment)
	return line.hasComment && len(words) > 0 && words[0] == "indirect"
}

func (line v1RequirementLine) direct() v1RequirementLine {
	if !line.indirect() {
		return line
	}
	line.comment = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line.comment), "indirect"))
	line.hasComment = line.comment != ""
	if line.hasComment {
		line.comment = " " + line.comment
		return line
	}
	line.body = strings.TrimRight(line.body, " \t")
	return line
}

// directV1Module excludes requirements recorded as derived // indirect lines.
// The parsed module keeps roots and replacements, while explicit requirements
// remain the only inputs to dependency resolution during an update.
func directV1Module(original []byte, module v1.Module) v1.Module {
	module.Requires = nil
	for _, line := range parseV1RequirementLines(strings.Split(string(original), "\n")) {
		if !line.indirect() {
			module.Requires = append(module.Requires, v1.Requirement{Module: line.module, Version: line.version})
		}
	}
	return module
}

func (line v1RequirementLine) String() string {
	if line.hasComment {
		return line.body + "//" + line.comment
	}
	return line.body
}

type v1ManifestRequirement struct {
	v1.Requirement
	indirect bool
}

func appendV1Requirements(original []byte, requirements []v1ManifestRequirement) []byte {
	if len(requirements) == 0 {
		return original
	}
	updated := append([]byte(nil), original...)
	if len(updated) > 0 && updated[len(updated)-1] != '\n' {
		updated = append(updated, '\n')
	}
	for _, requirement := range requirements {
		line := "require " + requirement.Module
		if requirement.Version != "" {
			line += " " + requirement.Version
		}
		if requirement.indirect {
			line += " // indirect"
		}
		updated = append(updated, []byte(line+"\n")...)
	}
	return updated
}

func splitV1ManifestComment(line string) (string, string, bool) {
	for i := 0; i+1 < len(line); i++ {
		if line[i] == '/' && line[i+1] == '/' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
			return line[:i], line[i+2:], true
		}
	}
	return line, "", false
}
