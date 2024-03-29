go mod init github.com/kcloutie/knot
go mod tidy

# =====================================================================================================
# Basic log example
# =====================================================================================================
$dataJson = @"
{
  "test": "test123",
  "prop1": "value1"
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
Invoke-WebRequest -Uri http://localhost:8080/api/v1/pubsub -Method Post -Body ($Payload | ConvertTo-Json -Compress) -ContentType "application/json" -Headers $Headers

# =====================================================================================================
# Github log example
# =====================================================================================================
$dataJson = Get-Content -Path "test/testdata/data-example.json" -Raw | Out-String
$enc = [system.Text.Encoding]::UTF8
$data = $enc.GetBytes($dataJson) 

$Payload = @{
  ID = "1234"
  Attributes = @{
    "test" = "github"
  }
  Data = $data
}
$Headers = @{
  "X-Cloud-Trace-Context" = (New-Guid | Select-Object -ExpandProperty Guid)
}
Invoke-WebRequest -Uri http://localhost:8080/api/v1/pubsub -Method Post -Body ($Payload | ConvertTo-Json -Compress) -ContentType "application/json" -Headers $Headers



export KUBECONFIG=/home/kcloutie/.kube/config.kind
kubectl logs deployment/knot-controller-manager -n knot
kubectl get deployment knot-controller-manager -n knot

kubectl delete deployment knot-controller-manager -n knot
make rdev
kubectl logs deployment/ingress-nginx-controller  -n ingress-nginx

kubectl get events --sort-by='.metadata.creationTimestamp' -A
kubectl get pods -n knot