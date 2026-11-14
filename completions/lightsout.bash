# bash completion for lightsout
_lightsout() {
	local cur=${COMP_WORDS[COMP_CWORD]}
	local prev=${COMP_WORDS[COMP_CWORD-1]}
	if [[ $COMP_CWORD -ge 2 && ${COMP_WORDS[1]} == completion ]]; then
		COMPREPLY=($(compgen -W "bash zsh fish powershell" -- "$cur"))
	elif [[ $prev == --timeout ]]; then
		COMPREPLY=()
	elif [[ $COMP_CWORD -eq 1 ]]; then
		COMPREPLY=($(compgen -W "keep unkeep completion --dry-run --yes --timeout --help --version" -- "$cur"))
	elif [[ ${COMP_WORDS[1]} != keep && ${COMP_WORDS[1]} != unkeep ]]; then
		COMPREPLY=($(compgen -W "--dry-run --yes --timeout" -- "$cur"))
	fi
}
complete -F _lightsout lightsout
