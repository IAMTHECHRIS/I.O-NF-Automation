package nfedist

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mschunke/gonfe/certificado"
	"github.com/mschunke/gonfe/evento"
	"github.com/mschunke/gonfe/nfe"
	"github.com/mschunke/gonfe/sefaz"
	"github.com/mschunke/gonfe/uf"
)

// ResultadoManifestacao traz somente os dados necessários para registrar o
// comprovante local e informar o operador. A Ciência da Operação é uma ação
// fiscal real e nunca é chamada automaticamente pela coleta periódica.
type ResultadoManifestacao struct {
	CStat      string
	XMotivo    string
	Protocolo  string
	Processado []byte
}

// ManifestarCiencia registra o evento 210210 para uma NF-e recebida.
// A manifestação do destinatário é sempre enviada ao Ambiente Nacional, mesmo
// quando a nota foi emitida em outra UF. O certificado precisa pertencer ao
// CNPJ-base do destinatário configurado.
func ManifestarCiencia(ctx context.Context, caminhoPFX, senhaPFX, tpAmb, cnpj, cUFAutor, chave string) (ResultadoManifestacao, error) {
	if len(strings.TrimSpace(chave)) != 44 {
		return ResultadoManifestacao{}, fmt.Errorf("a chave de acesso precisa ter 44 dígitos")
	}
	for _, r := range chave {
		if r < '0' || r > '9' {
			return ResultadoManifestacao{}, fmt.Errorf("a chave de acesso só pode conter dígitos")
		}
	}

	codigoUF, err := strconv.Atoi(strings.TrimSpace(cUFAutor))
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("código IBGE da UF inválido %q: %w", cUFAutor, err)
	}
	unidade, err := uf.PorCodigo(codigoUF)
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("UF do autor inválida: %w", err)
	}

	cert, err := certificado.CarregarArquivo(caminhoPFX, senhaPFX)
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("carregar certificado: %w", err)
	}
	ambiente := nfe.Producao
	if tpAmb == "2" {
		ambiente = nfe.Homologacao
	}

	ev, err := evento.NovaManifestacao(evento.DadosManifestacao{
		Chave: chave, CNPJ: cnpj, Ambiente: ambiente,
		Tipo: evento.TipoCienciaOperacao,
	})
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("montar Ciência da Operação: %w", err)
	}
	assinado, err := ev.AssinarCom(cert)
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("assinar Ciência da Operação: %w", err)
	}

	cliente, err := sefaz.NovoCliente(sefaz.Config{
		UF: unidade, Ambiente: ambiente, Modelo: nfe.ModeloNFe, Certificado: cert,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("criar cliente de eventos: %w", err)
	}
	ret, err := cliente.EnviarEvento(ctx, assinado)
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("enviar Ciência da Operação: %w", err)
	}
	if !ret.Registrado() {
		return ResultadoManifestacao{CStat: strconv.Itoa(ret.InfEvento.CStat), XMotivo: ret.InfEvento.XMotivo},
			fmt.Errorf("SEFAZ não registrou a Ciência da Operação (cStat=%d): %s", ret.InfEvento.CStat, ret.InfEvento.XMotivo)
	}
	processado, err := evento.MontarProcEvento(assinado, ret)
	if err != nil {
		return ResultadoManifestacao{}, fmt.Errorf("montar comprovante do evento: %w", err)
	}
	return ResultadoManifestacao{
		CStat: strconv.Itoa(ret.InfEvento.CStat), XMotivo: ret.InfEvento.XMotivo,
		Protocolo: ret.InfEvento.NProt, Processado: nfe.XMLDeclarado(processado),
	}, nil
}
