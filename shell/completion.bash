#!/usr/bin/env bash

_ezysearch_completion() {
    local cur prev opts managers shells
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--help --version --install --config --config-path --print-config --default-config --doctor --check --completion --completions --list-package-managers --package-manager --manager --auto --yay --pacman --apt --brew --homebrew --hombrew --dnf --zypper -h -v -V"
    managers="auto yay pacman apt brew homebrew dnf zypper"
    shells="bash zsh"

    if [[ ${prev} == "--package-manager" || ${prev} == "--manager" ]] ; then
        COMPREPLY=( $(compgen -W "${managers}" -- ${cur}) )
        return 0
    fi

    if [[ ${prev} == "--completion" || ${prev} == "--completions" ]] ; then
        COMPREPLY=( $(compgen -W "${shells}" -- ${cur}) )
        return 0
    fi

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _ezysearch_completion ezysearch
