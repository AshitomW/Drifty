package collector

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectCertificates(ctx context.Context) (map[string]models.Certificate, error) {
	certificates := make(map[string]models.Certificate)

	if !c.config.Certificates.Enabled {
		return certificates, nil
	}

	paths := c.config.Certificates.Paths
	if len(paths) == 0 {
		paths = []string{
			"/etc/ssl/certs",
			"/etc/letsencrypt",
			"/etc/kubernetes",
			"/usr/local/share/ca-certificates",
			os.Getenv("HOME") + "/.ssh",
		}
	}

	extensions := c.config.Certificates.Extensions
	if len(extensions) == 0 {
		extensions = []string{".pem", ".crt", ".cer", ".key", ".p12", ".pfx"}
	}

	for _, path := range paths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(filePath))
			hasExt := false
			for _, e := range extensions {
				if "."+strings.TrimPrefix(ext, ".") == e || ext == e {
					hasExt = true
					break
				}
			}

			if !hasExt {
				return nil
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil
			}

			certs, err := parseCertificates(data)
			if err != nil {
				return nil
			}

			for i, cert := range certs {
				key := filePath
				if len(certs) > 1 {
					key = fmt.Sprintf("%s:%d", filePath, i)
				}

				daysToExpire := int(time.Until(cert.NotAfter).Hours() / 24)
				isExpired := time.Now().After(cert.NotAfter)
				isValid := !isExpired && time.Now().After(cert.NotBefore)

				hash := md5.Sum(data)
				fingerprint := hex.EncodeToString(hash[:])

				certificates[key] = models.Certificate{
					Path:         filePath,
					Domain:       cert.Subject.CommonName,
					Issuer:       cert.Issuer.CommonName,
					Subject:      cert.Subject.CommonName,
					NotBefore:    cert.NotBefore,
					NotAfter:     cert.NotAfter,
					SerialNumber: cert.SerialNumber.String(),
					Fingerprint:  fingerprint,
					IsValid:      isValid,
					IsExpired:    isExpired,
					DaysToExpire: daysToExpire,
				}
			}

			return nil
		})

		if err != nil {
			return certificates, err
		}
	}

	return certificates, nil
}

func parseCertificates(data []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	block, rest := pem.Decode(data)
	if block == nil {
		tlsCert, err := tls.LoadX509KeyPair(string(data), "")
		if err == nil && len(tlsCert.Certificate) > 0 {
			cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
			if err == nil {
				certs = append(certs, cert)
			}
		}
		return certs, nil
	}

	if block.Type == "CERTIFICATE" || block.Type == "TRUSTED CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err == nil {
			certs = append(certs, cert)
		}
	}

	if len(rest) > 0 {
		moreCerts, err := parseCertificates(rest)
		if err == nil {
			certs = append(certs, moreCerts...)
		}
	}

	return certs, nil
}
