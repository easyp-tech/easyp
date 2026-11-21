---
phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
verified: 2026-11-18T22:28:47Z
status: passed
score: 23/23 must-haves verified
behavior_unverified: 0
overrides_applied: 0
prohibitions_flagged: 3 # judgment-tier prohibitions recorded with NON-AUTHORITATIVE LLM-judge verdicts — unverified-prohibition, human review recommended (see human_verification)
covered_files:
  - .golangci.yml
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-01-PLAN.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-01-SUMMARY.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-02-PLAN.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-02-SUMMARY.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-03-PLAN.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-03-SUMMARY.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-04-PLAN.md
  - .planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-04-SUMMARY.md
  - AGENTS.md
  - go.mod
  - internal/core/breaking_check.go
  - internal/core/breaking_checker.go
  - internal/core/breaking_checker_test.go
  - internal/core/breaking_profiles.go
  - internal/core/breaking_target_paths_test.go
  - internal/core/dom.go
  - internal/core/lint.go
  - internal/core/lint_proto_info.go
  - internal/core/position_parity_test.go
  - internal/core/proto_info_read.go
  - internal/core/proto_info_read_test.go
  - internal/core/testdata/position_contract_utf8.proto
  - internal/easypc/VENDORED.md
  - internal/easypc/cmd/easyp-protoc/main.go
  - internal/easypc/cmd/lexbench/main.go
  - internal/easypc/pkg/compiler/driver.go
  - internal/easypc/pkg/compiler/driver_test.go
  - internal/easypc/pkg/compiler/errors_test.go
  - internal/easypc/pkg/compiler/options.go
  - internal/easypc/pkg/lexer/lexer.go
  - internal/easypc/pkg/lexer/lexer_test.go
  - internal/easypc/pkg/lexer/token.go
  - internal/easypc/pkg/linker/linker_test.go
  - internal/easypc/pkg/linker/resolve.go
  - internal/easypc/pkg/linker/symbol.go
  - internal/easypc/pkg/linker/table.go
  - internal/easypc/pkg/linker/validate.go
  - internal/easypc/pkg/linker/visibility.go
  - internal/easypc/pkg/metrics/metrics.go
  - internal/easypc/pkg/options/evaluator.go
  - internal/easypc/pkg/options/options_test.go
  - internal/easypc/pkg/options/standard.go
  - internal/easypc/pkg/options/wire.go
  - internal/easypc/pkg/parser/ast.go
  - internal/easypc/pkg/parser/json_name.go
  - internal/easypc/pkg/parser/parser.go
  - internal/easypc/pkg/parser/parser_test.go
  - internal/easypc/pkg/parser/raw_spans_test.go
  - internal/easypc/pkg/serialize/wire.go
  - internal/easypc/pkg/serialize/wire_test.go
  - internal/easypc/pkg/sourceinfo/comments.go
  - internal/easypc/pkg/sourceinfo/generator.go
  - internal/easypc/pkg/sourceinfo/serialize.go
  - internal/easypc/pkg/sourceinfo/span.go
  - internal/easypc/pkg/sourceinfo/types.go
  - internal/easypc/protoc/compile.go
  - internal/easypc/protoc/diagnostic.go
  - internal/easypc/protoc/link.go
  - internal/easypc/protoc/link_test.go
  - internal/easypc/protoc/lsp.go
  - internal/easypc/protoc/malformed_decl_test.go
  - internal/easypc/protoc/options.go
  - internal/easypc/protoc/protoc_test.go
  - internal/easypc/protoc/resolve_shadow_test.go
  - internal/easypc/protoc/resolver.go
  - internal/easypc/protoc/result.go
  - internal/easypc/protoc/semantic_test.go
  - internal/easypc/protoc/string_concat_test.go
  - internal/easypc/protoc/syntax/syntax.go
  - internal/easypc/protoc/unresolved_refs_test.go
  - internal/easypc/protoc/wkt/google/protobuf/any.proto
  - internal/easypc/protoc/wkt/google/protobuf/api.proto
  - internal/easypc/protoc/wkt/google/protobuf/c_sharp_features.proto
  - internal/easypc/protoc/wkt/google/protobuf/compiler/plugin.proto
  - internal/easypc/protoc/wkt/google/protobuf/cpp_features.proto
  - internal/easypc/protoc/wkt/google/protobuf/cpp_file_options.proto
  - internal/easypc/protoc/wkt/google/protobuf/cpp_options.proto
  - internal/easypc/protoc/wkt/google/protobuf/descriptor.proto
  - internal/easypc/protoc/wkt/google/protobuf/duration.proto
  - internal/easypc/protoc/wkt/google/protobuf/empty.proto
  - internal/easypc/protoc/wkt/google/protobuf/field_mask.proto
  - internal/easypc/protoc/wkt/google/protobuf/go_features.proto
  - internal/easypc/protoc/wkt/google/protobuf/java_features.proto
  - internal/easypc/protoc/wkt/google/protobuf/json_enumvalue_options.proto
  - internal/easypc/protoc/wkt/google/protobuf/json_options.proto
  - internal/easypc/protoc/wkt/google/protobuf/source_context.proto
  - internal/easypc/protoc/wkt/google/protobuf/struct.proto
  - internal/easypc/protoc/wkt/google/protobuf/timestamp.proto
  - internal/easypc/protoc/wkt/google/protobuf/type.proto
  - internal/easypc/protoc/wkt/google/protobuf/wrappers.proto
  - internal/easypc/protoc/wkt/wkt.go
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe1_baseline.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe2_tag_perturbed.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe3_type_name_perturbed.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe4_unordered_fields.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe5_options_defaults.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe6_nested_oneof.proto
  - internal/easypc/testdata/audit_m2_iter2_perturb/probe7_proto2_extensions.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe1_perturbed_tags_names.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe2_added_fields_types.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe3_nested_messages_enums.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe4_oneofs_and_synthetic_proto3.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe5_proto2_defaults_quotes.proto
  - internal/easypc/testdata/audit_m2_iter4_perturb/probe6_out_of_order_and_reserved.proto
  - internal/easypc/testdata/audit_perturb/dynamic_test.proto
  - internal/easypc/testdata/audit_perturb/perturb_base.proto
  - internal/easypc/testdata/audit_perturb/perturb_editions.proto
  - internal/easypc/testdata/audit_perturb/perturb_mod_tags.proto
  - internal/easypc/testdata/audit_perturb/perturb_nested.proto
  - internal/easypc/testdata/audit_perturb/perturb_proto2.proto
  - internal/easypc/testdata/audit_probes/probe1_ext.proto
  - internal/easypc/testdata/audit_probes/probe1_options.proto
  - internal/easypc/testdata/audit_probes/probe2_imported.proto
  - internal/easypc/testdata/audit_probes/probe2_root.proto
  - internal/easypc/testdata/audit_probes/probe3_oneofs.proto
  - internal/easypc/testdata/e2e/tier1/feat01_tokens/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat01_tokens/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat01_tokens/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat01_tokens/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat01_tokens/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat02_strings/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat02_strings/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat02_strings/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat02_strings/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat02_strings/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat03_numbers/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat03_numbers/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat03_numbers/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat03_numbers/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat03_numbers/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat04_comments/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat04_comments/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat04_comments/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat04_comments/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat04_comments/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat05_special/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat05_special/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat05_special/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat05_special/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat05_special/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat06_lexbench/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat06_lexbench/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat06_lexbench/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat06_lexbench/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat06_lexbench/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat07_vector/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat07_vector/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat07_vector/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat07_vector/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat07_vector/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat08_grammar/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat08_grammar/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat08_grammar/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat08_grammar/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat08_grammar/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat09_flat_ir/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat09_flat_ir/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat09_flat_ir/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat09_flat_ir/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat09_flat_ir/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat10_wire/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat10_wire/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat10_wire/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat10_wire/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat10_wire/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat11_cli/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat11_cli/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat11_cli/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat11_cli/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat11_cli/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat12_symbols/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat12_symbols/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat12_symbols/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat12_symbols/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat12_symbols/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/dep01.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/dep02a.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/dep02b.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/diamond_b.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/diamond_c.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/diamond_d.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/pub_gateway.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/pub_inner.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/trans_b.proto
  - internal/easypc/testdata/e2e/tier1/feat13_imports/trans_c.proto
  - internal/easypc/testdata/e2e/tier1/feat14_types/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat14_types/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat14_types/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat14_types/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat14_types/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat15_options/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat15_options/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat15_options/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat15_options/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat15_options/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat16_wkt/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat16_wkt/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat16_wkt/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat16_wkt/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat16_wkt/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat17_spans/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat17_spans/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat17_spans/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat17_spans/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat17_spans/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat18_comments/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat18_comments/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat18_comments/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat18_comments/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat18_comments/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat19_json_name/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat19_json_name/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat19_json_name/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat19_json_name/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat19_json_name/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat20_metrics/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat20_metrics/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat20_metrics/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat20_metrics/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat20_metrics/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat21_diag_format/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat21_diag_format/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat21_diag_format/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat21_diag_format/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat21_diag_format/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat22_early_term/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat22_early_term/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat22_early_term/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat22_early_term/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat22_early_term/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat23_diag_spec/unused_lib.proto
  - internal/easypc/testdata/e2e/tier1/feat24_go_lib/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat24_go_lib/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat24_go_lib/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat24_go_lib/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat24_go_lib/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat25_multi_strat/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat25_multi_strat/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat25_multi_strat/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat25_multi_strat/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat25_multi_strat/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat26_multi_var/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat26_multi_var/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat26_multi_var/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat26_multi_var/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat26_multi_var/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat27_determinism/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat27_determinism/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat27_determinism/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat27_determinism/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat27_determinism/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat28_perf_parity/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat28_perf_parity/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat28_perf_parity/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat28_perf_parity/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat28_perf_parity/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat29_e2e_pass/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat29_e2e_pass/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat29_e2e_pass/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat29_e2e_pass/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat29_e2e_pass/case05.proto
  - internal/easypc/testdata/e2e/tier1/feat30_adversarial/case01.proto
  - internal/easypc/testdata/e2e/tier1/feat30_adversarial/case02.proto
  - internal/easypc/testdata/e2e/tier1/feat30_adversarial/case03.proto
  - internal/easypc/testdata/e2e/tier1/feat30_adversarial/case04.proto
  - internal/easypc/testdata/e2e/tier1/feat30_adversarial/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat01_tokens/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat01_tokens/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat01_tokens/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat01_tokens/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat01_tokens/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat02_strings/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat02_strings/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat02_strings/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat02_strings/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat02_strings/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat03_numbers/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat03_numbers/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat03_numbers/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat03_numbers/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat03_numbers/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat04_comments/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat04_comments/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat04_comments/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat04_comments/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat04_comments/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat05_special/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat05_special/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat05_special/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat05_special/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat05_special/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat06_lexbench/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat06_lexbench/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat06_lexbench/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat06_lexbench/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat06_lexbench/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat07_vector/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat07_vector/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat07_vector/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat07_vector/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat07_vector/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat08_grammar/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat08_grammar/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat08_grammar/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat08_grammar/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat08_grammar/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat09_flat_ir/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat09_flat_ir/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat09_flat_ir/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat09_flat_ir/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat09_flat_ir/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat10_wire/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat10_wire/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat10_wire/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat10_wire/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat10_wire/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat11_cli/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat11_cli/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat11_cli/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat11_cli/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat11_cli/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat12_symbols/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat12_symbols/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat12_symbols/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat12_symbols/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat12_symbols/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/a_dep.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/chain_b.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/chain_c.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/chain_d.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/chain_e.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/common_base.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/dep_solo.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/pub_l1.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/pub_l2.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/pub_l3.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/user_1.proto
  - internal/easypc/testdata/e2e/tier2/feat13_imports/z_dep.proto
  - internal/easypc/testdata/e2e/tier2/feat14_types/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat14_types/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat14_types/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat14_types/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat14_types/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat15_options/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat15_options/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat15_options/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat15_options/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat15_options/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat16_wkt/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat16_wkt/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat16_wkt/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat16_wkt/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat16_wkt/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat17_spans/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat17_spans/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat17_spans/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat17_spans/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat17_spans/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat18_comments/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat18_comments/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat18_comments/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat18_comments/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat18_comments/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat19_json_name/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat19_json_name/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat19_json_name/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat19_json_name/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat19_json_name/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat20_metrics/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat20_metrics/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat20_metrics/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat20_metrics/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat20_metrics/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat21_diag_format/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat21_diag_format/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat21_diag_format/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat21_diag_format/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat21_diag_format/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat22_early_term/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat22_early_term/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat22_early_term/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat22_early_term/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat22_early_term/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/unused_helper_1.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/unused_helper_2.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/unused_helper_3.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/unused_helper_4.proto
  - internal/easypc/testdata/e2e/tier2/feat23_diag_spec/unused_helper_5.proto
  - internal/easypc/testdata/e2e/tier2/feat24_go_lib/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat24_go_lib/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat24_go_lib/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat24_go_lib/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat24_go_lib/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat25_multi_strat/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat25_multi_strat/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat25_multi_strat/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat25_multi_strat/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat25_multi_strat/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat26_multi_var/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat26_multi_var/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat26_multi_var/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat26_multi_var/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat26_multi_var/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat27_determinism/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat27_determinism/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat27_determinism/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat27_determinism/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat27_determinism/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat28_perf_parity/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat28_perf_parity/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat28_perf_parity/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat28_perf_parity/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat28_perf_parity/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat29_e2e_pass/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat29_e2e_pass/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat29_e2e_pass/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat29_e2e_pass/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat29_e2e_pass/case05.proto
  - internal/easypc/testdata/e2e/tier2/feat30_adversarial/case01.proto
  - internal/easypc/testdata/e2e/tier2/feat30_adversarial/case02.proto
  - internal/easypc/testdata/e2e/tier2/feat30_adversarial/case03.proto
  - internal/easypc/testdata/e2e/tier2/feat30_adversarial/case04.proto
  - internal/easypc/testdata/e2e/tier2/feat30_adversarial/case05.proto
  - internal/easypc/testdata/e2e/tier3/combo01_ext_options.proto
  - internal/easypc/testdata/e2e/tier3/combo02_synthetic_oneof_json.proto
  - internal/easypc/testdata/e2e/tier3/combo03_wkt_maps_options.proto
  - internal/easypc/testdata/e2e/tier3/combo04_dep.proto
  - internal/easypc/testdata/e2e/tier3/combo04_nesting_imports.proto
  - internal/easypc/testdata/e2e/tier3/combo05_editions_features.proto
  - internal/easypc/testdata/e2e/tier3/combo06_str_concat_spans.proto
  - internal/easypc/testdata/e2e/tier3/combo07_hex_oct_ext.proto
  - internal/easypc/testdata/e2e/tier3/combo08_unicode_tabs.proto
  - internal/easypc/testdata/e2e/tier3/combo09_gateway.proto
  - internal/easypc/testdata/e2e/tier3/combo09_lib.proto
  - internal/easypc/testdata/e2e/tier3/combo09_public_import_types.proto
  - internal/easypc/testdata/e2e/tier3/combo10_recursive_oneof_json.proto
  - internal/easypc/testdata/e2e/tier3/combo11_all_wire_scalars.proto
  - internal/easypc/testdata/e2e/tier3/combo12_streaming_options.proto
  - internal/easypc/testdata/e2e/tier3/combo13_reserved_everything.proto
  - internal/easypc/testdata/e2e/tier3/combo14_branch_a.proto
  - internal/easypc/testdata/e2e/tier3/combo14_branch_b.proto
  - internal/easypc/testdata/e2e/tier3/combo14_core.proto
  - internal/easypc/testdata/e2e/tier3/combo14_diamond_root.proto
  - internal/easypc/testdata/e2e/tier3/combo15_maps_wkt_any.proto
  - internal/easypc/testdata/e2e/tier3/combo16_oneofs_proto2_defaults.proto
  - internal/easypc/testdata/e2e/tier3/combo17_large_ident_simd.proto
  - internal/easypc/testdata/e2e/tier3/combo18_all_level_options.proto
  - internal/easypc/testdata/e2e/tier3/combo19_unneeded.proto
  - internal/easypc/testdata/e2e/tier3/combo19_unused_import_warning.proto
  - internal/easypc/testdata/e2e/tier3/combo20_neg_enum_options.proto
  - internal/easypc/testdata/e2e/tier3/combo21_a.proto
  - internal/easypc/testdata/e2e/tier3/combo21_b.proto
  - internal/easypc/testdata/e2e/tier3/combo21_main.proto
  - internal/easypc/testdata/e2e/tier3/combo22_deep_packages.proto
  - internal/easypc/testdata/e2e/tier3/combo23_deprecated_extensions.proto
  - internal/easypc/testdata/e2e/tier3/combo24_proto2_groups.proto
  - internal/easypc/testdata/e2e/tier3/combo25_wkt_struct_dynamic.proto
  - internal/easypc/testdata/e2e/tier3/combo26_proto2_dep.proto
  - internal/easypc/testdata/e2e/tier3/combo26_proto3_root.proto
  - internal/easypc/testdata/e2e/tier3/combo27_json_repeated_maps.proto
  - internal/easypc/testdata/e2e/tier3/combo28_detached_header_comments.proto
  - internal/easypc/testdata/e2e/tier3/combo29_max_tag_reserved.proto
  - internal/easypc/testdata/e2e/tier3/combo30_full_blend.proto
  - internal/easypc/testdata/e2e/tier4/scenario01_hft_market_data.proto
  - internal/easypc/testdata/e2e/tier4/scenario02_cloud_k8s_control_plane.proto
  - internal/easypc/testdata/e2e/tier4/scenario03_telemetry_pipeline.proto
  - internal/easypc/testdata/e2e/tier4/scenario04_ecommerce_checkout.proto
  - internal/easypc/testdata/e2e/tier4/scenario05_geospatial_routing.proto
  - internal/easypc/testdata/e2e/tier4/scenario06_healthcare_ehr.proto
  - internal/easypc/testdata/e2e/tier4/scenario07_distributed_iam.proto
  - internal/easypc/testdata/e2e/tier4/scenario08_iot_sensor_stream.proto
  - internal/easypc/testdata/e2e/tier4/scenario09_social_graph.proto
  - internal/easypc/testdata/e2e/tier4/scenario10_banking_ledger.proto
  - internal/easypc/testdata/e2e/tier4/scenario11_webrtc_signalling.proto
  - internal/easypc/testdata/e2e/tier4/scenario12_ml_inference_serving.proto
  - internal/easypc/testdata/e2e/tier4/scenario13_autonomous_vehicle.proto
  - internal/easypc/testdata/e2e/tier4/scenario14_logistics_manifest.proto
  - internal/easypc/testdata/e2e/tier4/scenario15_multiplayer_gaming.proto
  - internal/easypc/testdata/errors/cycle_a.golden
  - internal/easypc/testdata/errors/cycle_a.proto
  - internal/easypc/testdata/errors/cycle_b.proto
  - internal/easypc/testdata/errors/duplicate_field_name.golden
  - internal/easypc/testdata/errors/duplicate_field_name.proto
  - internal/easypc/testdata/errors/duplicate_field_number.golden
  - internal/easypc/testdata/errors/duplicate_field_number.proto
  - internal/easypc/testdata/errors/duplicate_message.golden
  - internal/easypc/testdata/errors/duplicate_message.proto
  - internal/easypc/testdata/errors/invalid_option_value.golden
  - internal/easypc/testdata/errors/invalid_option_value.proto
  - internal/easypc/testdata/errors/null_byte.golden
  - internal/easypc/testdata/errors/null_byte.proto
  - internal/easypc/testdata/errors/syntax_invalid.golden
  - internal/easypc/testdata/errors/syntax_invalid.proto
  - internal/easypc/testdata/errors/tabstop8_col.golden
  - internal/easypc/testdata/errors/tabstop8_col.proto
  - internal/easypc/testdata/errors/tag_large.golden
  - internal/easypc/testdata/errors/tag_large.proto
  - internal/easypc/testdata/errors/tag_reserved.golden
  - internal/easypc/testdata/errors/tag_reserved.proto
  - internal/easypc/testdata/errors/tag_zero.golden
  - internal/easypc/testdata/errors/tag_zero.proto
  - internal/easypc/testdata/errors/unclosed_comment.golden
  - internal/easypc/testdata/errors/unclosed_comment.proto
  - internal/easypc/testdata/errors/unclosed_string.golden
  - internal/easypc/testdata/errors/unclosed_string.proto
  - internal/easypc/testdata/errors/undefined_type.golden
  - internal/easypc/testdata/errors/undefined_type.proto
  - internal/easypc/testdata/lexer/01_ident.proto
  - internal/easypc/testdata/lexer/01_ident.tokens
  - internal/easypc/testdata/lexer/02_int_dec.proto
  - internal/easypc/testdata/lexer/02_int_dec.tokens
  - internal/easypc/testdata/lexer/03_int_oct.proto
  - internal/easypc/testdata/lexer/03_int_oct.tokens
  - internal/easypc/testdata/lexer/04_int_hex.proto
  - internal/easypc/testdata/lexer/04_int_hex.tokens
  - internal/easypc/testdata/lexer/05_float.proto
  - internal/easypc/testdata/lexer/05_float.tokens
  - internal/easypc/testdata/lexer/06_string_simple.proto
  - internal/easypc/testdata/lexer/06_string_simple.tokens
  - internal/easypc/testdata/lexer/07_string_escapes.proto
  - internal/easypc/testdata/lexer/07_string_escapes.tokens
  - internal/easypc/testdata/lexer/08_string_utf8.proto
  - internal/easypc/testdata/lexer/08_string_utf8.tokens
  - internal/easypc/testdata/lexer/09_punct.proto
  - internal/easypc/testdata/lexer/09_punct.tokens
  - internal/easypc/testdata/lexer/10_line_comment.proto
  - internal/easypc/testdata/lexer/10_line_comment.tokens
  - internal/easypc/testdata/lexer/11_block_comment.proto
  - internal/easypc/testdata/lexer/11_block_comment.tokens
  - internal/easypc/testdata/lexer/12_whitespace.proto
  - internal/easypc/testdata/lexer/12_whitespace.tokens
  - internal/easypc/testdata/lexer/13_bom_start.proto
  - internal/easypc/testdata/lexer/13_bom_start.tokens
  - internal/easypc/testdata/lexer/14_null_in_comment.proto
  - internal/easypc/testdata/lexer/14_null_in_comment.tokens
  - internal/easypc/testdata/lexer/15_error_bad_chars.proto
  - internal/easypc/testdata/lexer/15_error_bad_chars.tokens
  - internal/easypc/testdata/lexer/16_error_null_byte.proto
  - internal/easypc/testdata/lexer/16_error_null_byte.tokens
  - internal/easypc/testdata/lexer/17_error_bom_middle.proto
  - internal/easypc/testdata/lexer/17_error_bom_middle.tokens
  - internal/easypc/testdata/lexer/18_error_bad_number.proto
  - internal/easypc/testdata/lexer/18_error_bad_number.tokens
  - internal/easypc/testdata/lexer/19_error_unclosed_string.proto
  - internal/easypc/testdata/lexer/19_error_unclosed_string.tokens
  - internal/easypc/testdata/lexer/20_error_bad_string_escape.proto
  - internal/easypc/testdata/lexer/20_error_bad_string_escape.tokens
  - internal/easypc/testdata/lexer/21_error_unclosed_comment.proto
  - internal/easypc/testdata/lexer/21_error_unclosed_comment.tokens
  - internal/easypc/testdata/lexer/22_error_invalid_utf8.proto
  - internal/easypc/testdata/lexer/22_error_invalid_utf8.tokens
  - internal/easypc/testdata/lexer/23_kitchen_sink.proto
  - internal/easypc/testdata/lexer/23_kitchen_sink.tokens
  - internal/easypc/testdata/lexer/24_string_escaped_backslash.proto
  - internal/easypc/testdata/lexer/24_string_escaped_backslash.tokens
  - internal/easypc/testdata/lexer/25_error_string_eof_escaped_quote_dq.proto
  - internal/easypc/testdata/lexer/25_error_string_eof_escaped_quote_dq.tokens
  - internal/easypc/testdata/lexer/26_error_string_eof_escaped_quote_sq.proto
  - internal/easypc/testdata/lexer/26_error_string_eof_escaped_quote_sq.tokens
  - internal/easypc/testdata/lexer/27_error_string_eof_lone_quote_dq.proto
  - internal/easypc/testdata/lexer/27_error_string_eof_lone_quote_dq.tokens
  - internal/easypc/testdata/lexer/28_error_string_eof_lone_quote_sq.proto
  - internal/easypc/testdata/lexer/28_error_string_eof_lone_quote_sq.tokens
  - internal/easypc/testdata/lexer/29_error_string_eof_esc_backslash_quote_dq.proto
  - internal/easypc/testdata/lexer/29_error_string_eof_esc_backslash_quote_dq.tokens
  - internal/easypc/testdata/lexer/30_error_string_eof_esc_backslash_quote_sq.proto
  - internal/easypc/testdata/lexer/30_error_string_eof_esc_backslash_quote_sq.tokens
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/b.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/c.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/d.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_l1.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_l2.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_l3.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_l4.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_l5.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/deep_root.proto
  - internal/easypc/testdata/m3_adversarial/adv01_diamond/root.proto
  - internal/easypc/testdata/m3_adversarial/adv02_public_imports/a.proto
  - internal/easypc/testdata/m3_adversarial/adv02_public_imports/b.proto
  - internal/easypc/testdata/m3_adversarial/adv02_public_imports/c.proto
  - internal/easypc/testdata/m3_adversarial/adv02_public_imports/d.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/pkg_shadow_lib.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/pkg_shadow_test.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/shadow_imported.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/shadow_test.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/sibling_consumer.proto
  - internal/easypc/testdata/m3_adversarial/adv03_scoping/sibling_helper.proto
  - internal/easypc/testdata/m3_adversarial/adv04_custom_options/options_def.proto
  - internal/easypc/testdata/m3_adversarial/adv04_custom_options/options_use.proto
  - internal/easypc/testdata/m3_adversarial/adv04_custom_options/options_use_braces.proto
  - internal/easypc/testdata/m3_adversarial/adv05_method_bodies/method_bodies.proto
  - internal/easypc/testdata/m3_adversarial/adv06_explicit_booleans/booleans.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_brackets_parens.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_def.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_delimiters.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_explicit_booleans.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_extremes.proto
  - internal/easypc/testdata/m3_adversarial/adv07_deep_stress/deep_oneofs.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/float_defaults_stress.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_bracket_extensions.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_combined_matrix.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_def.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_delimiters.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_extremes.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_scoping_base.proto
  - internal/easypc/testdata/m3_adversarial/adv08_iter3_stress/iter3_scoping_main.proto
  - internal/easypc/testdata/m4_adversarial/sci01_comments.proto
  - internal/easypc/testdata/m4_adversarial/sci02_tabs_utf8.proto
  - internal/easypc/testdata/m4_adversarial/sci03_deep_declarations.proto
  - internal/easypc/testdata/m4_adversarial/sci03b_zig_unqualified_opt.proto
  - internal/easypc/testdata/m4_adversarial/sci04_enum_negatives.proto
  - internal/easypc/testdata/m4_adversarial/sci04a_rust_neg_span.proto
  - internal/easypc/testdata/m4_adversarial/sci04b_go_min_int32.proto
  - internal/easypc/testdata/m4_adversarial/sci05_json_name_matrix.proto
  - internal/easypc/testdata/m5_adversarial/cycle3_a.proto
  - internal/easypc/testdata/m5_adversarial/cycle3_b.proto
  - internal/easypc/testdata/m5_adversarial/cycle3_c.proto
  - internal/easypc/testdata/m5_adversarial/cycle4_a.proto
  - internal/easypc/testdata/m5_adversarial/cycle4_b.proto
  - internal/easypc/testdata/m5_adversarial/cycle4_c.proto
  - internal/easypc/testdata/m5_adversarial/cycle4_d.proto
  - internal/easypc/testdata/m5_adversarial/duplicates_test.proto
  - internal/easypc/testdata/m5_adversarial/self_cycle.proto
  - internal/easypc/testdata/m5_adversarial/tab_stress.proto
  - internal/easypc/testdata/m5_adversarial/tag_0.proto
  - internal/easypc/testdata/m5_adversarial/tag_18999_20000_valid.proto
  - internal/easypc/testdata/m5_adversarial/tag_19000.proto
  - internal/easypc/testdata/m5_adversarial/tag_19999.proto
  - internal/easypc/testdata/m5_adversarial/tag_536870911_valid.proto
  - internal/easypc/testdata/m5_adversarial/tag_536870912.proto
  - internal/easypc/testdata/m5_adversarial/tag_neg.proto
  - internal/easypc/testdata/semantic/custom_option_ok.proto
  - internal/easypc/testdata/semantic/custom_option_unknown.proto
  - internal/easypc/testdata/semantic/custom_option_wrong_type.proto
  - internal/easypc/testdata/semantic/duplicate_extension_number.proto
  - internal/easypc/testdata/semantic/enum_value_scope_conflict.proto
  - internal/easypc/testdata/semantic/extension_number_out_of_range.proto
  - internal/easypc/testdata/semantic/field_and_message_same_name.proto
  - internal/easypc/testdata/semantic/json_name_conflict.proto
  - internal/easypc/testdata/semantic/proto3_default.proto
  - internal/easypc/testdata/semantic/proto3_enum_first_nonzero.proto
  - internal/easypc/testdata/semantic/proto3_extension_range.proto
  - internal/easypc/testdata/semantic/proto3_group.proto
  - internal/easypc/testdata/semantic/proto3_required.proto
  - internal/easypc/testdata/semantic/reserved_name_used.proto
  - internal/easypc/testdata/semantic/reserved_number_used.proto
  - internal/easypc/testdata/semantic/unknown_field_option.proto
  - internal/easypc/testdata/semantic/unknown_file_option.proto
  - internal/easypc/testdata/semantic/unknown_message_option.proto
  - internal/easypc/testdata/semantic/wrong_option_type.proto
  - internal/protoast/build.go
  - internal/protoast/comments.go
  - internal/protoast/outline.go
  - internal/protoast/position_test.go
  - internal/protoast/protoast.go
  - internal/protoast/protoast_test.go
  - internal/protoast/walk.go
  - internal/rules/casing.go
  - internal/rules/casing_test.go
  - internal/rules/lint_helpers.go
covered_digest: "v4:sha256:4d23499cb7273cba3d710338ac89d1a9cd05bf1d7fc7068aa7e6d0b86241fac1"
---

# Phase 1: Вендоринг компилятора и расчленение модели — Verification Report

**Phase Goal:** Компилятор `easypc` и модель `protoast` живут в репозитории как чистое добавление без единой новой зависимости; lint-путь больше не делит модель чтения с breaking, и это видно как пустой дифф поведения
**Verified:** 2026-11-18T22:28:47Z
**Status:** human_needed (all 23 must-haves verified; 3 judgment-tier prohibitions need human ratification — no silent pass)
**Re-verification:** No — initial verification

**MVP-mode note:** ROADMAP declares `Mode: mvp`, but the phase goal is not a User Story
(`gsd_run query user-story.validate` → `valid: false`; all four plans documented this at
planning time: "MVP story framing intentionally not invented"). MVP-mode User Flow Coverage
cannot apply; standard goal-backward verification was performed. If MVP-format verification
is wanted, run `/gsd:mvp-phase 1` to set a proper User Story goal.

## Goal Achievement

### Observable Truths

Roadmap SC rows are the contract; plan truths add detail. Score counts unique truths.

| #  | Truth | Status | Evidence |
|----|-------|--------|----------|
| 1  | SC-1: internal/easypc (целиком, incl. cmd/ + testdata/) and internal/protoast vendored with imports on github.com/easyp-tech/easyp/internal/...; go build green; vendored packages' own tests pass | ✓ VERIFIED | Tree: 48 .go files, pkg/{8 packages}, protoc(+syntax,wkt), cmd/{easyp-protoc,lexbench}, testdata/ 579 files 2.6M. `diff -r /tmp/compiler-lab-p1-04/easyp-cli/internal/easypc internal/easypc` → only VENDORED.md extra (clone at exactly 08cef10909c0722f7042ba17b5c11951f9c6b6f6). Wrong-prefix imports: 0. `go build ./...` exit 0. Vendored suites: 8 packages ok (compiler/lexer/linker/options/parser/serialize/protoc/protoast) |
| 2  | SC-2: go mod tidy leaves go.mod/go.sum unchanged — no new external module, no replace | ✓ VERIFIED | `go mod tidy && git diff --exit-code go.mod go.sum` → exit 0 (re-run this session); `grep -c '^replace' go.mod` → 0; no compiler-lab/easypc entries in go.mod |
| 3  | SC-3: task test / task lint green after vendoring + split; breaking/generate/mod tidy/migrate behavior unchanged for users | ✓ VERIFIED | `bin/golangci-lint run ./...` → 0 issues (exit 0, re-run this session). Full `go test ./...` green and `-race` 40 packages green (orchestrator fresh run); targeted `-race` runs of core/rules/protoast green. Single red spot: pre-existing wazero `GetWazeroVersion` data race in internal/adapters/plugin under `-race` — reproduced this session in the named third-party symbol; attributed pre-phase by pristine-HEAD reproduction with vendored trees removed (deferred-items.md, WINDOWS ledger id 1). Not a phase regression; all phase-touched packages green under -race |
| 4  | SC-4: multi-byte UTF-8 fixture proves protoast positions reproduce main's output contract (1-based line, 1-based rune column) before any rule porting | ✓ VERIFIED | Fixture internal/core/testdata/position_contract_utf8.proto: 19 lines, 4 tab lines, 9 multi-byte lines. D-07a TestPositionRuneColumns green; D-07b TestProtoastPositionsMatchLegacyContract green (10 anchors). No rule ported yet: helpers compiled-but-unwired (grep of helper symbols outside lint_helpers.go → 0) |
| 5  | 01-01: vendored tests run as normal packages, testdata/ retained | ✓ VERIFIED | `go test -count=1 ./internal/easypc/... ./internal/protoast/...` → 8 ok, 0 FAIL (re-run); testdata/ 579 files present |
| 6  | 01-01: VENDORED.md records upstream repo, branch main + easyp-cli/ dir, SHA 08cef10…, date, re-sync procedure | ✓ VERIFIED | internal/easypc/VENDORED.md lines 6–36: repo easyp-tech/compiler-lab (private), branch main with easyp-cli/ dir resolution, pinned SHA, vendored date 2026-11-17, import-path note, 5-step re-sync procedure incl. the four gates |
| 7  | 01-01: AGENTS.md Do-not section forbids edits inside internal/easypc with the exhausted-rewrite exemption | ✓ VERIFIED | AGENTS.md line 92, inside "## Do not" (line 86); cross-references internal/easypc/VENDORED.md |
| 8  | 01-01: golangci findings restricted to the known protoast QF1002 at the 01-01 snapshot; easypc excluded via .golangci.yml paths (not code edits) | ✓ VERIFIED | Superseded-and-strengthened: 01-02 fixed the QF1002 in code; final state `bin/golangci-lint run ./...` → 0 issues with internal/easypc excluded via `linters.exclusions.paths` only |
| 9  | 01-02: positions follow main's contract (1-based line, 1-based rune column, tab = +1) proven by D-07a on a multi-byte + tab fixture | ✓ VERIFIED | Behavioral test green this session: `go test -count=1 ./internal/protoast/ -run TestPosition` → ok; hand-computed exact-value assertions in internal/protoast/position_test.go over the UTF-8+tab fixture |
| 10 | 01-02: D-07b golden test proves anchor-by-anchor Line+Column parity between legacy go-protoparser read path and protoast on the same fixture | ✓ VERIFIED | Behavioral test green this session: `go test -count=1 ./internal/core/ -run TestProtoastPositions` → ok; 10 first-token anchors; expected values come from the legacy parser (independent oracle), zero position literals |
| 11 | 01-02: vendored protoast suite passes with the two tab-8 expectations updated to tab-as-1 | ✓ VERIFIED | diff vs upstream pinned SHA shows exactly 2 value changes in protoast_test.go (Pos.Column 9→2, NumberPos.Column 20→13) with rune-convention wording; suite green |
| 12 | 01-02: internal/protoast lint-clean — QF1002 fixed in code, protoast NOT excluded in .golangci.yml | ✓ VERIFIED | build.go:242 is now a tagged switch on n.TypeKind (staticcheck-clean); `grep -c 'internal/protoast' .golangci.yml` → 0; golangci 0 issues |
| 13 | 01-02: divergence from upstream recorded in VENDORED.md with re-apply instruction | ✓ VERIFIED | VENDORED.md "## protoast divergence (re-apply on every sync)" section names pos() rune columns, tab-as-1 expectations, package-doc wording, and mandates re-apply on every sync |
| 14 | 01-03: core.Issue.Position is protoast.Position (alias in dom.go) with conversion at exactly two funnels; 48 rule call sites and ~20 breaking call sites unchanged | ✓ VERIFIED | dom.go:20 `type Position = protoast.Position`; fromMetaPos at dom.go:169 field-for-field; AppendIssue keeps `pos meta.Position` param (dom.go:127) and converts once (dom.go:129); buildBreakingCheckIssue keeps meta.Position param and converts once (breaking_checker.go:370); meta.Position{ literals in production exist only at the two planned rule call sites + adapted lint_helpers call |
| 15 | 01-03: every expected-position literal keeps numeric values, type-name-only swap (empty diff-of-diffs) | ✓ VERIFIED | 56/56 protoast.Position{ in tests, 0 meta.Position{ in tests; values gate on plan diff 6563fab..0d7d8cb: multiset of removed == added numeric Line/Column/Offset lines (both empty — no numeric field line changed at all) |
| 16 | 01-03: lint and breaking no longer share a read-model type (D-06) | ✓ VERIFIED | bare protoInfoRead = 0; lintProtoInfoRead ×7 (def+caller+label+…), breakingProtoInfoRead ×7; lint owns ProtoInfo via internal/core/lint_proto_info.go; breaking solely owns BreakingProtoInfo/breakingProtoInfoRead in proto_info_read.go; twin-mirror tests green this session (TestProtoInfoReadTwinsMirroring + both error-label suites PASS) |
| 17 | 01-04: casing.go + casing_test.go byte-identical to compiler-lab main@08cef10 easyp-cli/internal/rules/, tests pass | ✓ VERIFIED | `diff /tmp/compiler-lab-p1-04/easyp-cli/internal/rules/casing.go internal/rules/casing.go` → identical (same for casing_test.go); TestCasingMatchesBuf PASS this session |
| 18 | 01-04: lint_helpers.go verbatim except the 3 model-coupled declarations, delta enumerated in header | ✓ VERIFIED | `diff -u` vs upstream: exactly 3 hunks — header+import block, packagePosition signature (file *protoast.File, path string), mismatches AppendIssue 5-arg call with field-for-field meta.Position conversion; zero changed lines in any other function; header carries pinned SHA + D-01 citation + sync pointer |
| 19 | 01-04: helpers allowed temporarily unused; D-02 fallback (if fired) scoped to lint_helpers.go only | ✓ VERIFIED | .golangci.yml paths: [internal/easypc, internal/rules/lint_helpers.go] — fallback fired on actual U1000 failure and is scoped to the single file; casing.go/casing_test.go NOT path-excluded |
| 20 | 01-04: task-test-equivalent and golangci stay green after the helper drop | ✓ VERIFIED | `go test -race -count=1 ./internal/rules/...` ok; golangci 0 issues; full-suite state as in row 3 |
| 21 | 01-04: tidy zero-diff holds (helpers import only stdlib + protoast + core) | ✓ VERIFIED | `go mod tidy && git diff --exit-code go.mod go.sum` → exit 0 (re-run this session); lint_helpers.go imports: strconv, strings, parser/meta, core, protoast — all in-module or existing |
| 22 | 01-01 prohibition P2 (check): no CLI command executes easypc/protoast parsing at runtime — type-level only | ✓ VERIFIED | Import graph: internal/easypc imported only by itself + internal/protoast; internal/protoast imported only by internal/core (type alias dom.go, one literal breaking_profiles.go, tests) and internal/rules (unwired lint_helpers.go, test literals); internal/api references: 0; lint.go parses via readProtoFileWithDirectives (go-protoparser) |
| 23 | 01-03 prohibition P3 (check): no new Options knobs or buildCore changes; split passes through the existing runtime seam | ✓ VERIFIED | `grep -rn 'lintProtoInfoRead\|breakingProtoInfoRead' internal/api/` → 0; plan diff 6563fab..0d7d8cb touches no file outside internal/core and internal/rules |

**Score:** 23/23 truths verified (0 present, behavior-unverified)

### Deferred Items

None. The pre-existing wazero race (deferred-items.md, WINDOWS ledger id 1) is not a failed
truth of this phase — it predates the phase (pristine-HEAD reproduction recorded), lives in
third-party wazero code, and is already tracked in the deferred-items ledger. No later
milestone phase claims it; it is an environmental note, not a phase gap.

### Advisory (New Scope, Unevidenced)

Not applicable — initial verification (no prior gap-closure contract to bound).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| internal/easypc/VENDORED.md | Provenance + re-sync procedure | ✓ VERIFIED | 58 lines (min 12); all D-03 fields present + protoast divergence section |
| internal/easypc/pkg/parser | Vendored parser package | ✓ VERIFIED | Directory artifact (tool's file-check quirk noted); ast.go/parser.go/parser_test.go etc. present, package tests ok; byte-identical to upstream per tree diff |
| internal/protoast/protoast.go | Position struct + entry points | ✓ VERIFIED | Present; Position{Filename,Offset,Line,Column}; Read entry used by parity test; rune-contract package doc |
| .golangci.yml | linters.exclusions.paths scoping internal/easypc | ✓ VERIFIED | paths: [internal/easypc, internal/rules/lint_helpers.go]; protoast not excluded |
| AGENTS.md | D-04 prohibition bullet | ✓ VERIFIED | Line 92 in "## Do not" |
| internal/core/testdata/position_contract_utf8.proto | UTF-8+tab fixture | ✓ VERIFIED | 19 lines (min 15); 4 tab lines, 9 multi-byte lines |
| internal/protoast/position_test.go | D-07a unit proof | ✓ VERIFIED | Table-driven exact hand-computed Line/Column values |
| internal/core/position_parity_test.go | D-07b golden proof | ✓ VERIFIED | 10 anchors, package core, uses readProtoFileWithDirectives + protoast.Read |
| internal/protoast/build.go | pos() rune adaptation + QF1002 fixed | ✓ VERIFIED | pos() at :453 with utf8.RuneCount+1 column and contract doc; tagged switch at :242 |
| internal/core/dom.go | Position alias + fromMetaPos | ✓ VERIFIED | Alias :20, fromMetaPos :169, AppendIssue meta.Position param preserved |
| internal/core/lint_proto_info.go | lint-owned read constructor | ✓ VERIFIED | 33 lines (min 20); same error-wrap labels; go-protoparser path |
| internal/core/proto_info_read.go | BreakingProtoInfo + breakingProtoInfoRead, shared helpers unchanged | ✓ VERIFIED | Type :48, constructor :20; readFilesFromImport/readFileFromImport/openImportFile/sourceImportPath/openDiskSource preserved |
| internal/rules/casing.go | Vendored casing helpers | ✓ VERIFIED | Byte-identical to upstream pinned SHA |
| internal/rules/casing_test.go | Vendored colocated tests | ✓ VERIFIED | Byte-identical; TestCasingMatchesBuf green |
| internal/rules/lint_helpers.go | Helper set with 3 seam adaptations + delta header | ✓ VERIFIED | 196 lines (min 40); diff vs upstream = exactly 3 hunks + header |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| .golangci.yml | internal/easypc | linters.exclusions.paths entry (D-04) | WIRED | Tool-verified + manual read |
| internal/easypc/VENDORED.md | AGENTS.md | edit-prohibition cross-reference | WIRED | Both directions reference each other |
| internal/protoast/build.go | internal/protoast/protoast.go | pos() constructs Position; namePos/attachComments route through it | WIRED | pos() :453, namePos :470 routes through pos() |
| internal/core/position_parity_test.go | internal/core/proto_comments.go | legacy side calls readProtoFileWithDirectives | WIRED | parity_test.go:27 |
| internal/core/position_parity_test.go | internal/protoast/protoast.go | protoast side parses fixture via public read entry | WIRED | Tool false-negative: protoast.Read at parity_test.go:29 (manual check) |
| internal/core/dom.go | internal/protoast/protoast.go | Position alias + fromMetaPos | WIRED | dom.go:20,169 |
| internal/core/lint.go | internal/core/lint_proto_info.go | c.lintProtoInfoRead call + label | WIRED | lint.go:34,36 |
| internal/core/breaking_check.go | internal/core/proto_info_read.go | c.breakingProtoInfoRead + collect([]BreakingProtoInfo) | WIRED | breaking_check.go:119,121 |
| internal/core/position_parity_test.go | proto_comments.go | shared helper untouched (parallel-wave safety) | WIRED | Helper intact at original location |
| internal/rules/lint_helpers.go | internal/core/dom.go | core.AppendIssue 5-arg adapted call | WIRED | lint_helpers.go:140 |
| internal/rules/casing_test.go | internal/rules/casing.go | colocated tests exercise casing helpers | WIRED | Tool false-negative: test calls toPascalCase/toLowerSnakeCase/toUpperSnakeCase unqualified in-package (manual check) |

### Data-Flow Trace (Level 4)

Not a UI/rendering phase; the dynamic-value paths are the read models.

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| internal/core/lint.go | protoInfo | c.lintProtoInfoRead → readProtoFileWithDirectives (go-protoparser parse) | Yes — real parse of walked files | ✓ FLOWING |
| internal/core/breaking_check.go | protoInfo | c.breakingProtoInfoRead → readProtoFileWithDirectives | Yes — real parse | ✓ FLOWING |
| internal/protoast positions | Position values | builder.pos() over b.ast.Src + b.syn.LineCol | Yes — computed from parsed source; proven by D-07a/b tests | ✓ FLOWING |

No static fallbacks, mocks, or hardcoded renders in any phase-touched path.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Module builds | go build ./... | exit 0 | ✓ PASS |
| D-07a position contract | go test -count=1 ./internal/protoast/ -run 'TestPosition' | ok | ✓ PASS |
| D-07b golden parity | go test -count=1 ./internal/core/ -run 'TestProtoastPositions' | ok | ✓ PASS |
| Vendored casing tests | go test -count=1 ./internal/rules/ -run 'TestCasingMatchesBuf' -v | PASS | ✓ PASS |
| Twin-mirror read models | go test -count=1 ./internal/core/ -run 'TestProtoInfoReadTwinsMirroring\|TestLintProtoInfoReadErrorLabels\|TestBreakingProtoInfoReadErrorLabels' -v | 3/3 PASS | ✓ PASS |
| Vendored packages' own tests | go test -count=1 ./internal/easypc/... ./internal/protoast/... | 8 ok, 0 FAIL | ✓ PASS |
| Lint gate | bin/golangci-lint run ./... | 0 issues, exit 0 | ✓ PASS |
| Tidy zero-diff | go mod tidy && git diff --exit-code go.mod go.sum | exit 0 | ✓ PASS |
| Values-unchanged (plan diff) | diff of numeric Line/Column/Offset multisets over 6563fab..0d7d8cb | empty, exit 0 | ✓ PASS |
| Pre-existing race attribution | go test -race -count=1 ./internal/adapters/plugin/ | DATA RACE in wazero GetWazeroVersion — matches deferred-items.md exactly | ✓ CONFIRMED (pre-existing, documented) |

### Probe Execution

SKIPPED — no probes declared in any plan/SUMMARY and none exist conventionally
(`find scripts -path '*/tests/probe-*.sh'` → empty).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-VEND-01 | 01-01, 01-02, 01-03, 01-04 | easypc (целиком, incl. cmd/ + testdata/) + protoast вкопированы из compiler-lab/easyp-cli с импортами на github.com/easyp-tech/easyp/internal/... | ✓ SATISFIED | Verbatim tree diffs vs pinned SHA 08cef10; imports all on easyp module (wrong-prefix count 0); helpers vendored per D-01 refinement (casing verbatim, lint_helpers with 3 enumerated adaptations); REQUIREMENTS.md traceability already marks Phase 1 Complete for both IDs |
| LINT-VEND-02 | 01-01 | go mod tidy → без новых внешних зависимостей (без внешнего модуля и без replace) | ✓ SATISFIED | 0 replace directives; tidy zero-diff re-proven this session; go-protoparser v4.14.2 and protocompile v0.14.1 retained; go-billy marker-only indirect→direct promotion (same module, same version — no new dependency; documented deviation, see human verification item 3) |

**Orphaned requirements:** none — REQUIREMENTS.md maps exactly {LINT-VEND-01, LINT-VEND-02} to Phase 1 and both are claimed by plans.

**Decision coverage:** 10/10 trackable CONTEXT.md decisions honored by shipped artifacts
(`gsd_run query check.decision-coverage-verify` → `{total: 10, honored: 10, not_honored: []}`).

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --------- | ---------- | ------ | ------- | -------- | --------------- | ------- |
| internal/protoast/position_test.go | LINT-VEND-01 (D-07a) | yes | 0 | no | Value (exact hand-computed Line/Column) | SUFFICIENT |
| internal/core/position_parity_test.go | LINT-VEND-01 (D-07b) | yes | 0 | no | Value vs independent oracle (legacy go-protoparser — the contract's definition) | SUFFICIENT — provenance VALID |
| internal/rules/casing_test.go | LINT-VEND-01 (D-01) | yes | 0 | no | Value (vendored buf-compat expectations) | SUFFICIENT |
| internal/core/proto_info_read_test.go (twin tests) | LINT-VEND-01 (D-06) | yes | 0 | no | Behavioral (field-by-field equality + error-label shapes) | SUFFICIENT — twin equality is the declared Phase-1 contract, divergence-catching by design |

Disabled tests on requirements: 0. Circular patterns: 0 (the os.WriteFile in
proto_info_read_test.go writes input fixtures, not expected values).
Note: review finding WR-03 (parity test asserts Line+Column but not Offset/Filename) is
recorded in the phase review system (01-REVIEW.md, disposition open); the truth as written
says "Line+Column equality", which the test proves. Offset parity remains relevant for
Phase 4 porting and is already tracked as WR-03.

### Prohibitions (ADR-550 D4 — judgment-tier, autonomous verdicts NON-AUTHORITATIVE)

All 12 flagged prohibitions checked. 9 have deterministic evidence of being honored:

| Prohibition | Verdict (LLM judge) | Evidence |
| ----------- | -------------------- | -------- |
| 01-02 P1: no Issue post-processing / per-rule shims; contract lives in pos() exactly once | honored | Conversion only at pos() + fromMetaPos two funnels; no rule file contains conversion logic |
| 01-02 P2: no easypc edit to fix positions | honored | git log internal/easypc code files: only vendor commit 690a9cc |
| 01-02 P3: protoast NOT excluded in .golangci.yml | honored | Config read: protoast absent from paths |
| 01-03 P1: no numeric Line/Column/Offset/Filename value changes | honored | Values gate empty on plan diff |
| 01-03 P2: no protoast/easypc in runtime parse paths | honored | lint_proto_info.go + proto_info_read.go parse via readProtoFileWithDirectives (go-protoparser); import graph shows no easypc outside itself+protoast |
| 01-03 P3: no new Options knobs / buildCore changes | honored | internal/api references 0; diff scope check clean |
| 01-04 P1: casing.go/casing_test.go not adapted, trimmed, or subset | honored | Byte-identical diff vs pinned upstream |
| 01-04 P3: no helper wired into a live rule | honored | Helper symbols referenced nowhere outside lint_helpers.go (production or tests) |
| 01-01 P2: no CLI runtime execution of easypc/protoast | honored | Import graph: internal/api clean; lint.go parses via go-protoparser |

3 need human ratification (letter-level conflicts with deterministic evidence of preserved
intent) — listed under Human Verification Required. **unverified-prohibition — human review
recommended.** None of the three is evidence of a behavior change; each is a documented
letter-vs-intent interpretation the executors themselves flagged.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | Debt markers TBD/FIXME/XXX/TODO/HACK in all phase-authored files | — | 0 occurrences (also 0 in the vendored tree) |
| (none) | — | Stub/placeholder/empty-implementation patterns | — | lint_helpers.go "compiled-but-unwired" is the plan's own declared D-01 intent, documented in Known Stubs — not a functional stub |

Phase 1 review findings WR-01..WR-03 / IN-01..IN-07 are recorded with open dispositions in
01-REVIEW.md / 01-REVIEW-DISPOSITION.md and are tracked by the review process; none rises to
a verification blocker (all suites green, no must-have affected).

### Human Verification Required

This is an infrastructure/foundation phase (vendoring + type-level refactor; no user-facing
UX), so no UAT-style manual steps are invented. The items below are Escalation-Gate decisions
on flagged judgment-tier prohibitions — they need a human ruling, not a test run:

### 1. Ratify the QF1001 finding-scoped exclusion on casing.go (executor-requested)

**Test:** Decide whether `linters.exclusions.rules: {staticcheck QF1001, paths: internal/rules/casing\.go}` in .golangci.yml (lines 27-29) is an acceptable resolution of the constraint triad (D-01 byte-identity + "casing files are never excluded" + green lint gate) given upstream casing.go:49 carries QF1001 and the lab's own config tolerates it.
**Expected:** A ratification decision (or an alternative: e.g. accept a red lint gate for verbatim files, or get the fix upstream and re-sync). The current scope silences only QF1001 for that one file; unused/errcheck/govet and all other staticcheck codes stay active on casing.go.
**Why human:** The prohibition's letter ("casing files are never excluded") conflicts with its intent (no masking of casing lint coverage) under an unplanned upstream finding; plan 01-04's coverage row D4 is explicitly `human_judgment: true` and asks for ratification before the pattern spreads to other verbatim files.

### 2. Acknowledge the VENDORED.md edit under the internal/easypc seal (01-01 P1 letter)

**Test:** Confirm that editing internal/easypc/VENDORED.md (commit 31885b4, planned explicitly by 01-02's files_modified) does not violate the "no file under internal/easypc may be modified" prohibition, whose D-04 intent targets the vendored upstream code.
**Expected:** Acknowledgment that provenance maintenance is the intended exception; vendored code files have exactly one touching commit (690a9cc).
**Why human:** The prohibition as written covers all files under internal/easypc including the provenance doc; the plan set resolves the contradiction explicitly but the letter remains over-broad.

### 3. Acknowledge the go-billy indirect→direct marker move against the byte-identical tidy gate (01-01 T3 letter, REQ-1.2)

**Test:** Confirm that commit 8f969d8 (go-billy v5.6.2 promoted indirect→direct in go.mod, landed as a dedicated pre-vendor commit so the vendor commit stays pure) satisfies LINT-VEND-02's substance: no new module, no replace, no new require entry, go.sum untouched.
**Expected:** Acknowledgment that the byte-identical expectation was falsified by a pre-existing stale go.mod condition (attribution: identical tidy diff on pristine HEAD with vendor trees removed), not by the vendoring; the zero-new-dependencies substance holds.
**Why human:** The plan's flagged assumption ("tidy leaves go.mod/go.sum byte-identical") was literally false at HEAD; the executor documented the deviation transparently and requested verifier/human attention on REQ-1.2 attribution.

### Gaps Summary

No gaps. All 23 must-haves verified with codebase and command evidence; all 10 claimed
commits exist; requirements LINT-VEND-01/LINT-VEND-02 satisfied; decision coverage 10/10;
zero debt markers; zero disabled/circular tests.

The phase goal is achieved in the codebase:

1. **Чистое добавление без единой новой зависимости** — proven at the strongest possible
   level: the vendored easypc tree is byte-identical to compiler-lab main@08cef10 (verified
   against a clone pinned at exactly that SHA), protoast differs only in the documented
   rune-column adaptation, and go.mod/go.sum are tidy-zero-diff with zero replace directives.
2. **Lint-путь больше не делит модель чтения с breaking** — lint owns ProtoInfo +
   lintProtoInfoRead; breaking solely owns BreakingProtoInfo + breakingProtoInfoRead; the
   shared bare name is gone; twin-mirror tests pin the Phase-1 behavioral identity.
3. **Пустой дифф поведения** — machine-checked: 56 test literals changed type name only
   (numeric-values gate empty over the whole plan diff), full suites green, lint 0 issues.

Status is human_needed solely because three flagged judgment-tier prohibitions require
human ratification (see above) — a silent pass is not permitted for flagged prohibitions.
Awaiting human decision; no code changes are required by this verification.

---

_Verified: 2026-11-18T22:28:47Z_
_Verifier: Claude (gsd-verifier)_

## Acknowledged Gaps

Acknowledged during /gsd-verify-work 1 artifact scan (2026-11-19), user confirmed proceed:

- deferred-items.md — "Pre-existing wazero data race fails `internal/adapters/plugin` under `-race` on this machine (go1.27.3)" — open; out of phase scope (reproduced on pristine HEAD before the vendor drop, upstream third-party code), dispositioned as FUTURE.
