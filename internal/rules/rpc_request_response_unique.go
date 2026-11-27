package rules

import (
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.ModuleRule = (*RPCRequestResponseUnique)(nil)

// RPCRequestResponseUnique this rule checks that every request and response type is used by one RPC only,
// across all linted files, and that no RPC uses one type for both. Every RPC sharing a type is reported.
type RPCRequestResponseUnique struct {
	methods []uniqueMethod
}

type uniqueMethod struct {
	path, name    string
	input, output string
	pos           core.Position
}

// Message implements lint.Rule.
func (r *RPCRequestResponseUnique) Message() string {
	return "request and response types must be unique across all RPCs"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (r *RPCRequestResponseUnique) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (r *RPCRequestResponseUnique) Collect(protoInfo core.ProtoInfo) error {
	pkg := protoInfo.File.PackageName()
	protoInfo.File.Walk(protoast.Visitor{RPC: func(service *protoast.Service, rpc *protoast.RPC) {
		input, output := qualifiedType(pkg, rpc.Request.Type), qualifiedType(pkg, rpc.Response.Type)
		if protoInfo.Resolved != nil {
			if types, ok := protoInfo.Resolved.RPCTypes[rpc.Pos.Offset]; ok {
				input, output = types[0], types[1]
			}
		}
		r.methods = append(r.methods, uniqueMethod{
			path:   protoInfo.Path,
			name:   joinName(pkg, service.Name+"."+rpc.Name),
			input:  input,
			output: output,
			pos:    rpc.Pos,
		})
	}})
	return nil
}

// Finish implements core.ModuleRule.
func (r *RPCRequestResponseUnique) Finish() []core.IssueInfo {
	var res []core.IssueInfo
	add := func(m uniqueMethod, typeName string) {
		for _, issue := range core.AppendIssue(nil, r, m.pos, typeName) {
			res = append(res, core.IssueInfo{Issue: issue, Path: m.path})
		}
	}
	users := make(map[string][]int)
	var types []string
	for i, m := range r.methods {
		if m.input == m.output {
			add(m, m.input)
		}
		for _, typeName := range []string{m.input, m.output} {
			if slices.Contains(users[typeName], i) {
				continue
			}
			if users[typeName] == nil {
				types = append(types, typeName)
			}
			users[typeName] = append(users[typeName], i)
		}
	}
	for _, typeName := range types {
		if len(users[typeName]) < 2 {
			continue
		}
		for _, i := range users[typeName] {
			add(r.methods[i], typeName)
		}
	}
	r.methods = nil
	return res
}

// qualifiedType approximates a type's full name from the syntax when the
// files could not be linked: absolute names lose their dot, others get the
// file's package.
func qualifiedType(pkg, written string) string {
	if strings.HasPrefix(written, ".") {
		return written[1:]
	}
	return joinName(pkg, written)
}

func joinName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}
