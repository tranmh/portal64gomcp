$body = @{
    name = "get_player_profile"
    arguments = @{
        player_id = "C0327-297"
    }
} | ConvertTo-Json

Write-Host "Testing get_player_profile with player_id: C0327-297"
Write-Host "Request body: $body"
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8888/tools/call" -Method POST -Body $body -ContentType "application/json"
    Write-Host "Response received successfully:"
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)"
    Write-Host "Status Description: $($_.Exception.Response.StatusDescription)"
}
