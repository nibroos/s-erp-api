$HealthUrl = "http://localhost:4001/api/v1/health"
$ConsumerHealthUrl = "http://localhost:4010/health"
$ContainerName = "s-erp-api-service-1"
$ConsumerContainerName = "s-erp-api-consumer-service-1"
$Interval = 5 # seconds

while ($true) {
    # Check REST API health
    try {
        $response = Invoke-WebRequest -Uri $HealthUrl -UseBasicParsing -TimeoutSec 3
        if ($response.StatusCode -eq 200) {
            Write-Host "$(Get-Date) REST API health check passed (HTTP 200)."
        } else {
            Write-Host "$(Get-Date) REST API health check failed (HTTP $($response.StatusCode)). Restarting container $ContainerName..."
            docker restart $ContainerName
        }
    } catch {
        Write-Host "$(Get-Date) REST API health check failed (no response). Restarting container $ContainerName..."
        docker restart $ContainerName
    }

    # Check Consumer service health
    try {
        $response = Invoke-WebRequest -Uri $ConsumerHealthUrl -UseBasicParsing -TimeoutSec 3
        if ($response.StatusCode -eq 200) {
            Write-Host "$(Get-Date) Consumer health check passed (HTTP 200)."
        } else {
            Write-Host "$(Get-Date) Consumer health check failed (HTTP $($response.StatusCode)). Restarting container $ConsumerContainerName..."
            docker restart $ConsumerContainerName
        }
    } catch {
        Write-Host "$(Get-Date) Consumer health check failed (no response). Restarting container $ConsumerContainerName..."
        docker restart $ConsumerContainerName
    }

    Start-Sleep -Seconds $Interval
}