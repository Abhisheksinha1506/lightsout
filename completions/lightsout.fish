# fish completion for lightsout
complete -c lightsout -f
complete -c lightsout -n __fish_use_subcommand -a keep -d 'Always keep apps open'
complete -c lightsout -n __fish_use_subcommand -a unkeep -d 'Stop always keeping apps open'
complete -c lightsout -n __fish_use_subcommand -a completion -d 'Print a completion script'
complete -c lightsout -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish powershell'
complete -c lightsout -n 'not __fish_seen_subcommand_from keep unkeep completion' -s n -l dry-run -d 'Show what would be quit'
complete -c lightsout -n 'not __fish_seen_subcommand_from keep unkeep completion' -s y -l yes -d 'Do not ask which app you are running from'
complete -c lightsout -n 'not __fish_seen_subcommand_from keep unkeep completion' -l timeout -x -d 'Seconds to wait for each app'
complete -c lightsout -s h -l help -d 'Show help'
complete -c lightsout -s v -l version -d 'Show the version'
