# SSL Implementation Design for Portal64 MCP Server

## Overview

This document outlines the comprehensive SSL/TLS implementation for the Portal64 MCP (Model Context Protocol) server, providing end-to-end encryption for both inbound MCP connections and outbound API calls.

## Current Architecture

```
┌─────────────────┐    HTTP    ┌─────────────────┐    HTTP    ┌─────────────────┐
│   MCP Client    │ ─────────► │   MCP Server    │ ─────────► │  Portal64 API   │
│                 │            │   (Port 8888)   │            │  (Port 8080)    │
└─────────────────┘            └─────────────────┘            └─────────────────┘
```

## Target Architecture

```
┌─────────────────┐   HTTPS    ┌─────────────────┐   HTTPS    ┌─────────────────┐
│   MCP Client    │ ─────────► │   MCP Server    │ ─────────► │  Portal64 API   │
│                 │    TLS     │   (Port 8888)   │    TLS     │  (Port 8443)    │
│  (w/ Client     │ ◄───────── │  (w/ Server     │ ◄───────── │  (w/ Server     │
│   Certs)        │   mTLS     │   Certs)        │   mTLS     │   Certs)        │
└─────────────────┘            └─────────────────┘            └─────────────────┘
```

## Design Decisions

### 1. SSL Scope: Full Stack (B)
- **MCP Server SSL**: HTTPS endpoints with TLS termination
- **API Client SSL**: HTTPS connections to Portal64 API
- **End-to-End Encryption**: Complete security coverage

### 2. Certificate Management: Hybrid Approach (D)
- **File-based certificates**: Production-ready cert/key files
- **Auto-generated certificates**: Self-signed fallback for development
- **Certificate validation**: Chain verification and expiration checks
- **Hot reload capability**: Certificate rotation without restart

### 3. Mutual TLS: Optional (B)
- **Server-only TLS**: Default HTTPS operation
- **Optional mTLS**: Client certificates when configured
- **Flexible authentication**: Works with or without client certs

### 4. Migration Strategy: Breaking Change (A)
- **SSL required**: When enabled, only HTTPS connections accepted
- **Configuration-driven**: SSL can be disabled for development
- **Clear security boundaries**: No HTTP fallbacks in production

### 5. Environment-Specific Configuration (B)
- **Development**: SSL disabled by default (`ssl.enabled: false`)
- **Production**: SSL enabled by default (`ssl.enabled: true`)
- **Auto-detection**: Environment variable overrides

### 6. Security Standards: Best Practices
- **TLS 1.3 preferred**: Modern encryption protocols
- **TLS 1.2 minimum**: Backward compatibility
- **Strong cipher suites**: AEAD ciphers with forward secrecy
- **HSTS headers**: HTTP Strict Transport Security
- **Security headers**: Comprehensive protection

## Architecture Components

### 1. Configuration Layer

```yaml
api:
  base_url: "https://localhost:8443"
  ssl:
    verify: true
    ca_file: ""
    client_cert: ""
    client_key: ""
    insecure_skip_verify: false

mcp:
  ssl:
    enabled: true
    cert_file: "certs/server.crt"
    key_file: "certs/server.key"
    ca_file: ""
    min_version: "1.2"
    max_version: "1.3"
    require_client_cert: false
    auto_generate_certs: true
    auto_cert_hosts: ["localhost", "127.0.0.1"]

development:
  mcp:
    ssl:
      enabled: false
```

### 2. SSL Utilities Package (`internal/ssl/`)

#### Certificate Manager
- **Certificate loading**: File-based cert/key pairs
- **Auto-generation**: Self-signed certificates for development
- **Validation**: Certificate chain and expiration checking
- **SAN support**: Subject Alternative Names for multiple hosts

#### TLS Configuration
- **Client TLS config**: For API client connections
- **Server TLS config**: For MCP server endpoints
- **Security headers**: HSTS, X-Frame-Options, etc.

### 3. Enhanced MCP Server (`internal/mcp/server.go`)

#### HTTPS Server
- **TLS termination**: SSL/TLS endpoint handling
- **Certificate management**: Automatic loading and validation
- **Graceful shutdown**: Proper connection cleanup
- **Performance monitoring**: SSL connection metrics

#### Security Middleware
- **Client certificate validation**: mTLS support
- **Security headers**: Comprehensive protection
- **SSL logging**: Connection details and certificate info
- **CORS handling**: Environment-specific policies

### 4. SSL-Enhanced API Client (`internal/api/client.go`)

#### HTTPS Client
- **Certificate verification**: Server certificate validation
- **Client certificates**: mTLS authentication
- **Custom CA support**: Private certificate authorities
- **Connection pooling**: Efficient SSL connection reuse

#### Error Handling
- **SSL diagnostics**: TLS-specific error reporting
- **Certificate warnings**: Expiration notifications
- **Fallback logic**: Graceful degradation options

## Security Features

### Transport Layer Security
- **Encryption**: AES-256-GCM with ECDHE key exchange
- **Authentication**: RSA-2048 or ECDSA-P256 certificates
- **Integrity**: SHA-256 message authentication
- **Forward Secrecy**: Ephemeral key exchange

### Certificate Management
- **Certificate rotation**: Hot reload without downtime
- **Expiration monitoring**: Automated renewal warnings
- **Chain validation**: Complete certificate path verification
- **Revocation checking**: OCSP stapling support (future)

### Security Headers
```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

### Client Authentication (mTLS)
- **Optional authentication**: Configurable client certificate requirements
- **Certificate validation**: Client cert verification against CA
- **Authorization mapping**: Client cert to permissions (future enhancement)

## Configuration Schema

### SSL Configuration Structure
```go
type MCPSSLConfig struct {
    Enabled              bool     `mapstructure:"enabled"`
    CertFile             string   `mapstructure:"cert_file"`
    KeyFile              string   `mapstructure:"key_file"`
    CAFile               string   `mapstructure:"ca_file"`
    MinVersion           string   `mapstructure:"min_version"`
    MaxVersion           string   `mapstructure:"max_version"`
    CipherSuites         []string `mapstructure:"cipher_suites"`
    RequireClientCert    bool     `mapstructure:"require_client_cert"`
    HSTSMaxAge           int64    `mapstructure:"hsts_max_age"`
    AutoGenerateCerts    bool     `mapstructure:"auto_generate_certs"`
    AutoCertHosts        []string `mapstructure:"auto_cert_hosts"`
}
```

### Environment Variables
```bash
# MCP Server SSL
MCP_SSL_ENABLED=true
MCP_SSL_CERT_FILE=/path/to/cert.pem
MCP_SSL_KEY_FILE=/path/to/key.pem

# API Client SSL
API_SSL_VERIFY=true
API_SSL_CA_FILE=/path/to/ca.pem
API_SSL_CLIENT_CERT=/path/to/client.pem
API_SSL_CLIENT_KEY=/path/to/client-key.pem
```

## Implementation Phases

### Phase 1: Configuration Enhancement ✓
- Enhanced configuration structs with SSL support
- Environment-specific overrides
- Validation and defaults

### Phase 2: SSL Utilities Package ✓
- Certificate management utilities
- TLS configuration builders
- Security header generators

### Phase 3: Server SSL Implementation ✓
- HTTPS server setup
- Certificate loading/generation
- Security middleware integration

### Phase 4: Client SSL Enhancement ✓
- HTTPS client configuration
- Certificate verification
- mTLS support

### Phase 5: Testing & Documentation ✓
- SSL-specific test cases
- Certificate generation scripts
- Operational documentation

## Development Workflow

### Local Development
```bash
# Generate development certificates
make ssl-certs

# Run with SSL disabled (development mode)
make dev

# Run with SSL enabled for testing
make dev-ssl

# Test SSL connection
make test-ssl
```

### Production Deployment
```bash
# Build production binary
make build

# Run with SSL enabled (production mode)
ENV=production ./bin/portal64-mcp

# Verify SSL configuration
curl -k https://localhost:8888/api/v1/ssl/info
```

## Certificate Management

### Development Certificates
- **Auto-generation**: Self-signed certificates for localhost
- **SAN support**: Multiple hostnames and IP addresses
- **Limited lifetime**: 365 days for security
- **Easy regeneration**: `make ssl-certs` command

### Production Certificates
- **External CA**: Let's Encrypt, internal CA, or commercial provider
- **Certificate monitoring**: Expiration tracking and alerts
- **Automated renewal**: Integration with ACME protocol (future)
- **Backup procedures**: Certificate and key backup strategies

### Certificate Locations
```
certs/
├── server.crt          # Server certificate
├── server.key          # Server private key
├── client.crt          # Client certificate (mTLS)
├── client.key          # Client private key (mTLS)
└── ca.crt              # Certificate Authority (optional)
```

## Monitoring & Logging

### SSL Connection Logging
```json
{
  "level": "info",
  "msg": "HTTPS request completed",
  "method": "POST",
  "path": "/tools/call",
  "status": 200,
  "duration_ms": 45,
  "tls_version": "1.3",
  "cipher_suite": "TLS_AES_256_GCM_SHA384",
  "client_cert_present": true,
  "client_cert_cn": "portal64-client"
}
```

### Certificate Monitoring
- **Expiration warnings**: 30-day advance notice
- **Certificate validation**: Chain verification status
- **Connection metrics**: SSL handshake performance
- **Error tracking**: TLS-specific error categorization

## Performance Considerations

### SSL Overhead
- **Handshake cost**: ~2-4ms additional latency
- **Encryption overhead**: ~5-10% CPU usage increase
- **Memory usage**: ~1KB per connection for SSL context
- **Connection pooling**: Amortize handshake cost

### Optimization Strategies
- **HTTP/2 support**: Multiplexed connections
- **Session resumption**: TLS session tickets
- **OCSP stapling**: Reduced certificate validation overhead
- **Hardware acceleration**: AES-NI instruction support

## Security Considerations

### Threat Model
- **Data in transit**: Protection against eavesdropping
- **Man-in-the-middle**: Certificate-based authentication
- **Certificate tampering**: Chain of trust validation
- **Downgrade attacks**: Minimum TLS version enforcement

### Security Best Practices
- **Regular updates**: Keep TLS libraries current
- **Certificate rotation**: Periodic key renewal
- **Strong random numbers**: Proper entropy sources
- **Side-channel protection**: Constant-time operations

## Future Enhancements

### Certificate Automation
- **ACME integration**: Automated Let's Encrypt certificates
- **Certificate rotation**: Zero-downtime certificate updates
- **Multi-certificate support**: SNI-based certificate selection

### Advanced Features
- **Certificate pinning**: Enhanced security for known endpoints
- **OCSP stapling**: Online certificate status protocol
- **Client certificate authentication**: Authorization integration
- **HSM integration**: Hardware security module support

## Testing Strategy

### Unit Tests
- Certificate generation and validation
- TLS configuration building
- Security header generation

### Integration Tests
- HTTPS server startup and shutdown
- Client certificate authentication
- Certificate expiration handling

### End-to-End Tests
- Complete SSL handshake verification
- mTLS authentication flow
- Certificate rotation scenarios

### Security Tests
- TLS configuration validation
- Certificate chain verification
- Cipher suite negotiation

## Rollback Strategy

### Graceful Degradation
- Configuration flag to disable SSL
- Environment variable overrides
- Backward compatibility maintenance

### Emergency Procedures
- SSL disable command-line flag
- Certificate bypass options (emergency only)
- Monitoring and alerting for SSL failures

## Compliance & Standards

### Standards Compliance
- **TLS 1.2/1.3**: RFC 5246 / RFC 8446 compliance
- **X.509 certificates**: RFC 5280 certificate format
- **HTTP/2**: RFC 7540 with TLS extension
- **Security headers**: OWASP recommendations

### Best Practices
- **Mozilla SSL Configuration**: Modern configuration profile
- **NIST guidelines**: SP 800-52 Rev. 2 recommendations
- **Industry standards**: PCI DSS, SOX compliance ready

---

**Document Version**: 1.0  
**Last Updated**: August 10, 2025  
**Author**: Portal64 Development Team  
**Review Status**: Ready for Implementation
