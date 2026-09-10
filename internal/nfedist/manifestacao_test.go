package nfedist

import (
	"context"
	"strings"
	"testing"
)

func TestManifestarCienciaRejeitaChaveInvalidaAntesDeLerCertificado(t *testing.T) {
	_, err := ManifestarCiencia(context.Background(), "/arquivo/que-nao-deve-ser-lido.pfx", "", "1", "123", "35", "invalida")
	if err == nil || !strings.Contains(err.Error(), "44 dígitos") {
		t.Fatalf("erro = %v; queria validação da chave", err)
	}
}

func TestManifestarCienciaRejeitaUFInvalidaAntesDeLerCertificado(t *testing.T) {
	chave := strings.Repeat("0", 44)
	_, err := ManifestarCiencia(context.Background(), "/arquivo/que-nao-deve-ser-lido.pfx", "", "1", "123", "999", chave)
	if err == nil || !strings.Contains(err.Error(), "UF do autor inválida") {
		t.Fatalf("erro = %v; queria validação da UF", err)
	}
}
