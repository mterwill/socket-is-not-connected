package main

import (
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"time"
)

func newSelfSignedCert(certFile, keyFile string) error {
	key, err := rsa.GenerateKey(crand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("generating private key: %s", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return fmt.Errorf("writing private key: %s", err)
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate: %s", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		return fmt.Errorf("writing certificate: %s", err)
	}

	return nil
}

var (
	proto    = flag.String("proto", "https", "http or https")
	upstream = flag.String("upstream", "localhost:9092", "upstream server")
)

func run(ctx context.Context) error {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   *upstream,
	})
	srv := http.Server{
		Handler: proxy,
		Addr:    ":10000",
	}
	log.Printf("Starting proxy at %s://localhost:10000", *proto)
	if *proto == "https" {
		dir := os.TempDir()
		certFile, keyFile := path.Join(dir, "localhost.crt"), path.Join(dir, "localhost.key")
		if err := newSelfSignedCert(certFile, keyFile); err != nil {
			return fmt.Errorf("generating new certificate: %s", err)
		}
		return srv.ListenAndServeTLS(certFile, keyFile)
	}
	return srv.ListenAndServe()
}

func main() {
	flag.Parse()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
