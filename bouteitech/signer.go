package bouteitech

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// Signer holds the CA certificate and private key for signing.
type Signer struct {
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

// NewSigner creates a new Signer instance by loading the CA certificate and private key from the specified file paths.
func NewSigner(caCertPath, caKeyPath string) (*Signer, error) {
	tlsCert, err := tls.LoadX509KeyPair(caCertPath, caKeyPath)
	if err != nil {
		return nil, err
	}
	caCert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, err
	}
	caKey, ok := tlsCert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, err
	}
	return &Signer{
		caCert: caCert,
		caKey:  caKey,
	}, nil
}

// GetDomainCertificate generates a TLS certificate for the specified host, signed by the CA.
func (r *Signer) GetDomainCertificate(host string) (*tls.Certificate, error) {
	serial, _ := rand.Int(rand.Reader, big.NewInt(1000).Lsh(big.NewInt(1), 128))
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"MITM Dev Cert"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}
	if ip := net.ParseIP(host); ip != nil {
		tpl.IPAddresses = []net.IP{ip}
	} else {
		tpl.DNSNames = []string{host}
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, r.caCert, &key.PublicKey, r.caKey)
	if err != nil {
		return nil, err
	}
	cert := &tls.Certificate{
		Certificate: [][]byte{der, r.caCert.Raw},
		PrivateKey:  key,
	}
	return cert, nil
}
