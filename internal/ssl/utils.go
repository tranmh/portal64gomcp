// Package ssl provides SSL/TLS utilities for certificate management and secure connections
package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/svw-info/portal64gomcp/internal/logger"
)

// CertificateManager handles SSL certificate operations
type CertificateManager struct {
	logger logger.Logger
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(logger logger.Logger) *CertificateManager {
	return &CertificateManager{
		logger: logger,
	}
}

// LoadOrGenerateCertificate loads existing certificates or generates new ones
func (cm *CertificateManager) LoadOrGenerateCertificate(certFile, keyFile string, hosts []string, autoGenerate bool) (tls.Certificate, error) {
	// Try to load existing certificate first
	if cm.certificatesExist(certFile, keyFile) {
		cm.logger.WithFields(map[string]interface{}{
			"cert_file": certFile,
			"key_file":  keyFile,
		}).Info("Loading existing SSL certificates")
		
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			if !autoGenerate {
				return tls.Certificate{}, fmt.Errorf("failed to load certificate: %w", err)
			}
			cm.logger.WithError(err).Warn("Failed to load existing certificates, generating new ones")
		} else {
			// Validate certificate
			if err := cm.validateCertificate(cert, hosts); err != nil {
				if !autoGenerate {
					return tls.Certificate{}, fmt.Errorf("certificate validation failed: %w", err)
				}
				cm.logger.WithError(err).Warn("Certificate validation failed, generating new ones")
			} else {
				return cert, nil
			}
		}
	}

	// Generate new certificate if needed
	if autoGenerate {
		cm.logger.WithFields(map[string]interface{}{
			"cert_file": certFile,
			"key_file":  keyFile,
			"hosts":     hosts,
		}).Info("Generating new self-signed SSL certificates")
		
		return cm.generateSelfSignedCertificate(certFile, keyFile, hosts)
	}

	return tls.Certificate{}, fmt.Errorf("certificates not found and auto-generation is disabled")
}

// certificatesExist checks if certificate files exist
func (cm *CertificateManager) certificatesExist(certFile, keyFile string) bool {
	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)
	return certErr == nil && keyErr == nil
}

// validateCertificate validates that the certificate is suitable for the given hosts
func (cm *CertificateManager) validateCertificate(cert tls.Certificate, hosts []string) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("no certificate found in certificate chain")
	}

	// Parse the certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check expiration
	now := time.Now()
	if now.Before(x509Cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}
	if now.After(x509Cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	// Warn if certificate expires soon
	if x509Cert.NotAfter.Sub(now) < 30*24*time.Hour {
		cm.logger.WithField("expires", x509Cert.NotAfter).Warn("Certificate expires soon")
	}

	// Check that certificate covers required hosts
	if len(hosts) > 0 {
		for _, host := range hosts {
			ip := net.ParseIP(host)
			if ip != nil {
				// Check IP SANs
				found := false
				for _, certIP := range x509Cert.IPAddresses {
					if certIP.Equal(ip) {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("certificate does not cover IP address: %s", host)
				}
			} else {
				// Check DNS SANs and CN
				if err := x509Cert.VerifyHostname(host); err != nil {
					return fmt.Errorf("certificate does not cover hostname %s: %w", host, err)
				}
			}
		}
	}

	cm.logger.WithFields(map[string]interface{}{
		"subject":    x509Cert.Subject,
		"expires":    x509Cert.NotAfter,
		"san_dns":    x509Cert.DNSNames,
		"san_ips":    x509Cert.IPAddresses,
	}).Info("Certificate validation successful")

	return nil
}

// generateSelfSignedCertificate generates a new self-signed certificate
func (cm *CertificateManager) generateSelfSignedCertificate(certFile, keyFile string, hosts []string) (tls.Certificate, error) {
	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(certFile), 0755); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate directory: %w", err)
	}
	if filepath.Dir(keyFile) != filepath.Dir(certFile) {
		if err := os.MkdirAll(filepath.Dir(keyFile), 0700); err != nil {
			return tls.Certificate{}, fmt.Errorf("failed to create key directory: %w", err)
		}
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Portal64 MCP Server"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add Subject Alternative Names
	for _, host := range hosts {
		ip := net.ParseIP(host)
		if ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	// If no hosts specified, add localhost
	if len(hosts) == 0 {
		template.DNSNames = append(template.DNSNames, "localhost")
		template.IPAddresses = append(template.IPAddresses, net.IPv4(127, 0, 0, 1))
		template.IPAddresses = append(template.IPAddresses, net.IPv6loopback)
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Save certificate to file
	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to write certificate: %w", err)
	}

	// Save private key to file
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to write private key: %w", err)
	}

	cm.logger.WithFields(map[string]interface{}{
		"cert_file": certFile,
		"key_file":  keyFile,
		"dns_names": template.DNSNames,
		"ip_addresses": template.IPAddresses,
		"expires": template.NotAfter,
	}).Info("Generated new self-signed certificate")

	// Load the certificate we just created
	return tls.LoadX509KeyPair(certFile, keyFile)
}

// LoadCACertificates loads CA certificates from file
func LoadCACertificates(caFile string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}
	
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	
	return caCertPool, nil
}

// CreateClientTLSConfig creates a TLS configuration for HTTP clients
func CreateClientTLSConfig(verify bool, caFile, clientCert, clientKey string, insecureSkipVerify bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Load CA certificates if specified
	if caFile != "" {
		caCertPool, err := LoadCACertificates(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA certificates: %w", err)
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate if specified (for mTLS)
	if clientCert != "" && clientKey != "" {
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// SecureHeaders returns a map of security headers
func SecureHeaders(hstsMaxAge int64) map[string]string {
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	if hstsMaxAge > 0 {
		headers["Strict-Transport-Security"] = fmt.Sprintf("max-age=%d; includeSubDomains", hstsMaxAge)
	}

	return headers
}
