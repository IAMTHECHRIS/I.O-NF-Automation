package adn

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func erroComunicacaoTLS(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls:") ||
		strings.Contains(msg, "handshake") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "wsarecv") ||
		strings.Contains(msg, "forcibly closed")
}

func (c *Client) getComFallback(endpoint string) ([]byte, int, error) {
	resp, err := c.HTTPClient.Get(endpoint)
	if err == nil {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, resp.StatusCode, fmt.Errorf("ler resposta ADN: %w", readErr)
		}
		return body, resp.StatusCode, nil
	}
	if !erroComunicacaoTLS(err) {
		return nil, 0, fmt.Errorf("chamar ADN falhou na comunicação TLS/rede (sem retry automático): %w", err)
	}

	if runtime.GOOS == "windows" && c.pfxPath != "" {
		body, status, psErr := c.getViaPowerShell(endpoint)
		if psErr == nil {
			return body, status, nil
		}

		body, status, pyErr := c.getViaPython(endpoint)
		if pyErr == nil {
			return body, status, nil
		}

		body, status, fallbackErr := c.getViaCurl(endpoint)
		if fallbackErr != nil {
			return nil, 0, fmt.Errorf("chamar ADN falhou no Go (%v), no fallback PowerShell/.NET (%v), no fallback Python/OpenSSL (%v) e no fallback curl/OpenSSL (%w)", err, psErr, pyErr, fallbackErr)
		}
		return body, status, nil
	}

	body, status, pyErr := c.getViaPython(endpoint)
	if pyErr == nil {
		return body, status, nil
	}

	body, status, fallbackErr := c.getViaCurl(endpoint)
	if fallbackErr != nil {
		return nil, 0, fmt.Errorf("chamar ADN falhou no Go (%v), no fallback Python/OpenSSL (%w) e no fallback curl/OpenSSL (%w)", err, pyErr, fallbackErr)
	}
	return body, status, nil
}

// getViaPowerShell contorna o "tls: bad record MAC" do crypto/tls puro do Go
// no Windows, usando o handshake TLS nativo do .NET (Schannel) via
// X509Certificate2 + HttpWebRequest. Confirmado manualmente contra o ADN de
// produção em 2026-09-08: o mesmo PFX que falha no Go funciona sem erro pelo
// PowerShell 5.1 já presente em qualquer Windows, sem exigir Python/OpenSSL.
func (c *Client) getViaPowerShell(endpoint string) ([]byte, int, error) {
	powershellPath, err := exec.LookPath("powershell")
	if err != nil {
		return nil, 0, errors.New("powershell não encontrado no sistema")
	}

	dir, err := os.MkdirTemp("", "io-nf-adn-ps-*")
	if err != nil {
		return nil, 0, fmt.Errorf("criar temporário: %w", err)
	}
	defer os.RemoveAll(dir)

	scriptPath := filepath.Join(dir, "adn.ps1")
	script := `$ErrorActionPreference = "Stop"
$pfxPath = $args[0]
$pfxPass = $args[1]
$url = $args[2]

$flags = [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::Exportable -bor [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::MachineKeySet
$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $pfxPass, $flags)

[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

try {
    $req = [System.Net.HttpWebRequest]::Create($url)
    $req.ClientCertificates.Add($cert) | Out-Null
    $req.Method = "GET"
    $req.Timeout = 30000
    $resp = $req.GetResponse()
    $status = [int]$resp.StatusCode
    $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
    $body = $reader.ReadToEnd()
    $resp.Close()
    [Console]::Out.Write($body)
    [Console]::Out.Write("` + "\nIO_NF_HTTP_STATUS:" + `" + $status)
} catch [System.Net.WebException] {
    if ($_.Exception.Response) {
        $errResp = $_.Exception.Response
        $status = [int]$errResp.StatusCode
        $reader = New-Object System.IO.StreamReader($errResp.GetResponseStream())
        $body = $reader.ReadToEnd()
        [Console]::Out.Write($body)
        [Console]::Out.Write("` + "\nIO_NF_HTTP_STATUS:" + `" + $status)
    } else {
        [Console]::Error.Write($_.Exception.Message)
        exit 1
    }
} catch {
    [Console]::Error.Write($_.Exception.Message)
    exit 1
}
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return nil, 0, fmt.Errorf("gravar script PowerShell temporário: %w", err)
	}

	cmd := exec.Command(powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath, c.pfxPath, c.pfxSenha, endpoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, 0, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return separarStatusCurl(out)
}

func (c *Client) getViaPython(endpoint string) ([]byte, int, error) {
	pythonPath, err := procurarPython()
	if err != nil {
		return nil, 0, err
	}

	dir, err := os.MkdirTemp("", "io-nf-adn-py-*")
	if err != nil {
		return nil, 0, fmt.Errorf("criar temporário: %w", err)
	}
	defer os.RemoveAll(dir)

	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := escreverPEM(c.cert, certPath, keyPath); err != nil {
		return nil, 0, err
	}

	script := `
import ssl
import sys
import urllib.request
import urllib.error

url, cert_path, key_path = sys.argv[1], sys.argv[2], sys.argv[3]
ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE
ctx.load_cert_chain(certfile=cert_path, keyfile=key_path)
req = urllib.request.Request(url, method="GET")
try:
    with urllib.request.urlopen(req, context=ctx, timeout=60) as resp:
        body = resp.read()
        status = resp.getcode()
except urllib.error.HTTPError as e:
    body = e.read()
    status = e.code
sys.stdout.buffer.write(body)
sys.stdout.buffer.write(b"\nIO_NF_HTTP_STATUS:" + str(status).encode("ascii"))
`

	cmd := exec.Command(pythonPath, "-c", script, endpoint, certPath, keyPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, 0, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return separarStatusCurl(out)
}

func procurarPython() (string, error) {
	candidatos := []string{"python", "python3"}
	for _, nome := range candidatos {
		if path, err := exec.LookPath(nome); err == nil {
			return path, nil
		}
	}
	if path, err := exec.LookPath("py"); err == nil {
		return path, nil
	}
	return "", errors.New("Python não encontrado no sistema")
}

func (c *Client) getViaCurl(endpoint string) ([]byte, int, error) {
	curlPath, err := exec.LookPath("curl")
	if err != nil {
		return nil, 0, errors.New("curl não encontrado no sistema")
	}

	dir, err := os.MkdirTemp("", "io-nf-adn-*")
	if err != nil {
		return nil, 0, fmt.Errorf("criar temporário: %w", err)
	}
	defer os.RemoveAll(dir)

	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := escreverPEM(c.cert, certPath, keyPath); err != nil {
		return nil, 0, err
	}

	args := []string{
		"--http1.1",
		"--silent",
		"--show-error",
		"--connect-timeout", "20",
		"--max-time", "60",
		"--insecure",
		"--cert", certPath,
		"--key", keyPath,
		"--write-out", "\nIO_NF_HTTP_STATUS:%{http_code}",
		endpoint,
	}
	cmd := exec.Command(curlPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if c.pfxPath != "" {
			body, status, pfxErr := c.getViaCurlPFX(curlPath, endpoint)
			if pfxErr == nil {
				return body, status, nil
			}
			return nil, 0, fmt.Errorf("curl PEM falhou (%v: %s); curl PFX falhou (%w)", err, strings.TrimSpace(string(out)), pfxErr)
		}
		return nil, 0, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}

	return separarStatusCurl(out)
}

func (c *Client) getViaCurlPFX(curlPath, endpoint string) ([]byte, int, error) {
	args := []string{
		"--http1.1",
		"--silent",
		"--show-error",
		"--connect-timeout", "20",
		"--max-time", "60",
		"--insecure",
		"--cert-type", "P12",
		"--cert", c.pfxPath + ":" + c.pfxSenha,
		"--write-out", "\nIO_NF_HTTP_STATUS:%{http_code}",
		endpoint,
	}
	cmd := exec.Command(curlPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, 0, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return separarStatusCurl(out)
}

func separarStatusCurl(out []byte) ([]byte, int, error) {
	marcador := []byte("\nIO_NF_HTTP_STATUS:")
	idx := bytes.LastIndex(out, marcador)
	if idx < 0 {
		return nil, 0, fmt.Errorf("curl não retornou marcador HTTP: %s", strings.TrimSpace(string(out)))
	}
	body := out[:idx]
	statusText := strings.TrimSpace(string(out[idx+len(marcador):]))
	var status int
	if _, err := fmt.Sscanf(statusText, "%d", &status); err != nil {
		return nil, 0, fmt.Errorf("status HTTP inválido do curl %q: %w", statusText, err)
	}
	return body, status, nil
}

func escreverPEM(cert tls.Certificate, certPath, keyPath string) error {
	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("criar cert.pem temporário: %w", err)
	}
	for _, der := range cert.Certificate {
		if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
			certFile.Close()
			return fmt.Errorf("gravar cert.pem temporário: %w", err)
		}
	}
	if err := certFile.Close(); err != nil {
		return fmt.Errorf("fechar cert.pem temporário: %w", err)
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		if signer, ok := cert.PrivateKey.(crypto.Signer); ok {
			keyBytes, err = x509.MarshalPKCS8PrivateKey(signer)
		}
	}
	if err != nil {
		return fmt.Errorf("converter chave privada para PEM: %w", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("criar key.pem temporário: %w", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}); err != nil {
		keyFile.Close()
		return fmt.Errorf("gravar key.pem temporário: %w", err)
	}
	if err := keyFile.Close(); err != nil {
		return fmt.Errorf("fechar key.pem temporário: %w", err)
	}

	_ = os.Chmod(certPath, 0o600)
	_ = os.Chmod(keyPath, 0o600)
	return nil
}
