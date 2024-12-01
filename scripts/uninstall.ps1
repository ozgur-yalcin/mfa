Write-Host "Uninstalling mfa" -ForegroundColor Green

$destination = "$env:LOCALAPPDATA\mfa"

Remove-Item -Path $destination -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "Uninstall mfa successfully" -ForegroundColor Green

$separator = [System.IO.Path]::PathSeparator
$modifiedPath = $env:Path -split $separator -ne $destination -join $separator
[Environment]::SetEnvironmentVariable("Path", $modifiedPath, [EnvironmentVariableTarget]::User)

Write-Host "The mfa entry in the environment variable Path has been deleted" -ForegroundColor Green