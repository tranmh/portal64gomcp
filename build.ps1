$env:CGO_ENABLED = "0"

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags="-w -s" -o "bin\portal64gomcp-windows-amd64.exe" .\cmd\server\main.go

$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-w -s" -o "bin\portal64gomcp-linux-amd64" .\cmd\server\main.go
