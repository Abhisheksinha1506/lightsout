# PowerShell completion for lightsout. Add to your profile with:
#   lightsout completion powershell | Out-String | Invoke-Expression
Register-ArgumentCompleter -Native -CommandName lightsout, lightsout.exe -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $words = $commandAst.CommandElements | ForEach-Object { $_.ToString() }
    if ($words.Count -ge 2 -and $words[1] -eq 'completion') {
        $choices = 'bash', 'zsh', 'fish', 'powershell'
    } elseif ($words.Count -le 2) {
        $choices = 'keep', 'unkeep', 'completion', '--dry-run', '--yes', '--timeout', '--help', '--version'
    } else {
        $choices = '--dry-run', '--yes', '--timeout'
    }
    $choices | Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}
