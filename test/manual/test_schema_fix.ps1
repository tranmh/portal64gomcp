$getToolsBody = @{
    name = "get_tools"
    arguments = @{}
} | ConvertTo-Json

Write-Host "Testing tool schema for get_player_profile"
Write-Host "=========================================="
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8888/tools/list" -Method POST -ContentType "application/json"
    Write-Host "Available tools and their schemas:"
    $tools = $response.tools | Where-Object { $_.name -eq "get_player_profile" }
    if ($tools) {
        Write-Host "get_player_profile tool found with schema:"
        $tools | ConvertTo-Json -Depth 10
    } else {
        Write-Host "get_player_profile tool not found in list"
    }
} catch {
    Write-Host "Error getting tools list: $($_.Exception.Message)"
}

Write-Host ""
Write-Host "=========================================="
Write-Host "Testing get_player_profile function call"
Write-Host "=========================================="

$getPlayerBody = @{
    name = "get_player_profile"
    arguments = @{
        player_id = "C0327-297"
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8888/tools/call" -Method POST -Body $getPlayerBody -ContentType "application/json"
    Write-Host "Response received successfully:"
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $($_.Exception.Message)"
}
