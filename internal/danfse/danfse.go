package danfse

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/jung-kurt/gofpdf"
)

type nfseNacional struct {
	InfNFSe struct {
		NNFSe       string `xml:"nNFSe"`
		ChaveAcesso string `xml:"chNFSe"`
		DhProc      string `xml:"dhProc"`
		Emit        pessoa `xml:"emit"`
		Valores     struct {
			VServPrest  string `xml:"vServPrest"`
			VDescIncond string `xml:"vDescIncond"`
			VDescCond   string `xml:"vDescCond"`
			VBC         string `xml:"vBC"`
			VISSQN      string `xml:"vISSQN"`
			VISSQNRet   string `xml:"vISSQNRet"`
			VLiq        string `xml:"vLiq"`
		} `xml:"valores"`
		DPS struct {
			InfDPS struct {
				DCompet string `xml:"dCompet"`
				Toma    pessoa `xml:"toma"`
				Serv    struct {
					CServ struct {
						CTribNac  string `xml:"cTribNac"`
						CTribMun  string `xml:"cTribMun"`
						CNAE      string `xml:"CNAE"`
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
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(true, 8)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	cell := func(w, h float64, txt string, border string, ln int, align string, fill bool) {
		pdf.CellFormat(w, h, tr(txt), border, ln, align, fill, 0, "")
	}
	multi := func(w, h float64, txt string, border string) {
		pdf.MultiCell(w, h, tr(txt), border, "L", false)
	}
	campo := func(label, valor string, w, h float64) {
		x, y := pdf.GetX(), pdf.GetY()
		pdf.Rect(x, y, w, h, "D")
		pdf.SetFont("Arial", "", 6)
		pdf.SetXY(x+1, y+1)
		cell(w-2, 3, label, "", 0, "L", false)
		pdf.SetFont("Arial", "B", 8)
		pdf.SetXY(x+1, y+4.2)
		cell(w-2, h-4.8, valorOuTraco(valor), "", 0, "L", false)
		pdf.SetXY(x+w, y)
	}
	secao := func(titulo string) {
		pdf.SetFillColor(232, 232, 232)
		pdf.SetFont("Arial", "B", 8)
		cell(0, 6, titulo, "1", 1, "L", true)
	}

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 13)
	cell(112, 9, "DANFSe - Documento Auxiliar da NFS-e", "1", 0, "C", false)
	pdf.SetFont("Arial", "", 7)
	cell(40, 9, "NFS-e Nacional", "1", 0, "C", false)
	pdf.SetFont("Arial", "B", 9)
	cell(0, 9, "No. "+valorOuTraco(nfse.InfNFSe.NNFSe), "1", 1, "C", false)

	chave := strings.TrimSpace(nfse.InfNFSe.ChaveAcesso)
	pdf.SetFont("Arial", "", 6.5)
	cell(0, 4, "Chave de acesso", "LR", 1, "C", false)
	pdf.SetFont("Arial", "B", 8)
	cell(0, 5, chaveFormatada(chave), "LRB", 1, "C", false)
	if chave != "" {
		if img, err := codigoBarras(chave); err == nil {
			opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
			pdf.RegisterImageOptionsReader("chave-nfse", opt, img)
			pdf.ImageOptions("chave-nfse", 48, pdf.GetY()+1.5, 114, 12, false, opt, 0, "")
			pdf.Ln(15)
		}
	}

	campo("Competência", dataBR(nfse.InfNFSe.DPS.InfDPS.DCompet), 48, 10)
	campo("Processamento", dataHoraBR(nfse.InfNFSe.DhProc), 58, 10)
	campo("Direção", strings.ToUpper(direcao), 34, 10)
	campo("Valor líquido", "R$ "+valorBR(nfse.InfNFSe.Valores.VLiq), 0, 10)
	pdf.Ln(12)

	secao("PRESTADOR DO SERVIÇO")
	pessoaBox(pdf, tr, nfse.InfNFSe.Emit)
	pdf.Ln(2)

	secao("TOMADOR DO SERVIÇO")
	pessoaBox(pdf, tr, nfse.InfNFSe.DPS.InfDPS.Toma)
	pdf.Ln(2)

	secao("DISCRIMINAÇÃO DO SERVIÇO")
	svc := nfse.InfNFSe.DPS.InfDPS.Serv.CServ
	linha2(pdf, tr, "Trib. nacional", svc.CTribNac, "Trib. municipal", svc.CTribMun)
	linha2(pdf, tr, "CNAE", svc.CNAE, "Número NFS-e", nfse.InfNFSe.NNFSe)
	descricao := strings.TrimSpace(nfse.InfNFSe.DPS.InfDPS.Serv.CServ.XDescServ)
	if descricao == "" {
		descricao = "Sem descrição no XML."
	}
	pdf.SetFont("Arial", "", 8)
	multi(0, 4.2, descricao, "LRB")
	pdf.Ln(2)

	secao("VALORES E TRIBUTAÇÃO")
	val := nfse.InfNFSe.Valores
	linha4(pdf, tr, "Serviços", val.VServPrest, "Desc. incond.", val.VDescIncond, "Desc. cond.", val.VDescCond, "Base ISS", val.VBC)
	linha4(pdf, tr, "ISS", val.VISSQN, "ISS retido", val.VISSQNRet, "Líquido", val.VLiq, "Situação", statusISS(val))

	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	cell(0, 5, "Documento auxiliar gerado a partir do XML oficial. Gerado por "+gerador+" em "+time.Now().Format("02/01/2006 15:04"), "", 1, "R", false)

	return pdf.Output(w)
}

func pessoaBox(pdf *gofpdf.Fpdf, tr func(string) string, p pessoa) {
	x, y := pdf.GetX(), pdf.GetY()
	w := 194.0
	pdf.Rect(x, y, w, 21, "D")
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(x+1, y+1)
	pdf.CellFormat(w-2, 4, tr(valorOuTraco(p.XNome)), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 7)
	pdf.SetX(x + 1)
	pdf.CellFormat(60, 4, tr("CNPJ/CPF: "+doc(p)), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 4, tr("Endereço: "+valorOuTraco(endereco(p))), "", 1, "L", false, 0, "")
	pdf.SetX(x + 1)
	pdf.CellFormat(45, 4, tr("Município: "+valorOuTraco(p.End.EndNac.XMun)), "", 0, "L", false, 0, "")
	pdf.CellFormat(18, 4, tr("UF: "+valorOuTraco(p.End.EndNac.UF)), "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 4, tr("CEP: "+valorOuTraco(p.End.EndNac.CEP)), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 4, tr("Código município: "+valorOuTraco(p.End.EndNac.CMun)), "", 1, "L", false, 0, "")
	pdf.SetXY(x, y+23)
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

func linha4(pdf *gofpdf.Fpdf, tr func(string) string, a, b, c, d, e, f, g, h string) {
	cols := []struct {
		label string
		valor string
	}{{a, b}, {c, d}, {e, f}, {g, h}}
	for i, col := range cols {
		pdf.SetFont("Arial", "B", 7)
		pdf.CellFormat(24, 7, tr(col.label+":"), "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 7)
		ln := 0
		if i == len(cols)-1 {
			ln = 1
		}
		pdf.CellFormat(24.5, 7, tr(valorMonetarioOuTexto(col.valor)), "1", ln, "L", false, 0, "")
	}
}

func statusISS(val struct {
	VServPrest  string `xml:"vServPrest"`
	VDescIncond string `xml:"vDescIncond"`
	VDescCond   string `xml:"vDescCond"`
	VBC         string `xml:"vBC"`
	VISSQN      string `xml:"vISSQN"`
	VISSQNRet   string `xml:"vISSQNRet"`
	VLiq        string `xml:"vLiq"`
}) string {
	if strings.TrimSpace(val.VISSQNRet) != "" && strings.TrimSpace(val.VISSQNRet) != "0" && strings.TrimSpace(val.VISSQNRet) != "0.00" {
		return "ISS retido"
	}
	return "Sem retenção"
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

func valorMonetarioOuTexto(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if strings.ContainsAny(s, "0123456789") && !strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") {
		return "R$ " + valorBR(s)
	}
	return s
}

func valorOuTraco(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return strings.TrimSpace(s)
}

func chaveFormatada(chave string) string {
	chave = strings.TrimSpace(chave)
	var partes []string
	for len(chave) > 0 {
		n := 4
		if len(chave) < n {
			n = len(chave)
		}
		partes = append(partes, chave[:n])
		chave = chave[n:]
	}
	return strings.Join(partes, " ")
}

func codigoBarras(chave string) (io.Reader, error) {
	bar, err := code128.Encode(chave)
	if err != nil {
		return nil, err
	}
	scaled, err := barcode.Scale(bar, 900, 90)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, ensureRGBA(scaled)); err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func ensureRGBA(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	return dst
}
