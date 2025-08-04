# Portal64 MCP Server - Deployment Guide

This guide covers deployment options and best practices for the Portal64 MCP Server.

## Prerequisites

### System Requirements
- Go 1.21 or later
- Access to Portal64 API (default: http://localhost:8080)
- Sufficient disk space for binary (~10MB)
- Network connectivity to Portal64 API

### Runtime Dependencies
The compiled binary has no external runtime dependencies. All required libraries are statically linked.

## Build Process

### Local Development Build
```bash
# Navigate to project directory
cd portal64gomcp

# Download dependencies
go mod tidy

# Build for current platform
go build -o bin/portal64-mcp ./cmd/server

# Or use make
make build
```

### Production Build
```bash
# Build with optimizations
make build-prod

# Cross-compile for different platforms
make build-all
```

### Cross-Platform Builds
The Makefile supports building for multiple platforms:

- Linux AMD64: `portal64-mcp-linux-amd64`
- Linux ARM64: `portal64-mcp-linux-arm64`
- Windows AMD64: `portal64-mcp-windows-amd64.exe`
- macOS AMD64: `portal64-mcp-darwin-amd64`
- macOS ARM64: `portal64-mcp-darwin-arm64`

## Configuration

### Environment Variables
Set these environment variables for runtime configuration:

```bash
# Portal64 API configuration
export PORTAL64_API_URL="http://localhost:8080"
export API_TIMEOUT="30s"

# Logging configuration
export LOG_LEVEL="info"
```

### Configuration File
Create a `config.yaml` file for structured configuration:

```yaml
api:
  base_url: "http://localhost:8080"
  timeout: "30s"

mcp:
  port: 3000

logging:
  level: "info"
  format: "json"
```

## Deployment Options

### Option 1: Direct Execution (Stdio Mode)
The MCP server runs in stdio mode by default, suitable for direct integration with MCP clients.

```bash
# Run with default configuration
./portal64-mcp

# Run with custom config
./portal64-mcp -config /path/to/config.yaml

# Run with debug logging
./portal64-mcp -log-level debug
```

### Option 2: Systemd Service (Linux)
Create a systemd service for automated startup and management.

**Create service file:** `/etc/systemd/system/portal64-mcp.service`

```ini
[Unit]
Description=Portal64 MCP Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=mcp
Group=mcp
WorkingDirectory=/opt/portal64-mcp
ExecStart=/opt/portal64-mcp/bin/portal64-mcp -config /opt/portal64-mcp/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Environment variables
Environment=PORTAL64_API_URL=http://localhost:8080
Environment=LOG_LEVEL=info

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=/opt/portal64-mcp/logs

[Install]
WantedBy=multi-user.target
```

**Enable and start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable portal64-mcp
sudo systemctl start portal64-mcp

# Check status
sudo systemctl status portal64-mcp

# View logs
sudo journalctl -u portal64-mcp -f
```

### Option 3: Docker Container
Create a Docker container for isolated deployment.

**Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-w -s" -o portal64-mcp ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/portal64-mcp .
COPY config.yaml .

CMD ["./portal64-mcp", "-config", "config.yaml"]
```

**Build and run:**
```bash
# Build image
docker build -t portal64-mcp .

# Run container
docker run -d \
  --name portal64-mcp \
  -e PORTAL64_API_URL=http://host.docker.internal:8080 \
  -v $(pwd)/config.yaml:/root/config.yaml:ro \
  portal64-mcp
```

### Option 4: Windows Service
Use NSSM (Non-Sucking Service Manager) to run as Windows service.

```cmd
# Install NSSM
# Download from https://nssm.cc/download

# Install service
nssm install Portal64MCP "C:\portal64-mcp\portal64-mcp.exe"
nssm set Portal64MCP AppDirectory "C:\portal64-mcp"
nssm set Portal64MCP AppParameters "-config config.yaml"
nssm set Portal64MCP DisplayName "Portal64 MCP Server"
nssm set Portal64MCP Description "MCP server for Portal64 chess rating system"

# Start service
nssm start Portal64MCP
```

## Directory Structure

### Recommended Production Layout
```
/opt/portal64-mcp/
├── bin/
│   └── portal64-mcp              # Binary executable
├── config.yaml                   # Configuration file
├── logs/                         # Log files (if file logging enabled)
└── docs/                         # Documentation
    ├── high-level-design.md
    ├── api-reference.md
    └── deployment.md
```

### File Permissions
```bash
# Set appropriate permissions
sudo chown -R mcp:mcp /opt/portal64-mcp
sudo chmod 755 /opt/portal64-mcp/bin/portal64-mcp
sudo chmod 644 /opt/portal64-mcp/config.yaml
sudo chmod 755 /opt/portal64-mcp/logs
```

## Health Monitoring

### Health Check Endpoint
The server doesn't expose HTTP endpoints, but you can check health via the MCP protocol or by monitoring the Portal64 API directly.

### Log Monitoring
Monitor logs for error patterns:

```bash
# View real-time logs
tail -f /var/log/syslog | grep portal64-mcp

# With systemd
journalctl -u portal64-mcp -f

# Search for errors
journalctl -u portal64-mcp | grep ERROR
```

### Process Monitoring
```bash
# Check if process is running
pgrep -f portal64-mcp

# Monitor resource usage
top -p $(pgrep portal64-mcp)

# Detailed process information
ps aux | grep portal64-mcp
```

## Security Considerations

### Network Security
- The MCP server communicates via stdio by default (no network exposure)
- Outbound connections only to Portal64 API
- No authentication required (follows Portal64 API security model)

### File System Security
- Run under dedicated user account (not root)
- Limit file system access with systemd security settings
- Use read-only configuration files
- Separate log directory with appropriate permissions

### Process Security
```bash
# Create dedicated user
sudo useradd -r -s /bin/false -d /opt/portal64-mcp mcp

# Secure the installation directory
sudo chown -R mcp:mcp /opt/portal64-mcp
sudo chmod 750 /opt/portal64-mcp
```

## Performance Tuning

### Resource Limits
Configure resource limits to prevent resource exhaustion:

```ini
# In systemd service file
[Service]
MemoryLimit=512M
CPUQuota=50%
TasksMax=100
```

### Connection Pool Tuning
The HTTP client uses connection pooling. Adjust if needed:

- `MaxIdleConns`: 100 (default)
- `MaxIdleConnsPerHost`: 10 (default)
- `IdleConnTimeout`: 90 seconds (default)

### API Timeout Configuration
Adjust API timeout based on network conditions:

```yaml
api:
  timeout: "30s"  # Increase for slow networks
```

## Troubleshooting

### Common Issues

#### 1. API Connection Failed
```bash
# Check Portal64 API availability
curl http://localhost:8080/api/v1/health

# Check network connectivity
ping localhost

# Verify configuration
./portal64-mcp -config config.yaml 2>&1 | grep -i error
```

#### 2. Permission Denied
```bash
# Check file permissions
ls -la /opt/portal64-mcp/bin/portal64-mcp

# Fix permissions
sudo chmod +x /opt/portal64-mcp/bin/portal64-mcp
```

#### 3. Configuration Errors
```bash
# Validate configuration
./portal64-mcp -config config.yaml -log-level debug

# Check environment variables
env | grep PORTAL64
```

### Debug Mode
Enable debug logging for troubleshooting:

```bash
# Command line
./portal64-mcp -log-level debug

# Environment variable
export LOG_LEVEL=debug
./portal64-mcp

# Configuration file
logging:
  level: debug
```

### Log Analysis
Common log patterns to monitor:

```bash
# Connection errors
grep "API request failed" /var/log/portal64-mcp.log

# Tool execution errors
grep "Tool execution failed" /var/log/portal64-mcp.log

# Configuration issues
grep "Invalid configuration" /var/log/portal64-mcp.log
```

## Backup and Recovery

### Configuration Backup
```bash
# Backup configuration
cp /opt/portal64-mcp/config.yaml /backup/portal64-mcp-config-$(date +%Y%m%d).yaml

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/portal64-mcp"
DATE=$(date +%Y%m%d-%H%M%S)
mkdir -p $BACKUP_DIR
cp /opt/portal64-mcp/config.yaml $BACKUP_DIR/config-$DATE.yaml
```

### Recovery Procedure
1. Stop the service
2. Restore configuration file
3. Validate configuration
4. Restart service

```bash
sudo systemctl stop portal64-mcp
sudo cp /backup/config.yaml /opt/portal64-mcp/
./portal64-mcp -config /opt/portal64-mcp/config.yaml -log-level debug
sudo systemctl start portal64-mcp
```

## Maintenance

### Regular Maintenance Tasks
- Monitor log files for errors
- Check disk space usage
- Update binary when new versions are available
- Review and rotate log files
- Verify API connectivity

### Update Procedure
1. Build new version
2. Stop service
3. Backup current binary
4. Replace binary
5. Test configuration
6. Start service

```bash
# Build new version
make build-prod

# Stop service
sudo systemctl stop portal64-mcp

# Backup and replace
sudo cp /opt/portal64-mcp/bin/portal64-mcp /opt/portal64-mcp/bin/portal64-mcp.backup
sudo cp bin/portal64-mcp /opt/portal64-mcp/bin/

# Test and restart
sudo systemctl start portal64-mcp
sudo systemctl status portal64-mcp
```

## MCP Client Integration

### Claude Desktop Configuration
Add to Claude Desktop configuration:

```json
{
  "mcpServers": {
    "portal64": {
      "command": "/opt/portal64-mcp/bin/portal64-mcp",
      "args": ["-config", "/opt/portal64-mcp/config.yaml"],
      "env": {
        "PORTAL64_API_URL": "http://localhost:8080",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### Custom MCP Client
Integrate with custom MCP clients by launching the server as a subprocess and communicating via stdio.

## Support and Documentation

### Log Locations
- systemd: `journalctl -u portal64-mcp`
- File logging: `/opt/portal64-mcp/logs/portal64-mcp.log`
- Docker: `docker logs portal64-mcp`

### Documentation
- High-level design: `docs/high-level-design.md`
- API reference: `docs/api-reference.md`
- This deployment guide: `docs/deployment.md`

### Getting Help
1. Check log files for error messages
2. Verify Portal64 API connectivity
3. Test with debug logging enabled
4. Review configuration settings
5. Consult documentation
