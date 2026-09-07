package danfse

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type nfseNacional struct {
	InfNFSe struct {
		NNFSe       string `xml:"nNFSe"`
		ChaveAcesso string `xml:"chNFSe"`
		DhProc      string `xml:"dhProc"`
		Emit        pessoa `xml:"emit"`
		Valores     struct {
			VLiq string `xml:"vLiq"`
		} `xml:"valores"`
		DPS struct {
			InfDPS struct {
				DCompet string `xml:"dCompet"`
				Toma    pessoa `xml:"toma"`
				Serv    struct {
					CServ struct {
						CTribNac  string `xml:"cTribNac"`
						XDescServ string `xml:"xDescServ"`
					} `xml:"cServ"`
				} `xml:"serv"`
			} `xml:"infDPS"`
		} `xml:"DPS"`
	} `xml:"infNFSe"`
}

type pessoa struct {
	CNPJ  string `xml:"CNPJ"`
	CPF   string `xml:"CPF"`
	XNome string `xml:"xNome"`
	End   struct {
		EndNac struct {
			CMun  string `xml:"cMun"`
			CEP   string `xml:"CEP"`
			UF    string `xml:"UF"`
			XMun  string `xml:"xMun"`
			XLgr  string `xml:"xLgr"`
			Nro   string `xml:"nro"`
			XBair string `xml:"xBairro"`
		} `xml:"endNac"`
	} `xml:"end"`
}

func GerarDeXML(xmlBytes []byte, direcao, gerador string) ([]byte, error) {
	var nfse nfseNacional
	if err := xml.Unmarshal(xmlBytes, &nfse); err != nil {
		return nil, fmt.Errorf("parse NFSe para PDF: %w", err)
	}
	var buf bytes.Buffer
	if err := Gerar(&buf, nfse, direcao, gerador); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Gerar(w io.Writer, nfse nfseNacional, direcao, gerador string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	cell := func(w, h float64, txt string, border string, ln int, align string, fill bool) {
		pdf.CellFormat(w, h, tr(txt), border, ln, align, fill, 0, "")
	}
	multi := func(w, h float64, txt string, border string) {
		pdf.MultiCell(w, h, tr(txt), border, "L", false)
	}

	pdf.SetFillColor(37, 41, 45)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	cell(0, 12, "DANFSe - Documento Auxiliar da NFS-e", "1", 1, "C", true)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)
	cell(0, 7, "Representação simplificada gerada a partir do XML oficial da NFS-e Nacional.", "LRB", 1, "C", false)
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 11)
	cell(0, 8, "Identificação", "1", 1, "L", false)
	pdf.SetFont("Arial", "", 10)
	linha2(pdf, tr, "Número", nfse.InfNFSe.NNFSe, "Direção", strings.ToUpper(direcao))
	linha2(pdf, tr, "Competência", dataBR(nfse.InfNFSe.DPS.InfDPS.DCompet), "Processamento", dataHoraBR(nfse.InfNFSe.DhProc))
	linha2(pdf, tr, "Valor líquido", "R$ "+valorBR(nfse.InfNFSe.Valores.VLiq), "Chave", valorOuTraco(nfse.InfNFSe.ChaveAcesso))
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	cell(0, 8, "Prestador", "1", 1, "L", false)
	pdf.SetFont("Arial", "", 10)
	multi(0, 6, nfse.InfNFSe.Emit.XNome, "LR")
	multi(0, 6, "Documento: "+doc(nfse.InfNFSe.Emit), "LR")
	if end := endereco(nfse.InfNFSe.Emit); end != "" {
		multi(0, 6, "Endereço: "+end, "LRB")
	} else {
		cell(0, 6, "", "LRB", 1, "L", false)
	}
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	cell(0, 8, "Tomador", "1", 1, "L", false)
	pdf.SetFont("Arial", "", 10)
	multi(0, 6, nfse.InfNFSe.DPS.InfDPS.Toma.XNome, "LR")
	multi(0, 6, "Documento: "+doc(nfse.InfNFSe.DPS.InfDPS.Toma), "LR")
	if end := endereco(nfse.InfNFSe.DPS.InfDPS.Toma); end != "" {
		multi(0, 6, "Endereço: "+end, "LRB")
	} else {
		cell(0, 6, "", "LRB", 1, "L", false)
	}
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	cell(0, 8, "Serviço", "1", 1, "L", false)
	pdf.SetFont("Arial", "", 10)
	descricao := strings.TrimSpace(nfse.InfNFSe.DPS.InfDPS.Serv.CServ.XDescServ)
	if descricao == "" {
		descricao = "Sem descrição no XML."
	}
	multi(0, 6, descricao, "LRB")

	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	cell(0, 5, "Gerado por "+gerador+" em "+time.Now().Format("02/01/2006 15:04"), "", 1, "R", false)

	return pdf.Output(w)
}

func linha2(pdf *gofpdf.Fpdf, tr func(string) string, a, b, c, d string) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(25, 7, tr(a+":"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(65, 7, tr(valorOuTraco(b)), "1", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(25, 7, tr(c+":"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 7, tr(valorOuTraco(d)), "1", 1, "L", false, 0, "")
}

func doc(p pessoa) string {
	if strings.TrimSpace(p.CNPJ) != "" {
		return p.CNPJ
	}
	if strings.TrimSpace(p.CPF) != "" {
		return p.CPF
	}
	return "-"
}

func endereco(p pessoa) string {
	e := p.End.EndNac
	partes := []string{e.XLgr, e.Nro, e.XBair, e.XMun, e.UF, e.CEP}
	var out []string
	for _, p := range partes {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return strings.Join(out, ", ")
}

func dataBR(s string) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return valorOuTraco(s)
	}
	return t.Format("02/01/2006")
}

func dataHoraBR(s string) string {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(s))
	if err != nil {
		return valorOuTraco(s)
	}
	return t.Format("02/01/2006 15:04")
}

func valorBR(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "0,00"
	}
	return strings.ReplaceAll(s, ".", ",")
}

func valorOuTraco(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return strings.TrimSpace(s)
}
