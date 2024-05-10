package api

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var _ Handler = (*Mod)(nil)

type Completion struct{}

func (c Completion) Command() *cli.Command {
	zsh := &cli.Command{
		Name:        "zsh",
		Usage:       "zsh",
		UsageText:   "zsh",
		Description: "zsh",
		Action:      c.completionZsh,
	}
	bash := &cli.Command{
		Name:        "bash",
		Usage:       "bash",
		UsageText:   "bash",
		Description: "bash",
		Action:      c.completionBash,
	}

	return &cli.Command{
		Name:                   "completion",
		Usage:                  "completion",
		UsageText:              "completion",
		Description:            "completion",
		ArgsUsage:              "",
		Category:               "",
		BashComplete:           nil,
		Before:                 nil,
		After:                  nil,
		Action:                 nil,
		OnUsageError:           nil,
		Subcommands:            []*cli.Command{zsh, bash},
		Flags:                  []cli.Flag{},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

func (c Completion) completionZsh(ctx *cli.Context) error {
	fmt.Println(`
#compdef easyp

_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}

compdef _cli_zsh_autocomplete easyp`)
	return nil
}

func (c Completion) completionBash(ctx *cli.Context) error {
	fmt.Println(`
#! /bin/bash

: ${easyp:=$(basename ${BASH_SOURCE})}

# Macs have bash3 for which the bash-completion package doesn't include
# _init_completion. This is a minimal version of that function.
_cli_init_completion() {
  COMPREPLY=()
  _get_comp_words_by_ref "$@" cur prev words cword
}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base words
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if declare -F _init_completion >/dev/null 2>&1; then
      _init_completion -n "=:" || return
    else
      _cli_init_completion -n "=:" || return
    fi
    words=("${words[@]:0:$cword}")
    if [[ "$cur" == "-"* ]]; then
      requestComp="${words[*]} ${cur} --generate-bash-completion"
    else
      requestComp="${words[*]} --generate-bash-completion"
    fi
    opts=$(eval "${requestComp}" 2>/dev/null)
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete $easyp
unset easyp`)
	return nil
}
