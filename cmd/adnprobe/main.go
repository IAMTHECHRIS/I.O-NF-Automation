package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"io-nf-automation/internal/appconfig"
	"io-nf-automation/internal/certload"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		fmt.Printf("CONFIG_ERRO: %v\n", err)
		os.Exit(1)
	}
	cert, err := certload.FromPFXValidado(cfg.CertificadoPfx, cfg.CertificadoSenha)
	if err != nil {
		fmt.Printf("CERT_ERRO: %v\n", err)
		os.Exit(1)
	}
	endpoint := "https://adn.nfse.gov.br/contribuintes/DFe/339"
	q := url.Values{}
	q.Set("tipoNSU", "DISTRIBUICAO")
	q.Set("lote", "true")
	endpoint += "?" + q.Encode()

	tests := []struct {
		name string
		cfg  *tls.Config
	}{
		{"default-min12", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}},
		{"tls12-only", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}},
		{"tls13-only", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13}},
		{"tls13-no-hybrid", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13, CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521}}},
		{"min12-no-hybrid", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12, CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521}}},
		{"tls12-servername", &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12, ServerName: "adn.nfse.gov.br"}},
	}

	failed := false
	for _, tt := range tests {
		tr := &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			ForceAttemptHTTP2: false,
			TLSNextProto:      map[string]func(string, *tls.Conn) http.RoundTripper{},
			TLSClientConfig:   tt.cfg,
		}
		client := &http.Client{Timeout: 45 * time.Second, Transport: tr}
		resp, err := client.Get(endpoint)
		if err != nil {
			fmt.Printf("%s ERROR %T %v\n", tt.name, err, err)
			failed = true
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		vers := 0
		cipher := 0
		if resp.TLS != nil {
			vers = int(resp.TLS.Version)
			cipher = int(resp.TLS.CipherSuite)
		}
		fmt.Printf("%s OK status=%d tls=0x%x cipher=0x%x body=%q\n", tt.name, resp.StatusCode, vers, cipher, string(body))
	}
	if failed {
		os.Exit(1)
	}
}
