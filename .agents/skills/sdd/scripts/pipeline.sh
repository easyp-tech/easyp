#!/usr/bin/env sh
# Spec-Driven Dev Pipeline — state machine (POSIX sh, zero dependencies)
# Usage: sh pipeline.sh [--feature <name>] <command> [args]
#
# Shell compatibility: requires sh with `local` support (bash, dash, ash, zsh).
#
# Global flags:
#   --feature <name>      Specify which feature to operate on (required when
#                         multiple pipelines are active simultaneously)
#
# Commands:
#   init [--branch|--no-branch] <feature-name>
#                         Start a new pipeline for a feature
#                         --branch: create git branch <prefix><name> (prefix from config, default: feature/)
#                         --no-branch: skip branch creation even if auto_branch is set in config
#   status                Show current phase, feature, and artifacts
#   approve               Advance to next phase (requires artifact)
#   artifact [path]       Register artifact for current phase
#   history               Show all features and their status
#   revisions [phase]     Show revision history for current or specified phase
#   docs-check            Check project documentation status
#   task <T-N>            Mark implementation task as completed (resume tracking)
#   version               Show version
#   help                  Show this help message

set -e

PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FEATURES_DIR="$PROJECT_ROOT/.spec/features"
CONFIG_FILE="$PROJECT_ROOT/.spec/config.yaml"

# --- helpers ---

VERSION="1.5.0"
EXPLICIT_FEATURE=""

die() { echo "ERROR: $*" >&2; exit 1; }
info() { echo "→ $*"; }
warn() { echo "⚠ $*" >&2; }

iso_now() {
  date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ"
}

iso_now_compact() {
  date -u +"%Y-%m-%dT%H-%M-%SZ" 2>/dev/null || date +"%Y-%m-%dT%H-%M-%SZ"
}

# Escape a string for safe embedding in JSON values (RFC 8259)
json_escape() {
  printf '%s' "$1" | awk '
    BEGIN { ORS="" }
    {
      gsub(/\\/, "\\\\")
      gsub(/"/, "\\\"")
      gsub(/\t/, "\\t")
      gsub(/\r/, "\\r")
      if (NR > 1) printf "\\n"
      printf "%s", $0
    }'
}

# Read a value from .spec/config.yaml (simple grep-based, no YAML parser)
# Usage: read_config <key> [default]
# Returns the value or default (empty string if no default)
read_config() {
  local key="$1" default="${2:-}"
  if [ -f "$CONFIG_FILE" ]; then
    local val
    val="$(grep "^${key}:" "$CONFIG_FILE" 2>/dev/null | head -1 | sed "s/^${key}:[[:space:]]*//" | sed 's/[[:space:]]*$//')"
    if [ -n "$val" ]; then
      printf '%s' "$val"
      return
    fi
  fi
  printf '%s' "$default"
}

# --- per-feature state ---

# Current feature paths (set by set_feature_context / resolve_feature)
FEATURE_DIR=""
STATE_FILE=""
KV_FILE=""
REVISIONS_DIR=""
APPROVED_DIR=""

set_feature_context() {
  # set_feature_context <feature-name> — sets global paths for the feature
  FEATURE_DIR="$FEATURES_DIR/$1"
  KV_FILE="$FEATURE_DIR/pipeline.kv"
  STATE_FILE="$FEATURE_DIR/pipeline.json"
  REVISIONS_DIR="$FEATURE_DIR/revisions"
  APPROVED_DIR="$FEATURE_DIR/approved"
}

ensure_feature_dir() {
  # ensure_feature_dir <feature-name> — creates feature directory structure
  local fdir="$FEATURES_DIR/$1"
  mkdir -p "$fdir" "$fdir/revisions" "$fdir/approved"
}

read_field() {
  [ -f "$KV_FILE" ] || return 1
  local _line
  _line="$(grep "^$1=" "$KV_FILE" 2>/dev/null | head -1)" || return 1
  [ -n "$_line" ] || return 1
  printf '%s' "$_line" | cut -d'=' -f2-
}

validate_kv() {
  # Verify required fields exist in KV store; die with diagnostic on failure
  [ -f "$KV_FILE" ] || die "Pipeline state file missing: $KV_FILE"
  local missing=""
  for field in feature phase created_at; do
    grep -q "^${field}=" "$KV_FILE" 2>/dev/null || missing="$missing $field"
  done
  if [ -n "$missing" ]; then
    die "Corrupted pipeline state ($KV_FILE): missing fields:$missing. Fix the file manually or remove and re-init."
  fi
  # Verify every line matches key=value format (key: lowercase + digits + underscore)
  local line_num=0
  while IFS= read -r line || [ -n "$line" ]; do
    line_num=$((line_num + 1))
    case "$line" in
      "") continue ;;  # skip blank lines
      [a-z_]*=*) ;;    # valid key=value
      *) die "Corrupted pipeline state ($KV_FILE): invalid line $line_num: $line" ;;
    esac
  done < "$KV_FILE"
}

# Escape a value for safe use in sed replacement string
kv_escape_sed() {
  printf '%s' "$1" | sed -e 's/[&\\/|]/\\&/g'
}

# Validate that a value is safe for the KV store (no =, |, or newlines)
kv_validate_value() {
  case "$1" in
    *'='*) die "KV value must not contain '=': $1" ;;
    *'|'*) die "KV value must not contain '|': $1" ;;
  esac
  # Check for newlines by comparing line count (portable across POSIX shells)
  local line_count
  line_count="$(printf '%s' "$1" | wc -l)"
  if [ "$line_count" -ne 0 ]; then
    die "KV value must not contain newlines: $1"
  fi
}

# Validate artifact path: reject directory traversal and control characters
validate_artifact_path() {
  case "$1" in
    */../*|*/..) die "Artifact path must not contain '..' traversal" ;;
    ../*|..)     die "Artifact path must not contain '..' traversal" ;;
  esac
  if printf '%s' "$1" | grep -q '[[:cntrl:]]' 2>/dev/null; then
    die "Artifact path must not contain control characters"
  fi
}

write_field() {
  kv_validate_value "$2"
  if [ -f "$KV_FILE" ] && grep -q "^$1=" "$KV_FILE" 2>/dev/null; then
    local tmp="$KV_FILE.tmp"
    local escaped
    escaped="$(kv_escape_sed "$2")"
    sed "s|^$1=.*|$1=$escaped|" "$KV_FILE" > "$tmp" && mv "$tmp" "$KV_FILE"
  else
    echo "$1=$2" >> "$KV_FILE"
  fi
}

detect_active_feature() {
  # Scan all features, return the one with phase != done
  [ -d "$FEATURES_DIR" ] || return 0
  local active=""
  local count=0
  for kv in "$FEATURES_DIR"/*/pipeline.kv; do
    [ -f "$kv" ] || continue
    local phase
    phase="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
    if [ -n "$phase" ] && [ "$phase" != "done" ]; then
      local fname
      fname="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
      active="$fname"
      count=$((count + 1))
    fi
  done
  if [ "$count" -gt 1 ]; then
    warn "Multiple active pipelines found:"
    for kv in "$FEATURES_DIR"/*/pipeline.kv; do
      [ -f "$kv" ] || continue
      local phase fname
      phase="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
      fname="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
      if [ -n "$phase" ] && [ "$phase" != "done" ]; then
        echo "  - $fname (phase: $phase)" >&2
      fi
    done
    return 1
  fi
  [ -n "$active" ] && echo "$active"
}

resolve_feature() {
  if [ -n "$EXPLICIT_FEATURE" ]; then
    # Validate that the explicitly specified feature exists
    if [ ! -f "$FEATURES_DIR/$EXPLICIT_FEATURE/pipeline.kv" ]; then
      die "Feature '$EXPLICIT_FEATURE' not found. Run 'pipeline.sh history' to list features."
    fi
    set_feature_context "$EXPLICIT_FEATURE"
    validate_kv
    return 0
  fi
  local feat
  feat="$(detect_active_feature)" || { warn "Hint: use --feature <name> to select one."; return 1; }
  if [ -z "$feat" ]; then
    return 1
  fi
  set_feature_context "$feat"
  validate_kv
  return 0
}

next_phase() {
  case "$1" in
    explore)        echo "requirements" ;;
    requirements)   echo "design" ;;
    design)         echo "task-plan" ;;
    task-plan)      echo "implementation" ;;
    implementation) echo "review" ;;
    review)         echo "done" ;;
    done)           echo "" ;;
    *)              echo "" ;;
  esac
}

phase_number() {
  case "$1" in
    explore)        echo "1" ;;
    requirements)   echo "2" ;;
    design)         echo "3" ;;
    task-plan)      echo "4" ;;
    implementation) echo "5" ;;
    review)         echo "6" ;;
    done)           echo "✓" ;;
    *)              echo "?" ;;
  esac
}

# Numeric ordering for comparisons (phase_number is for display)
phase_order() {
  case "$1" in
    explore)        echo 1 ;;
    requirements)   echo 2 ;;
    design)         echo 3 ;;
    task-plan)      echo 4 ;;
    implementation) echo 5 ;;
    review)         echo 6 ;;
    done)           echo 7 ;;
    *)              echo 0 ;;
  esac
}

# Rebuild the JSON file from KV store (for agents to read)
# Uses atomic write (tmp + mv) to prevent corruption on interruption
rebuild_json() {
  validate_kv
  local feature phase created artifact
  feature="$(json_escape "$(read_field feature)")"
  phase="$(read_field phase)"
  created="$(read_field created_at)"
  artifact="$(read_field current_artifact)"
  local history_count
  history_count="$(read_field history_count)"
  [ -z "$history_count" ] && history_count=0

  local tmp_file="$STATE_FILE.tmp"
  {
    printf '{\n'
    printf '  "feature": "%s",\n' "$feature"
    printf '  "phase": "%s",\n' "$phase"
    printf '  "created_at": "%s",\n' "$created"
    if [ -n "$artifact" ]; then
      printf '  "current_artifact": "%s",\n' "$(json_escape "$artifact")"
    else
      printf '  "current_artifact": null,\n'
    fi
    printf '  "history": [\n'

    local i=0
    while [ "$i" -lt "$history_count" ]; do
      local h_phase h_artifact h_approved
      h_phase="$(read_field "history_${i}_phase")"
      h_artifact="$(json_escape "$(read_field "history_${i}_artifact")")"
      h_approved="$(read_field "history_${i}_approved_at")"
      [ "$i" -gt 0 ] && printf ',\n'
      printf '    {"phase": "%s", "artifact": "%s", "approved_at": "%s"}' \
        "$h_phase" "$h_artifact" "$h_approved"
      i=$((i + 1))
    done

    printf '\n  ],\n'

    # Include review_base_commit if set
    local rbc
    rbc="$(read_field review_base_commit 2>/dev/null || echo "")"
    if [ -n "$rbc" ]; then
      printf '  "review_base_commit": "%s",\n' "$(json_escape "$rbc")"
    else
      printf '  "review_base_commit": null,\n'
    fi

    # Include branch if set
    local br
    br="$(read_field branch 2>/dev/null || echo "")"
    if [ -n "$br" ]; then
      printf '  "branch": "%s",\n' "$(json_escape "$br")"
    else
      printf '  "branch": null,\n'
    fi

    # Include worktree if set
    local wt
    wt="$(read_field worktree 2>/dev/null || echo "")"
    if [ -n "$wt" ]; then
      printf '  "worktree": "%s",\n' "$(json_escape "$wt")"
    else
      printf '  "worktree": null,\n'
    fi

    # Include last_completed_task if set
    local lct
    lct="$(read_field last_completed_task 2>/dev/null || echo "")"
    if [ -n "$lct" ]; then
      printf '  "last_completed_task": "%s",\n' "$(json_escape "$lct")"
    else
      printf '  "last_completed_task": null,\n'
    fi

    # Include finish fields if set
    local fa ft fb
    fa="$(read_field finish_action 2>/dev/null || echo "")"
    ft="$(read_field finished_at 2>/dev/null || echo "")"
    fb="$(read_field finish_base 2>/dev/null || echo "")"
    if [ -n "$fa" ]; then
      printf '  "finish_action": "%s",\n' "$(json_escape "$fa")"
      printf '  "finished_at": "%s",\n' "$(json_escape "$ft")"
      if [ -n "$fb" ]; then
        printf '  "finish_base": "%s"\n' "$(json_escape "$fb")"
      else
        printf '  "finish_base": null\n'
      fi
    else
      printf '  "finish_action": null,\n'
      printf '  "finished_at": null,\n'
      printf '  "finish_base": null\n'
    fi

    printf '}\n'
  } > "$tmp_file"
  mv -f "$tmp_file" "$STATE_FILE"
}

# --- commands ---

cmd_init() {
  # Parse init-specific flags
  local do_branch=""
  local do_worktree=""
  local feature=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --branch)      do_branch="yes"; shift ;;
      --worktree)    do_worktree="yes"; shift ;;
      --no-branch)   do_branch="no"; do_worktree="no"; shift ;;
      -*)            die "Unknown flag for init: $1" ;;
      *)
        [ -n "$feature" ] && die "Unexpected argument: $1"
        feature="$1"; shift
        ;;
    esac
  done

  [ -z "$feature" ] && die "Usage: pipeline.sh init [--branch|--worktree|--no-branch] <feature-name>"

  # Mutual exclusion
  if [ "$do_branch" = "yes" ] && [ "$do_worktree" = "yes" ]; then
    die "--branch and --worktree are mutually exclusive."
  fi

  # Validate feature name (kebab-case)
  case "$feature" in
    *[!a-z0-9-]*) die "Feature name must be kebab-case (e.g. grpc-streaming-support)" ;;
    -*|*-)        die "Feature name must be kebab-case (e.g. grpc-streaming-support)" ;;
    *--*)         die "Feature name must be kebab-case (e.g. grpc-streaming-support)" ;;
    [!a-z]*)      die "Feature name must be kebab-case (e.g. grpc-streaming-support)" ;;
  esac

  if [ ${#feature} -gt 64 ]; then
    die "Feature name too long (max 64 chars): $feature"
  fi

  # Resolve branch/worktree creation: flag > config > default (neither)
  if [ -z "$do_branch" ] && [ -z "$do_worktree" ]; then
    local auto_branch auto_worktree
    auto_worktree="$(read_config auto_worktree "false")"
    case "$auto_worktree" in
      true|yes|1) do_worktree="yes" ;;
    esac
    if [ "$do_worktree" != "yes" ]; then
      auto_branch="$(read_config auto_branch "false")"
      case "$auto_branch" in
        true|yes|1) do_branch="yes" ;;
        *)          do_branch="no" ;;
      esac
    fi
  fi

  local branch_name=""
  local worktree_path=""

  if [ "$do_worktree" = "yes" ]; then
    # --- Worktree mode ---
    if ! command -v git >/dev/null 2>&1; then
      die "Git not found. Cannot create worktree."
    fi
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
      die "Not a git repository. Cannot create worktree."
    fi

    local prefix wt_dir
    prefix="$(read_config branch_prefix "feature/")"
    wt_dir="$(read_config worktree_dir ".worktrees")"
    branch_name="${prefix}${feature}"
    worktree_path="${wt_dir}/${feature}"

    # Check if branch already exists
    if git rev-parse --verify "$branch_name" >/dev/null 2>&1; then
      die "Branch '$branch_name' already exists."
    fi

    # Warn if worktree_dir is not in .gitignore
    if [ -f ".gitignore" ]; then
      if ! grep -qx "$wt_dir" .gitignore 2>/dev/null && ! grep -qx "$wt_dir/" .gitignore 2>/dev/null; then
        warn "Worktree directory '$wt_dir' is not in .gitignore. Consider adding it."
      fi
    else
      warn "No .gitignore found. Consider adding '$wt_dir' to .gitignore."
    fi

    git worktree add "$worktree_path" -b "$branch_name" || die "Failed to create worktree at '$worktree_path'."
    info "Created worktree: $worktree_path (branch: $branch_name)"

  elif [ "$do_branch" = "yes" ]; then
    # Verify git is available
    if ! command -v git >/dev/null 2>&1; then
      die "Git not found. Cannot create branch."
    fi
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
      die "Not a git repository. Cannot create branch."
    fi

    local prefix
    prefix="$(read_config branch_prefix "feature/")"
    branch_name="${prefix}${feature}"

    # Check if branch already exists
    if git rev-parse --verify "$branch_name" >/dev/null 2>&1; then
      die "Branch '$branch_name' already exists."
    fi

    # Warn about dirty working tree
    if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
      warn "Working tree has uncommitted changes."
    fi

    git checkout -b "$branch_name" || die "Failed to create branch '$branch_name'."
    info "Created branch: $branch_name"
  fi

  local fdir="$FEATURES_DIR/$feature"

  # Check if feature already exists
  if [ -f "$fdir/pipeline.kv" ]; then
    local existing_phase
    existing_phase="$(grep "^phase=" "$fdir/pipeline.kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
    if [ "$existing_phase" = "done" ]; then
      die "Feature '$feature' already completed. Choose a different name."
    else
      warn "Active pipeline for '$feature' exists (phase: $existing_phase)"
      die "Complete or choose a different feature name."
    fi
  fi

  ensure_feature_dir "$feature"
  set_feature_context "$feature"

  # Initialize KV store
  {
    echo "feature=$feature"
    echo "phase=explore"
    echo "created_at=$(iso_now)"
    echo "current_artifact="
    echo "history_count=0"
    if [ -n "$branch_name" ]; then
      echo "branch=$branch_name"
    fi
    if [ -n "$worktree_path" ]; then
      echo "worktree=$worktree_path"
    fi
  } > "$KV_FILE"

  rebuild_json
  info "Pipeline initialized for '$feature'"
  if [ -n "$worktree_path" ]; then
    info "Worktree: $worktree_path (branch: $branch_name)"
  elif [ -n "$branch_name" ]; then
    info "Branch: $branch_name"
  fi
  info "Phase: [1/6] explore"
  info "Artifacts: .spec/features/$feature/"
  info "Read template: ./templates/explore.md"
}

cmd_status() {
  if ! resolve_feature; then
    info "No active pipeline."
    # Show completed features if any
    if [ -d "$FEATURES_DIR" ]; then
      local has_completed=0
      for kv in "$FEATURES_DIR"/*/pipeline.kv; do
        [ -f "$kv" ] || continue
        local phase
        phase="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        if [ "$phase" = "done" ]; then
          if [ "$has_completed" -eq 0 ]; then
            echo ""
            echo "Completed features:"
            has_completed=1
          fi
          local fname
          fname="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
          printf "  ✓ %s\n" "$fname"
        fi
      done
    fi
    echo ""
    info "Run: pipeline.sh init <feature-name>"
    return 0
  fi

  local feature phase artifact history_count
  feature="$(read_field feature)"
  phase="$(read_field phase)"
  artifact="$(read_field current_artifact)"
  history_count="$(read_field history_count)"
  [ -z "$history_count" ] && history_count=0

  echo ""
  echo "┌─────────────────────────────────────────────┐"
  printf "│ Feature: %-35s│\n" "$feature"
  printf "│ Phase:   [%s/6] %-30s│\n" "$(phase_number "$phase")" "$phase"
  # Show branch/worktree info
  local br_info wt_info
  br_info="$(read_field branch 2>/dev/null || echo "")"
  wt_info="$(read_field worktree 2>/dev/null || echo "")"
  if [ -n "$wt_info" ]; then
    printf "│ Worktree: %-34s│\n" "$wt_info"
  elif [ -n "$br_info" ]; then
    printf "│ Branch:  %-35s│\n" "$br_info"
  fi
  if [ -n "$artifact" ]; then
    printf "│ Artifact: %-34s│\n" "$artifact"
  else
    printf "│ Artifact: %-34s│\n" "(none — register before approve)"
  fi
  # Show last completed task during implementation phase
  if [ "$phase" = "implementation" ]; then
    local lct
    lct="$(read_field last_completed_task 2>/dev/null || echo "")"
    if [ -n "$lct" ]; then
      printf "│ Last task: %-33s│\n" "$lct"
    fi
  fi
  echo "├─────────────────────────────────────────────┤"

  # Show pipeline progress
  local e_mark="○" r_mark="○" d_mark="○" t_mark="○" i_mark="○" rev_mark="○"
  case "$phase" in
    explore)        e_mark="●" ;;
    requirements)   e_mark="✓"; r_mark="●" ;;
    design)         e_mark="✓"; r_mark="✓"; d_mark="●" ;;
    task-plan)      e_mark="✓"; r_mark="✓"; d_mark="✓"; t_mark="●" ;;
    implementation) e_mark="✓"; r_mark="✓"; d_mark="✓"; t_mark="✓"; i_mark="●" ;;
    review)         e_mark="✓"; r_mark="✓"; d_mark="✓"; t_mark="✓"; i_mark="✓"; rev_mark="●" ;;
    done)           e_mark="✓"; r_mark="✓"; d_mark="✓"; t_mark="✓"; i_mark="✓"; rev_mark="✓" ;;
  esac
  printf "│ %s Ex → %s Rq → %s Ds → %s Tp → %s Im → %s Rv │\n" "$e_mark" "$r_mark" "$d_mark" "$t_mark" "$i_mark" "$rev_mark"
  echo "└─────────────────────────────────────────────┘"

  # Show history
  if [ "$history_count" -gt 0 ]; then
    echo ""
    echo "Completed phases:"
    local i=0
    while [ "$i" -lt "$history_count" ]; do
      local h_phase h_artifact h_approved
      h_phase="$(read_field "history_${i}_phase")"
      h_artifact="$(read_field "history_${i}_artifact")"
      h_approved="$(read_field "history_${i}_approved_at")"
      printf "  [%s] %-15s → %s (approved: %s)\n" "$((i+1))" "$h_phase" "$h_artifact" "$h_approved"
      i=$((i + 1))
    done
  fi

  # Hint for next action
  echo ""
  if [ "$phase" = "done" ]; then
    info "Pipeline complete."
  elif [ -z "$artifact" ]; then
    info "Next: register artifact with 'pipeline.sh artifact <path>'"
    info "Then: 'pipeline.sh approve' after user approval"
  else
    info "Artifact registered. Ask user to approve, then run 'pipeline.sh approve'"
  fi
  echo ""
}

cmd_artifact() {
  resolve_feature || die "No active pipeline. Run 'pipeline.sh init <feature>' first."

  local phase
  phase="$(read_field phase)"
  [ "$phase" = "done" ] && die "Pipeline is complete. Nothing to register."

  local path="$1"

  # If no path given, use the default: .spec/features/<feature>/<phase>.md
  if [ -z "$path" ]; then
    path="$FEATURE_DIR/${phase}.md"
  fi

  validate_artifact_path "$path"
  [ -f "$path" ] || die "Artifact file does not exist: $path"

  # Save a snapshot of the artifact being registered (revision tracking)
  local rev_count
  rev_count="$(read_field "revision_count_${phase}")"
  [ -z "$rev_count" ] && rev_count=0
  rev_count=$((rev_count + 1))
  local rev_name
  rev_name="${phase}-rev-${rev_count}-$(iso_now_compact).md"
  cp "$path" "$REVISIONS_DIR/$rev_name"
  write_field "revision_count_${phase}" "$rev_count"
  if [ "$rev_count" -gt 1 ]; then
    info "Revision $rev_count saved: $rev_name"
  fi

  write_field current_artifact "$path"
  rebuild_json
  info "Artifact registered for phase '$phase': $path"
}

cmd_approve() {
  resolve_feature || die "No active pipeline."

  local phase artifact history_count
  phase="$(read_field phase)"
  artifact="$(read_field current_artifact)"
  history_count="$(read_field history_count)"
  [ -z "$history_count" ] && history_count=0

  [ "$phase" = "done" ] && die "Pipeline already complete."
  [ -z "$artifact" ] && die "No artifact registered for phase '$phase'. Run 'pipeline.sh artifact <path>' first."
  [ -f "$artifact" ] || die "Artifact file no longer exists: $artifact. Re-register with 'pipeline.sh artifact <path>'."

  # Snapshot artifact contents
  cp "$artifact" "$APPROVED_DIR/${phase}.md"

  # Record base commit for review phase (git diff source)
  if [ "$phase" = "task-plan" ]; then
    local base_commit
    base_commit="$(git rev-parse HEAD 2>/dev/null || echo "")"
    write_field review_base_commit "$base_commit"
  fi

  # Record in history
  write_field "history_${history_count}_phase" "$phase"
  write_field "history_${history_count}_artifact" "$artifact"
  write_field "history_${history_count}_approved_at" "$(iso_now)"
  history_count=$((history_count + 1))
  write_field history_count "$history_count"

  # Advance phase
  local next
  next="$(next_phase "$phase")"
  write_field phase "$next"
  write_field current_artifact ""

  # Clear task tracking when leaving implementation
  if [ "$phase" = "implementation" ]; then
    write_field last_completed_task ""
  fi

  rebuild_json

  if [ "$next" = "done" ]; then
    echo ""
    echo "✅ Pipeline complete!"
    echo ""
    echo "All artifacts:"
    local i=0
    while [ "$i" -lt "$history_count" ]; do
      printf "  [%s] %s → %s\n" "$((i+1))" \
        "$(read_field "history_${i}_phase")" \
        "$(read_field "history_${i}_artifact")"
      i=$((i + 1))
    done
    echo ""
    local feat
    feat="$(read_field feature)"
    info "Artifacts saved in: .spec/features/$feat/"
    info "Next: check documentation (docs-check), then finish branch (pipeline.sh finish)"
  else
    info "Phase '$phase' approved."
    info "Advanced to: [$(phase_number "$next")/6] $next"
    info "Read template: ./templates/${next}.md"
  fi
}

cmd_task() {
  local task_id="$1"
  [ -z "$task_id" ] && die "Usage: pipeline.sh task <T-N>"

  resolve_feature || die "No active pipeline."

  local phase
  phase="$(read_field phase)"
  [ "$phase" = "implementation" ] || die "Task tracking is only available during implementation phase (current: $phase)."

  write_field last_completed_task "$task_id"
  rebuild_json
  info "Task $task_id marked complete"
}

cmd_history() {
  if [ ! -d "$FEATURES_DIR" ]; then
    info "No features found."
    return 0
  fi

  local found=0
  for kv in "$FEATURES_DIR"/*/pipeline.kv; do
    [ -f "$kv" ] || continue
    found=1
    local fname phase created
    fname="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
    phase="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
    created="$(grep "^created_at=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"

    local status_icon
    if [ "$phase" = "done" ]; then
      status_icon="✓"
    else
      status_icon="●"
    fi

    printf "  %s %-25s [%s/6] %-15s (created: %s)\n" \
      "$status_icon" "$fname" "$(phase_number "$phase")" "$phase" "$created"
  done

  if [ "$found" -eq 0 ]; then
    info "No features found."
  fi
}

cmd_revisions() {
  resolve_feature || die "No active pipeline."

  local phase
  phase="$(read_field phase)"

  local target_phase="${1:-$phase}"
  # Validate target phase
  case "$target_phase" in
    explore|requirements|design|task-plan|implementation|review|all) ;;
    *) die "Unknown phase: $target_phase. Use: explore, requirements, design, task-plan, implementation, review, or all." ;;
  esac

  local found=0
  local tmp
  tmp="$(mktemp)"
  if [ "$target_phase" = "all" ]; then
    echo "All revisions:"
    for p in explore requirements design task-plan implementation review; do
      find "$REVISIONS_DIR" -name "${p}-rev-*" 2>/dev/null | sort > "$tmp"
      while IFS= read -r f; do
        printf "  [%s] %s\n" "$p" "$(basename "$f")"
        found=1
      done < "$tmp"
    done
  else
    echo "Revisions for phase '$target_phase':"
    find "$REVISIONS_DIR" -name "${target_phase}-rev-*" 2>/dev/null | sort > "$tmp"
    while IFS= read -r f; do
      printf "  %s\n" "$(basename "$f")"
      found=1
    done < "$tmp"
  fi
  rm -f "$tmp"

  if [ "$found" -eq 0 ]; then
    info "No revisions recorded yet."
  fi
}

cmd_version() {
  echo "Spec-Driven Dev Pipeline v${VERSION}"
}

# check_file_staleness <file> <templates_dir> <freshness_days> <now_epoch>
# Outputs tab-separated: generated template age_days stale scope_changed
check_file_staleness() {
  local f="$1" templates_dir="$2" freshness_days="$3" now_epoch="$4"
  local generated="null" template="null" age_days="null" stale="false" scope_changed="null"

  local first_line
  first_line="$(head -1 "$f" 2>/dev/null)"
  case "$first_line" in
    *"<!-- generated:"*"template:"*"-->"*)
      local gen_date gen_tmpl
      gen_date="$(echo "$first_line" | sed 's/.*<!-- generated: \([0-9-]*\),.*/\1/')"
      gen_tmpl="$(echo "$first_line" | sed 's/.*template: \([^ ]*\) -->.*/\1/')"
      if [ -n "$gen_date" ]; then
        generated="\"$gen_date\""
        template="\"$gen_tmpl\""
        local gen_epoch
        gen_epoch="$(date -j -f '%Y-%m-%d' "$gen_date" '+%s' 2>/dev/null || date -d "$gen_date" '+%s' 2>/dev/null || echo 0)"
        if [ "$gen_epoch" -gt 0 ] && [ "$now_epoch" -gt 0 ]; then
          age_days=$(( (now_epoch - gen_epoch) / 86400 ))

          local tmpl_file="$templates_dir/$gen_tmpl"
          if [ -f "$tmpl_file" ]; then
            local scope_line
            scope_line="$(head -1 "$tmpl_file" 2>/dev/null)"
            case "$scope_line" in
              "<!-- scope:"*"-->")
                local patterns
                patterns="$(echo "$scope_line" | sed 's/<!-- scope: //' | sed 's/ -->//' | sed 's/,[[:space:]]*/\n/g' | tr '\n' ' ' | sed 's/[[:space:]]*$//')"
                if [ -n "$patterns" ]; then
                  local git_hits
                  # shellcheck disable=SC2086
                  git_hits="$(cd "$PROJECT_ROOT" && set -f && git log --oneline --since="$gen_date" -- $patterns 2>/dev/null | head -1)"
                  if [ -n "$git_hits" ]; then
                    scope_changed="true"
                    if [ "$age_days" -gt "$freshness_days" ]; then
                      stale="true"
                    fi
                  else
                    scope_changed="false"
                  fi
                fi
                ;;
              *)
                if [ "$age_days" -gt "$freshness_days" ]; then
                  stale="true"
                fi
                ;;
            esac
          else
            if [ "$age_days" -gt "$freshness_days" ]; then
              stale="true"
            fi
          fi
        fi
      fi
      ;;
  esac
  printf '%s\t%s\t%s\t%s\t%s' "$generated" "$template" "$age_days" "$stale" "$scope_changed"
}

cmd_docs_check() {
  local config_file="$CONFIG_FILE"
  local docs_dir=".spec"
  local freshness_days=30

  # Read docs_dir and doc_freshness_days from config.yaml if it exists
  if [ -f "$config_file" ]; then
    local configured_dir
    configured_dir="$(grep '^docs_dir:' "$config_file" 2>/dev/null | head -1 | sed 's/^docs_dir:[[:space:]]*//' | sed 's/[[:space:]]*$//')"
    if [ -n "$configured_dir" ]; then
      docs_dir="$configured_dir"
    fi
    local configured_days
    configured_days="$(grep '^doc_freshness_days:' "$config_file" 2>/dev/null | head -1 | sed 's/^doc_freshness_days:[[:space:]]*//' | sed 's/[[:space:]]*$//')"
    if [ -n "$configured_days" ]; then
      freshness_days="$configured_days"
    fi
  fi

  local full_path="$PROJECT_ROOT/$docs_dir"
  local templates_dir="$SKILL_DIR/templates/docs"
  local now_epoch
  now_epoch="$(date +%s 2>/dev/null || echo 0)"

  if [ -d "$full_path" ]; then
    printf '{"exists": true, "dir": "%s", "freshness_days": %d, "files": [' "$(json_escape "$docs_dir")" "$freshness_days"

    # Single scan: find files into tmpfile, iterate once, capture stale names
    local tmp_files tmp_stale
    tmp_files="$(mktemp)"
    tmp_stale="$(mktemp)"
    find "$full_path" -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort > "$tmp_files"

    local first=1
    while IFS= read -r f; do
      local fname result generated template age_days stale scope_changed
      fname="$(basename "$f")"
      result="$(check_file_staleness "$f" "$templates_dir" "$freshness_days" "$now_epoch")"

      generated="$(printf '%s' "$result" | cut -f1)"
      template="$(printf '%s' "$result" | cut -f2)"
      age_days="$(printf '%s' "$result" | cut -f3)"
      stale="$(printf '%s' "$result" | cut -f4)"
      scope_changed="$(printf '%s' "$result" | cut -f5)"

      if [ "$first" -eq 1 ]; then
        first=0
      else
        printf ', '
      fi
      printf '{"name": "%s", "generated": %s, "template": %s, "age_days": %s, "stale": %s, "scope_changed": %s}' \
        "$(json_escape "$fname")" "$generated" "$template" "$age_days" "$stale" "$scope_changed"

      if [ "$stale" = "true" ]; then
        echo "$fname" >> "$tmp_stale"
      fi
    done < "$tmp_files"

    printf '], "stale": ['
    # Read stale names from tmpfile (no re-scan needed)
    local sfirst=1
    if [ -s "$tmp_stale" ]; then
      while IFS= read -r sname; do
        if [ "$sfirst" -eq 1 ]; then
          sfirst=0
        else
          printf ', '
        fi
        printf '"%s"' "$(json_escape "$sname")"
      done < "$tmp_stale"
    fi
    printf ']}\n'

    rm -f "$tmp_files" "$tmp_stale"
  else
    printf '{"exists": false, "dir": "%s", "freshness_days": %d, "files": [], "stale": []}\n' "$(json_escape "$docs_dir")" "$freshness_days"
  fi
}

# --- standalone docs queue ---

DOCS_QUEUE_FILE="$PROJECT_ROOT/.spec/.docs-queue.kv"

docs_queue_read() {
  # docs_queue_read <key> — read value from queue file
  [ -f "$DOCS_QUEUE_FILE" ] || return 1
  local _line
  _line="$(grep "^$1=" "$DOCS_QUEUE_FILE" 2>/dev/null | head -1)" || return 1
  [ -n "$_line" ] || return 1
  printf '%s' "$_line" | sed "s/^$1=//"
}

docs_queue_write_status() {
  # docs_queue_write_status <index> <status>
  local idx="$1" status="$2"
  local tmp="$DOCS_QUEUE_FILE.tmp"
  grep -v "^template_${idx}_status=" "$DOCS_QUEUE_FILE" > "$tmp" 2>/dev/null || true
  echo "template_${idx}_status=$status" >> "$tmp"
  mv -f "$tmp" "$DOCS_QUEUE_FILE"
}

docs_template_name() {
  # Extract template name from generated file metadata (line 1)
  local f="$1"
  local first_line
  first_line="$(head -1 "$f" 2>/dev/null)"
  case "$first_line" in
    "<!-- generated:"*"-->")
      printf '%s' "$first_line" | sed -n 's/.*template:[[:space:]]*\([^[:space:]]*\)\.md[[:space:]]*-->/\1/p'
      ;;
  esac
}

cmd_docs_init() {
  local mode="all"
  local explicit_templates=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --all)    mode="all"; shift ;;
      --update) mode="update"; shift ;;
      -*)       die "Unknown flag for docs-init: $1" ;;
      *)
        explicit_templates="$explicit_templates $1"
        mode="explicit"
        shift
        ;;
    esac
  done

  [ -f "$DOCS_QUEUE_FILE" ] && die "Docs queue already exists at $DOCS_QUEUE_FILE. Run 'pipeline.sh docs-reset' first."

  local templates_dir="$SKILL_DIR/templates/docs"
  local docs_dir
  docs_dir="$(read_config docs_dir ".spec")"
  local full_path="$PROJECT_ROOT/$docs_dir"

  local templates=""
  case "$mode" in
    all)
      for f in "$templates_dir"/*.md; do
        local name
        name="$(basename "$f" .md)"
        [ "$name" = "README" ] && continue
        templates="$templates $name"
      done
      ;;
    update)
      [ -d "$full_path" ] || die "Docs directory '$docs_dir' does not exist. Use --all to bootstrap."
      local freshness_days
      freshness_days="$(read_config doc_freshness_days "30")"
      local now_epoch
      now_epoch="$(date +%s 2>/dev/null || echo 0)"
      # Iterate stale files, extract template names
      local tmp_list
      tmp_list="$(mktemp)"
      find "$full_path" -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort > "$tmp_list"
      while IFS= read -r f; do
        local result stale tmpl
        result="$(check_file_staleness "$f" "$templates_dir" "$freshness_days" "$now_epoch")"
        stale="$(printf '%s' "$result" | cut -f4)"
        if [ "$stale" = "true" ]; then
          tmpl="$(docs_template_name "$f")"
          if [ -n "$tmpl" ]; then
            case " $templates " in
              *" $tmpl "*) ;;
              *) templates="$templates $tmpl" ;;
            esac
          fi
        fi
      done < "$tmp_list"
      rm -f "$tmp_list"
      ;;
    explicit)
      templates="$explicit_templates"
      ;;
  esac

  mkdir -p "$(dirname "$DOCS_QUEUE_FILE")"

  # Build queue file
  local count=0
  local tmp_queue="$DOCS_QUEUE_FILE.tmp"
  {
    echo "created_at=$(iso_now)"
    echo "docs_dir=$docs_dir"
    echo "mode=$mode"
  } > "$tmp_queue"

  for t in $templates; do
    if [ ! -f "$templates_dir/$t.md" ]; then
      warn "Template not found, skipping: $t"
      continue
    fi
    echo "template_${count}=$t" >> "$tmp_queue"
    echo "template_${count}_status=pending" >> "$tmp_queue"
    count=$((count + 1))
  done
  echo "total=$count" >> "$tmp_queue"

  if [ "$count" -eq 0 ]; then
    rm -f "$tmp_queue"
    info "No templates to queue (nothing stale or none selected)."
    return 0
  fi

  mv -f "$tmp_queue" "$DOCS_QUEUE_FILE"

  info "Docs queue created: $count template(s), mode=$mode."
  info "Next: pipeline.sh docs-next"
  info ""
  info "Execution strategy:"
  info "  - If your toolset supports subagent dispatch (Task/Composer/etc):"
  info "    use SUBAGENT mode — dispatch up to 3 templates in parallel."
  info "  - Otherwise: SEQUENTIAL mode — one template per iteration."
  info "  See ./templates/docs-maintenance.md § Standalone Documentation Workflow."
}

cmd_docs_next() {
  [ -f "$DOCS_QUEUE_FILE" ] || die "No docs queue. Run 'pipeline.sh docs-init' first."

  local total
  total="$(docs_queue_read total)"
  [ -z "$total" ] && total=0

  local i=0
  local templates_dir="$SKILL_DIR/templates/docs"
  local docs_dir
  docs_dir="$(docs_queue_read docs_dir)"

  while [ "$i" -lt "$total" ]; do
    local status
    status="$(docs_queue_read "template_${i}_status")"
    if [ "$status" = "pending" ]; then
      local name
      name="$(docs_queue_read "template_${i}")"
      printf '%s\t%s\n' "$name" "$templates_dir/$name.md"
      # Sequential mode hint after position 3
      if [ "$i" -ge 3 ]; then
        info "" >&2
        info "Tip: if context feels heavy, start a fresh chat and resume" >&2
        info "     with 'pipeline.sh docs-status' (sequential mode)." >&2
      fi
      return 0
    fi
    i=$((i + 1))
  done

  info "Docs queue complete: all $total template(s) processed."
  info "Run 'pipeline.sh docs-reset' to clear the queue."
  return 0
}

cmd_docs_done() {
  local name="${1:-}"
  [ -z "$name" ] && die "Usage: pipeline.sh docs-done <template>"
  [ -f "$DOCS_QUEUE_FILE" ] || die "No docs queue. Run 'pipeline.sh docs-init' first."

  local total
  total="$(docs_queue_read total)"
  [ -z "$total" ] && total=0

  local i=0
  while [ "$i" -lt "$total" ]; do
    local entry
    entry="$(docs_queue_read "template_${i}")"
    if [ "$entry" = "$name" ]; then
      local status
      status="$(docs_queue_read "template_${i}_status")"
      if [ "$status" = "done" ]; then
        warn "Template '$name' already marked done."
        return 0
      fi
      docs_queue_write_status "$i" "done"
      info "Marked done: $name ($((i + 1))/$total)"
      # Check if queue is now complete
      local remaining=0
      local j=0
      while [ "$j" -lt "$total" ]; do
        local s
        s="$(docs_queue_read "template_${j}_status")"
        [ "$s" = "pending" ] && remaining=$((remaining + 1))
        j=$((j + 1))
      done
      if [ "$remaining" -eq 0 ]; then
        info "Docs queue complete. Run 'pipeline.sh docs-reset' to clear it."
      fi
      return 0
    fi
    i=$((i + 1))
  done

  die "Template '$name' not found in queue."
}

cmd_docs_status() {
  if [ ! -f "$DOCS_QUEUE_FILE" ]; then
    printf '{"exists": false}\n'
    return 0
  fi

  local total docs_dir mode created_at
  total="$(docs_queue_read total)"
  docs_dir="$(docs_queue_read docs_dir)"
  mode="$(docs_queue_read mode)"
  created_at="$(docs_queue_read created_at)"
  [ -z "$total" ] && total=0

  local completed=0
  local pending_list=""
  local current=""
  local first_pending_set=0
  local i=0
  while [ "$i" -lt "$total" ]; do
    local name status
    name="$(docs_queue_read "template_${i}")"
    status="$(docs_queue_read "template_${i}_status")"
    if [ "$status" = "done" ]; then
      completed=$((completed + 1))
    else
      if [ "$first_pending_set" -eq 0 ]; then
        current="$name"
        first_pending_set=1
      fi
      if [ -z "$pending_list" ]; then
        pending_list="\"$(json_escape "$name")\""
      else
        pending_list="$pending_list, \"$(json_escape "$name")\""
      fi
    fi
    i=$((i + 1))
  done

  printf '{"exists": true, "total": %d, "completed": %d, "current": %s, "pending": [%s], "mode": "%s", "docs_dir": "%s", "created_at": "%s"}\n' \
    "$total" \
    "$completed" \
    "$([ -n "$current" ] && printf '"%s"' "$(json_escape "$current")" || printf 'null')" \
    "$pending_list" \
    "$(json_escape "$mode")" \
    "$(json_escape "$docs_dir")" \
    "$(json_escape "$created_at")"
}

cmd_docs_reset() {
  if [ ! -f "$DOCS_QUEUE_FILE" ]; then
    info "No docs queue to reset."
    return 0
  fi

  local total completed=0
  total="$(docs_queue_read total)"
  [ -z "$total" ] && total=0
  local i=0
  while [ "$i" -lt "$total" ]; do
    local s
    s="$(docs_queue_read "template_${i}_status")"
    [ "$s" = "done" ] && completed=$((completed + 1))
    i=$((i + 1))
  done

  rm -f "$DOCS_QUEUE_FILE"
  info "Docs queue reset (was: $completed/$total completed)."
}

cmd_config_check() {
  [ -f "$CONFIG_FILE" ] || { info "No config file found: $CONFIG_FILE"; return 0; }

  local valid_keys=" context rules.explore rules.requirements rules.design rules.task-plan rules.implementation rules.review rules.docs test_skill test_reference docs_dir doc_freshness_days auto_branch branch_prefix auto_worktree worktree_dir "
  local errors=0

  info "Checking $CONFIG_FILE ..."

  # Extract keys and validate against whitelist
  local tmp
  tmp="$(mktemp)"
  grep '^[a-z]' "$CONFIG_FILE" 2>/dev/null > "$tmp" || true
  while IFS= read -r line; do
    local key
    key="$(printf '%s' "$line" | sed 's/:.*//')"
    case "$valid_keys" in
      *" $key "*) ;;
      *) warn "Unknown key: '$key'"; errors=$((errors + 1)) ;;
    esac
  done < "$tmp"
  rm -f "$tmp"

  # Type checks
  local val
  val="$(read_config doc_freshness_days "")"
  if [ -n "$val" ]; then
    case "$val" in
      *[!0-9]*) warn "doc_freshness_days must be numeric, got: '$val'"; errors=$((errors + 1)) ;;
    esac
  fi

  val="$(read_config auto_branch "")"
  if [ -n "$val" ]; then
    case "$val" in
      true|false|yes|no|1|0) ;;
      *) warn "auto_branch must be boolean (true/false/yes/no/1/0), got: '$val'"; errors=$((errors + 1)) ;;
    esac
  fi

  val="$(read_config auto_worktree "")"
  if [ -n "$val" ]; then
    case "$val" in
      true|false|yes|no|1|0) ;;
      *) warn "auto_worktree must be boolean (true/false/yes/no/1/0), got: '$val'"; errors=$((errors + 1)) ;;
    esac
  fi

  if [ "$errors" -eq 0 ]; then
    info "Config OK — all keys valid."
  else
    warn "$errors problem(s) found."
    return 1
  fi
}

cmd_inject() {
  local target_phase="${1:-}"
  local artifact_path="${2:-}"

  if [ -z "$target_phase" ] || [ -z "$artifact_path" ]; then
    die "Usage: pipeline.sh inject <phase> <path>"
  fi

  # Validate target phase
  case "$target_phase" in
    explore|requirements|design|task-plan|implementation|review) ;;
    *) die "Unknown phase: $target_phase. Use: explore, requirements, design, task-plan, implementation, review." ;;
  esac

  resolve_feature || die "No active pipeline. Run 'pipeline.sh init <feature>' first."

  local current_phase
  current_phase="$(read_field phase)"
  [ "$current_phase" = "done" ] && die "Pipeline already complete."

  # Validate current phase <= target phase
  local current_num target_num
  current_num="$(phase_order "$current_phase")"
  target_num="$(phase_order "$target_phase")"
  if [ "$current_num" -gt "$target_num" ]; then
    die "Cannot inject backward: current phase is '$current_phase' ($current_num), target is '$target_phase' ($target_num)."
  fi

  validate_artifact_path "$artifact_path"
  [ -f "$artifact_path" ] || die "Artifact file does not exist: $artifact_path"

  # Lightweight content validation
  case "$target_phase" in
    requirements)
      if ! grep -q 'WHEN\|SHALL' "$artifact_path" 2>/dev/null; then
        warn "Requirements artifact should contain WHEN/SHALL keywords."
      fi
      ;;
    design)
      if ! grep -q 'Correctness\|Property' "$artifact_path" 2>/dev/null; then
        warn "Design artifact should contain Correctness Properties."
      fi
      ;;
  esac

  # Skip intermediate phases (record as injected in history)
  local p="$current_phase"
  local history_count
  history_count="$(read_field history_count)"
  [ -z "$history_count" ] && history_count=0

  while [ "$p" != "$target_phase" ]; do
    write_field "history_${history_count}_phase" "$p"
    write_field "history_${history_count}_artifact" "(injected)"
    write_field "history_${history_count}_approved_at" "$(iso_now)"
    history_count=$((history_count + 1))
    p="$(next_phase "$p")"
  done

  # Set to target phase and register artifact
  write_field phase "$target_phase"
  write_field current_artifact "$artifact_path"
  write_field history_count "$history_count"

  # Save revision snapshot
  local rev_count
  rev_count="$(read_field "revision_count_${target_phase}")"
  [ -z "$rev_count" ] && rev_count=0
  rev_count=$((rev_count + 1))
  local rev_name
  rev_name="${target_phase}-rev-${rev_count}-$(iso_now_compact).md"
  cp "$artifact_path" "$REVISIONS_DIR/$rev_name"
  write_field "revision_count_${target_phase}" "$rev_count"

  # Capture review_base_commit if injecting into implementation/review and not already set
  case "$target_phase" in
    implementation|review)
      local rbc
      rbc="$(read_field review_base_commit 2>/dev/null || echo "")"
      if [ -z "$rbc" ]; then
        local head_commit
        head_commit="$(git rev-parse HEAD 2>/dev/null || echo "")"
        if [ -n "$head_commit" ]; then
          write_field review_base_commit "$head_commit"
          info "Captured review_base_commit: $(printf '%.8s' "$head_commit")"
        fi
      fi
      ;;
  esac

  rebuild_json

  local skipped=$((target_num - current_num))
  if [ "$skipped" -gt 0 ]; then
    info "$skipped phase(s) skipped to reach '$target_phase'."
  fi
  info "Artifact injected for phase '$target_phase': $artifact_path"
  info "Ask user to approve, then run 'pipeline.sh approve'"
}

cmd_finish() {
  # Parse finish-specific flags and action
  local action=""
  local confirm=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --confirm) confirm="yes"; shift ;;
      -*)        die "Unknown flag for finish: $1" ;;
      *)
        [ -n "$action" ] && die "Unexpected argument: $1"
        action="$1"; shift
        ;;
    esac
  done

  [ -z "$action" ] && die "Usage: pipeline.sh finish <merge|pr|keep|discard> [--confirm]"

  # Validate action
  case "$action" in
    merge|pr|keep|discard) ;;
    *) die "Unknown finish action: $action. Use: merge, pr, keep, discard." ;;
  esac

  # Resolve feature (supports completed pipelines)
  if [ -n "$EXPLICIT_FEATURE" ]; then
    if [ ! -f "$FEATURES_DIR/$EXPLICIT_FEATURE/pipeline.kv" ]; then
      die "Feature '$EXPLICIT_FEATURE' not found."
    fi
    set_feature_context "$EXPLICIT_FEATURE"
    validate_kv
  else
    # For finish, we need to find a done-but-not-finished feature
    local found=""
    local count=0
    if [ -d "$FEATURES_DIR" ]; then
      for kv in "$FEATURES_DIR"/*/pipeline.kv; do
        [ -f "$kv" ] || continue
        local p fa
        p="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        fa="$(grep "^finish_action=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        if [ "$p" = "done" ] && [ -z "$fa" ]; then
          local fn
          fn="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
          found="$fn"
          count=$((count + 1))
        fi
      done
    fi
    if [ "$count" -gt 1 ]; then
      warn "Multiple completed pipelines awaiting finish:"
      for kv in "$FEATURES_DIR"/*/pipeline.kv; do
        [ -f "$kv" ] || continue
        local p fa fn
        p="$(grep "^phase=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        fa="$(grep "^finish_action=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        fn="$(grep "^feature=" "$kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
        if [ "$p" = "done" ] && [ -z "$fa" ]; then
          echo "  - $fn" >&2
        fi
      done
      die "Use --feature <name> to select one."
    fi
    if [ -z "$found" ]; then
      die "No completed pipeline awaiting finish. Run 'pipeline.sh history' to list features."
    fi
    set_feature_context "$found"
    validate_kv
  fi

  local phase
  phase="$(read_field phase)"
  [ "$phase" != "done" ] && die "Pipeline not complete (current phase: $phase). Finish is only available after all phases are done."

  # Check idempotency
  local existing_action
  existing_action="$(read_field finish_action 2>/dev/null || echo "")"
  if [ -n "$existing_action" ]; then
    die "Already finished (action: $existing_action). Nothing to do."
  fi

  # Git availability check
  if [ "$action" != "keep" ]; then
    if ! command -v git >/dev/null 2>&1; then
      die "Git not found. Use 'finish keep' to skip git operations."
    fi
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
      die "Not a git repository. Use 'finish keep' to skip git operations."
    fi
  fi

  local feat branch current_branch worktree
  feat="$(read_field feature)"
  branch="$(read_field branch 2>/dev/null || echo "")"
  worktree="$(read_field worktree 2>/dev/null || echo "")"
  current_branch="$(git branch --show-current 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")"

  # Determine base branch for merge/discard
  local base_branch=""
  if [ "$action" = "merge" ] || [ "$action" = "discard" ]; then
    # Try to find default branch
    if git rev-parse --verify main >/dev/null 2>&1; then
      base_branch="main"
    elif git rev-parse --verify master >/dev/null 2>&1; then
      base_branch="master"
    else
      die "Cannot determine base branch (no main or master). Merge/discard manually."
    fi

    # Check for uncommitted changes
    if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
      die "Working tree has uncommitted changes. Commit or stash before '$action'."
    fi
  fi

  # Determine which branch to operate on
  local target_branch="${branch:-$current_branch}"

  case "$action" in
    merge)
      [ -z "$target_branch" ] && die "No branch to merge (not on a feature branch and no branch recorded)."
      [ "$target_branch" = "$base_branch" ] && die "Already on $base_branch. Nothing to merge."
      # If worktree, remove it first (must not be in worktree dir when removing)
      if [ -n "$worktree" ] && git worktree list 2>/dev/null | grep -q "$worktree"; then
        git worktree remove "$worktree" || die "Failed to remove worktree '$worktree'."
        info "Removed worktree: $worktree"
      fi
      git checkout "$base_branch" || die "Failed to checkout $base_branch."
      git merge "$target_branch" || die "Merge failed. Resolve conflicts and retry."
      write_field finish_action "merge"
      write_field finished_at "$(iso_now)"
      write_field finish_base "$base_branch"
      rebuild_json
      info "Merged '$target_branch' into '$base_branch'."
      info "Tip: run tests to verify, then 'git branch -d $target_branch' to clean up."
      ;;

    pr)
      [ -z "$target_branch" ] && die "No branch to push (not on a feature branch and no branch recorded)."
      [ "$target_branch" = "$base_branch" ] && die "Already on $base_branch. Nothing to push."
      git push -u origin "$target_branch" || die "Failed to push '$target_branch'."
      write_field finish_action "pr"
      write_field finished_at "$(iso_now)"
      write_field finish_base "${base_branch:-}"
      rebuild_json
      info "Branch '$target_branch' pushed to origin."
      info "Create a pull request: gh pr create --fill"
      ;;

    keep)
      write_field finish_action "keep"
      write_field finished_at "$(iso_now)"
      rebuild_json
      info "Branch kept as-is. Handle manually when ready."
      ;;

    discard)
      [ "$confirm" != "yes" ] && die "Discard deletes the branch and all unmerged commits. Re-run with --confirm to proceed: pipeline.sh finish discard --confirm"
      [ -z "$target_branch" ] && die "No branch to discard (not on a feature branch and no branch recorded)."
      [ "$target_branch" = "$base_branch" ] && die "Cannot discard $base_branch."
      # If worktree, force-remove it first
      if [ -n "$worktree" ] && git worktree list 2>/dev/null | grep -q "$worktree"; then
        git worktree remove --force "$worktree" || die "Failed to remove worktree '$worktree'."
        info "Removed worktree: $worktree"
      fi
      git checkout "$base_branch" || die "Failed to checkout $base_branch."
      git branch -D "$target_branch" || die "Failed to delete branch '$target_branch'."
      write_field finish_action "discard"
      write_field finished_at "$(iso_now)"
      write_field finish_base "$base_branch"
      rebuild_json
      info "Branch '$target_branch' discarded."
      ;;
  esac
}

cmd_abandon() {
  local feature="${1:-}"

  # If --feature was specified globally, use it
  if [ -z "$feature" ] && [ -n "$EXPLICIT_FEATURE" ]; then
    feature="$EXPLICIT_FEATURE"
  fi

  # If still empty, try to resolve active feature
  if [ -z "$feature" ]; then
    feature="$(detect_active_feature)" || die "Multiple active pipelines. Use: pipeline.sh abandon <feature-name>"
    [ -z "$feature" ] && die "No active pipeline to abandon."
  fi

  local fdir="$FEATURES_DIR/$feature"
  [ -f "$fdir/pipeline.kv" ] || die "Feature '$feature' not found."

  local phase
  phase="$(grep "^phase=" "$fdir/pipeline.kv" 2>/dev/null | head -1 | cut -d'=' -f2-)"
  [ "$phase" = "done" ] && die "Feature '$feature' is already completed."

  set_feature_context "$feature"
  write_field phase "done"
  write_field abandoned_at "$(iso_now)"
  write_field finish_action "abandoned"
  write_field finished_at "$(iso_now)"

  # Clean up worktree if present
  local wt
  wt="$(read_field worktree 2>/dev/null || echo "")"
  if [ -n "$wt" ] && command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    if git worktree list 2>/dev/null | grep -q "$wt"; then
      git worktree remove --force "$wt" 2>/dev/null && info "Removed worktree: $wt"
    fi
  fi

  rebuild_json

  info "Feature '$feature' abandoned (was in phase: $phase)."
  info "Artifacts remain in: .spec/features/$feature/"
}

cmd_help() {
  echo "Spec-Driven Dev Pipeline v${VERSION}"
  echo ""
  echo "Usage: sh pipeline.sh [--feature <name>] <command> [args]"
  echo ""
  echo "Global flags:"
  echo "  --feature <name>  Select feature (needed when multiple are active)"
  echo ""
  echo "Commands:"
  echo "  init [--branch|--worktree|--no-branch] <feature>"
  echo "                    Start a new pipeline (kebab-case name)"
  echo "                    --branch: create git branch <prefix><name>"
  echo "                    --worktree: create git worktree in <worktree_dir>/<name>"
  echo "                    --no-branch: skip auto-branch/worktree from config"
  echo "  status            Show current phase, artifacts, progress"
  echo "  artifact [path]   Register output artifact for current phase"
  echo "  approve           Advance to next phase (needs artifact)"
  echo "  revisions [phase] Show revision history (current phase or specify: explore, all)"
  echo "  history           Show all features and their status"
  echo "  docs-check        Check project documentation status (JSON)"
  echo "  docs-init [--all|--update|<template>...]"
  echo "                    Create standalone docs generation queue"
  echo "                    --all: queue all available templates"
  echo "                    --update: queue only stale templates"
  echo "                    <template>...: queue explicit templates by name"
  echo "  docs-next         Print next pending template (name + path)"
  echo "  docs-done <name>  Mark template as completed in queue"
  echo "  docs-status       Show docs queue progress (JSON)"
  echo "  docs-reset        Clear the docs queue"
  echo "  task <T-N>        Mark implementation task as completed (resume tracking)"
  echo "  config-check      Validate .spec/config.yaml keys and types"
  echo "  inject <phase> <path>"
  echo "                    Inject pre-written artifact and skip to that phase"
  echo "  finish <action>   Finalize branch after pipeline completes"
  echo "                    Actions: merge, pr, keep, discard (--confirm)"
  echo "  abandon [feature] Abandon an active pipeline (marks as done)"
  echo "  version           Show version"
  echo "  help              Show this message"
  echo ""
  echo "Workflow (6 phases):"
  echo "  1. init my-feature"
  echo "  2. (agent reads templates/explore.md, investigates)"
  echo "  3. artifact  ← writes .spec/features/my-feature/explore.md"
  echo "  4. approve   ← user confirms"
  echo "  5. (agent reads templates/requirements.md, generates doc)"
  echo "  6. artifact  ← writes .spec/features/my-feature/requirements.md"
  echo "  7. approve   ← user confirms"
  echo "  8. (agent reads templates/design.md, generates doc)"
  echo "  9. artifact  ← writes .spec/features/my-feature/design.md"
  echo " 10. approve   ← user confirms"
  echo " 11. (agent reads templates/task-plan.md, creates TDD plan)"
  echo " 12. artifact  ← writes .spec/features/my-feature/task-plan.md"
  echo " 13. approve   ← user confirms"
  echo " 14. (agent reads templates/implementation.md, executes TDD plan)"
  echo " 15. artifact  ← writes .spec/features/my-feature/implementation.md"
  echo " 16. approve   ← user confirms"
  echo " 17. (agent reads templates/review.md, reviews code)"
  echo " 18. artifact  ← writes .spec/features/my-feature/review.md"
  echo " 19. approve   ← user confirms → done!"
  echo " 20. docs-check ← update project documentation if needed"
  echo " 21. finish     ← merge, push PR, keep, or discard branch"
  echo ""
  echo "All artifacts are saved permanently in .spec/features/<feature>/ and tracked by git."
  echo "Tip: use 'revisions' to see previous versions of an artifact within a phase."
}

# --- main ---

# Parse global flags before command dispatch
while [ $# -gt 0 ]; do
  case "$1" in
    --feature)
      [ -n "$2" ] || die "--feature requires a value"
      EXPLICIT_FEATURE="$2"
      shift 2
      ;;
    *) break ;;
  esac
done

case "${1:-help}" in
  init)     shift; cmd_init "$@" ;;
  status)   cmd_status ;;
  artifact) shift; cmd_artifact "$@" ;;
  approve)  cmd_approve ;;
  revisions) shift; cmd_revisions "$@" ;;
  history)  cmd_history ;;
  docs-check) cmd_docs_check ;;
  task)     shift; cmd_task "$@" ;;
  config-check) cmd_config_check ;;
  inject)   shift; cmd_inject "$@" ;;
  finish)   shift; cmd_finish "$@" ;;
  abandon)  shift; cmd_abandon "$@" ;;
  docs-init)   shift; cmd_docs_init "$@" ;;
  docs-next)   cmd_docs_next ;;
  docs-done)   shift; cmd_docs_done "$@" ;;
  docs-status) cmd_docs_status ;;
  docs-reset)  cmd_docs_reset ;;
  version|--version|-v) cmd_version ;;
  help|--help|-h) cmd_help ;;
  *)        die "Unknown command: $1. Run 'pipeline.sh help' for usage." ;;
esac
