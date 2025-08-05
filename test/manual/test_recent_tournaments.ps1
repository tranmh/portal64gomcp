$body = @{
    name = "get_recent_tournaments"
    arguments = @{
        days = 30
        limit = 20
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8888/tools/call" -Method POST -Body $body -ContentType "application/json"
    $response | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Error: $($_.Exception.Message)"
}
