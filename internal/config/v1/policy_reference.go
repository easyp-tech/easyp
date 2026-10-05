package v1

import (
	"fmt"
	"path"
	"strings"
)

// PolicyReference is a parsed section-level extends value. Parsing is purely
// syntactic: it never touches Git, the module cache or the filesystem, so the
// parser and the published schema can accept the contract without performing
// any resolution.
type PolicyReference struct {
	// Raw is the trimmed reference exactly as written by the policy author.
	Raw string
	// Local marks ./ and ../ references resolved against the referring file.
	Local bool
	// Path is the local policy file or directory as written, for example
	// ./base.yaml, ../policies or ..
	Path string
	// Module is the declared module identity of a remote reference.
	Module string
	// Relative is the explicit # fragment of a remote reference. The RFC form
	// without a fragment is split by the resolver against known identities.
	Relative string
	// Fragment marks a reference written with an explicit # fragment.
	Fragment bool
}

// Empty reports whether no base policy is requested.
func (r PolicyReference) Empty() bool { return r.Raw == "" }

// String returns the reference as written.
func (r PolicyReference) String() string { return r.Raw }

// ParsePolicyReference validates a policy extends reference without resolving
// it. Accepted forms are:
//
//	./base.yaml, ../policies           local, relative to the referring policy
//	<module>                          exact module identity, its easyp.yaml
//	<module>#base.yaml|<module>#dir   explicit remote policy file or directory
//	<module>/dir                      RFC form: longest known module prefix
//
// Versions are rejected: a dependency version is declared exactly once, in
// protobuf.mod and protobuf.lock, and never repeated in a policy reference.
func ParsePolicyReference(raw string) (PolicyReference, error) {
	reference := PolicyReference{Raw: strings.TrimSpace(raw)}
	if reference.Raw == "" {
		return PolicyReference{}, nil
	}
	if err := portablePolicyText(reference.Raw); err != nil {
		return PolicyReference{}, err
	}
	if strings.Contains(reference.Raw, "@") {
		return PolicyReference{}, fmt.Errorf("policy extends %q must not repeat a version; declare it once in protobuf.mod and protobuf.lock", reference.Raw)
	}
	if isLocalPolicyReference(reference.Raw) {
		return parseLocalPolicyReference(reference)
	}
	return parseRemotePolicyReference(reference)
}

func isLocalPolicyReference(raw string) bool {
	return raw == "." || raw == ".." ||
		strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../")
}

// parseLocalPolicyReference accepts a policy file or directory relative to the
// referring policy file. "." and ".." are directories, so a directory policy can
// extend the easyp.yaml of its own or an ancestor directory; the resolver keeps
// the result inside the referring container and detects self-reference.
func parseLocalPolicyReference(reference PolicyReference) (PolicyReference, error) {
	reference.Local = true
	reference.Path = path.Clean(reference.Raw)
	return reference, nil
}

func parseRemotePolicyReference(reference PolicyReference) (PolicyReference, error) {
	module, relative, fragment, err := splitRemotePolicyReference(reference.Raw)
	if err != nil {
		return PolicyReference{}, err
	}
	reference.Module = module
	reference.Relative = relative
	reference.Fragment = fragment
	return reference, nil
}

func splitRemotePolicyReference(raw string) (module, relative string, fragment bool, err error) {
	if strings.Contains(raw, "#") {
		module, relative, _ = strings.Cut(raw, "#")
		fragment = true
		if strings.Contains(relative, "#") {
			return "", "", false, fmt.Errorf("policy extends %q has more than one # fragment", raw)
		}
		if relative != "" {
			if err := validateRemotePolicyPath(raw, relative); err != nil {
				return "", "", false, err
			}
		}
	} else {
		module = raw
	}
	module = strings.TrimSuffix(module, "/")
	if err := validateModuleIdentity(raw, module); err != nil {
		return "", "", false, err
	}
	return module, relative, fragment, nil
}

// validateModuleIdentity checks the shape of a remote module identity. A
// transport spelling is accepted only because a declared module identity may
// itself be a URL; no reference is ever fetched, so resolution is limited to
// identities already verified by the consumer graph.
func validateModuleIdentity(raw, module string) error {
	if module == "" {
		return fmt.Errorf("policy extends %q has no module identity", raw)
	}
	logical := module
	if scheme, rest, ok := strings.Cut(module, "://"); ok {
		if scheme == "" || strings.ContainsAny(scheme, "/ \t") {
			return fmt.Errorf("policy extends %q has an invalid module identity", raw)
		}
		if rest == "" {
			return fmt.Errorf("policy extends %q has an invalid module identity", raw)
		}
		logical = rest
	}
	logical = strings.Trim(logical, "/")
	if !strings.Contains(logical, "/") {
		return fmt.Errorf("policy extends %q must name a module identity such as github.com/acme/proto-policy, or a module declared in protobuf.mod", raw)
	}
	for _, element := range strings.Split(logical, "/") {
		if element == "" || element == "." || element == ".." {
			return fmt.Errorf("policy extends %q has an invalid module identity", raw)
		}
		if strings.ContainsAny(element, " \t") {
			return fmt.Errorf("policy extends %q has an invalid module identity", raw)
		}
	}
	return nil
}

// validateRemotePolicyPath keeps a module-relative policy path portable: it is
// resolved inside a verified module and must not reach outside it.
func validateRemotePolicyPath(raw, relative string) error {
	if strings.HasPrefix(relative, "/") || strings.HasPrefix(relative, "\\") {
		return fmt.Errorf("policy extends %q must use a module-relative policy path", raw)
	}
	if len(relative) > 1 && relative[1] == ':' {
		return fmt.Errorf("policy extends %q must use a module-relative policy path", raw)
	}
	if err := portablePolicyText(relative); err != nil {
		return err
	}
	for _, element := range strings.Split(relative, "/") {
		if element == "" {
			return fmt.Errorf("policy extends %q must use a portable module-relative policy path", raw)
		}
		if element == ".." {
			return fmt.Errorf("policy extends %q must not leave its module with .. segments", raw)
		}
	}
	if path.Clean(relative) == "." {
		return fmt.Errorf("policy extends %q selects no policy", raw)
	}
	return nil
}

// portablePolicyText rejects characters that would not survive a checked-in
// portable policy path or module identity.
func portablePolicyText(raw string) error {
	if strings.ContainsAny(raw, "\x00\r\n\\") {
		return fmt.Errorf("policy extends %q contains unsupported characters", raw)
	}
	return nil
}
