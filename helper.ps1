go mod init github.com/kcloutie/knot
go mod tidy

$dataJson = @"
{
  "test": "test123"
}
"@
$enc = [system.Text.Encoding]::UTF8
$data = $enc.GetBytes($dataJson) 

$Payload = @{
  ID = "1234"
  Attributes = @{
    "enabled" = "true"
  }
  Data = $data
}
$Headers = @{
  "X-Cloud-Trace-Context" = (New-Guid | Select-Object -ExpandProperty Guid)
}
Invoke-RestMethod -Uri http://localhost:8080/api/v1/pubsub -Method Post -Body ($Payload | ConvertTo-Json -Compress) -ContentType "application/json" -Headers $Headers
