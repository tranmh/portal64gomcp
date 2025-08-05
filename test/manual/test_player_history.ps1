$body = @{
    name = "get_player_rating_history"
    arguments = @{
        player_id = "C0327-297"
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8888/tools/call" -Method POST -Body $body -ContentType "application/json"
    $response | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Error: $($_.Exception.Message)"
}
