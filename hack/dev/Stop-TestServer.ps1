if (-Not (Test-Path currentJob.Id)) {
  Write-Warning "Server is not running...nothing to stop"
  break
}

$JobId = Get-Content -Path currentJob.Id -Raw -Encoding ascii



Write-Host "Stopping Job: $JobId"
Stop-Job -Id $JobId

Write-Host "Job Id: $JobId"
Receive-Job -Id $JobId

Write-Host "Removing Job: $JobId"
Remove-Job -Id $JobId
Remove-Item -Path currentJob.Id
