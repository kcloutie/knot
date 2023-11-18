if (Test-Path currentJob.Id) {
  Write-Error "Server is already running..."
  break
}

$ConfigFilePath = "test/files/serverConfig.json"
if ($null -ne $env:KNOT_CONFIG) {
  $ConfigFilePath = $env:KNOT_CONFIG
}

$ListeningAddr = ":8080"
if ($null -ne $env:KNOT_LISTEN_ADDR) {
  $ListeningAddr = $env:KNOT_LISTEN_ADDR
}

Write-Host "Starting server for e2e tests..."

$Job = Start-Job -ScriptBlock {
  Write-Host "ConfigFilePath $($args[0])"
  Write-Host "ListeningAddr $($args[1])"
  Write-Host "Starting server..."
  &./bin/knot run server --config-file-path $args[0] --listen-addr $args[1] --no-color
} -ArgumentList $ConfigFilePath, $ListeningAddr -Name "KNOT_TEST_SVR"

$Job
"$($Job.Id)" | Set-Content -Path currentJob.Id -Encoding ascii -NoNewline
Start-Sleep -Seconds 3
