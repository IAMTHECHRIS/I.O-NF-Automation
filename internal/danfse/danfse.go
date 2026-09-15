package danfse

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
)

type nfseNacional struct {
	InfNFSe struct {
		ID            string        `xml:"Id,attr"`
		XLocEmi       string        `xml:"xLocEmi"`
		XLocPrestacao string        `xml:"xLocPrestacao"`
		XLocIncid     string        `xml:"xLocIncid"`
		NNFSe         string        `xml:"nNFSe"`
		ChaveAcesso   string        `xml:"chNFSe"`
		XTribNac      string        `xml:"xTribNac"`
		AmbGer        string        `xml:"ambGer"`
		CStat         string        `xml:"cStat"`
		DhProc        string        `xml:"dhProc"`
		Emit          pessoa        `xml:"emit"`
		Valores       valoresResumo `xml:"valores"`
		IBSCBS        ibsCBSNFS     `xml:"IBSCBS"`
		DPS           struct {
			InfDPS infDPS `xml:"infDPS"`
		} `xml:"DPS"`
	} `xml:"infNFSe"`
}

type infDPS struct {
	TpAmb   string `xml:"tpAmb"`
	DhEmi   string `xml:"dhEmi"`
	Serie   string `xml:"serie"`
	NDPS    string `xml:"nDPS"`
	DCompet string `xml:"dCompet"`
	Prest   struct {
		Fone    string `xml:"fone"`
		Email   string `xml:"email"`
		RegTrib struct {
			OpSimpNac string `xml:"opSimpNac"`
		} `xml:"regTrib"`
	} `xml:"prest"`
	Toma pessoa `xml:"toma"`
	Serv struct {
		LocPrest struct {
			CLocPrestacao string `xml:"cLocPrestacao"`
		} `xml:"locPrest"`
		CServ struct {
			CTribNac  string `xml:"cTribNac"`
			CTribMun  string `xml:"cTribMun"`
			CNBS      string `xml:"cNBS"`
			XDescServ string `xml:"xDescServ"`
		} `xml:"cServ"`
		Obra struct {
			CObra string `xml:"cObra"`
		} `xml:"obra"`
		InfoCompl struct {
			XPed     string `xml:"xPed"`
			GItemPed struct {
				XItemPed string `xml:"xItemPed"`
			} `xml:"gItemPed"`
		} `xml:"infoCompl"`
	} `xml:"serv"`
	Valores struct {
		VServPrest struct {
			VServ string `xml:"vServ"`
		} `xml:"vServPrest"`
		Trib tribDPS `xml:"trib"`
	} `xml:"valores"`
	IBSCBS struct {
		FinNFSe string `xml:"finNFSe"`
		CIndOp  string `xml:"cIndOp"`
		Valores struct {
			Trib struct {
				GIBSCBS struct {
					CST        string `xml:"CST"`
					CClassTrib string `xml:"cClassTrib"`
				} `xml:"gIBSCBS"`
			} `xml:"trib"`
		} `xml:"valores"`
	} `xml:"IBSCBS"`
}

type valoresResumo struct {
	VBC       string `xml:"vBC"`
	PAliq     string `xml:"pAliqAplic"`
	VISSQN    string `xml:"vISSQN"`
	VTotalRet string `xml:"vTotalRet"`
	VLiq      string `xml:"vLiq"`
}

type tribDPS struct {
	TribMun struct {
		TribISSQN  string `xml:"tribISSQN"`
		TpRetISSQN string `xml:"tpRetISSQN"`
	} `xml:"tribMun"`
	TribFed struct {
		PISCOFINS struct {
			VPis    string `xml:"vPis"`
			VCofins string `xml:"vCofins"`
		} `xml:"piscofins"`
		VRetIRRF string `xml:"vRetIRRF"`
		VRetCP   string `xml:"vRetCP"`
	} `xml:"tribFed"`
	TotTrib struct {
		PTotTrib struct {
			PTotTribFed string `xml:"pTotTribFed"`
			PTotTribEst string `xml:"pTotTribEst"`
			PTotTribMun string `xml:"pTotTribMun"`
		} `xml:"pTotTrib"`
	} `xml:"totTrib"`
}

type ibsCBSNFS struct {
	CLocalidadeIncid string `xml:"cLocalidadeIncid"`
	XLocalidadeIncid string `xml:"xLocalidadeIncid"`
	Valores          struct {
		VBC string `xml:"vBC"`
		UF  struct {
			PIBSUF      string `xml:"pIBSUF"`
			PRedAliqUF  string `xml:"pRedAliqUF"`
			PAliqEfetUF string `xml:"pAliqEfetUF"`
		} `xml:"uf"`
		Mun struct {
			PIBSMun      string `xml:"pIBSMun"`
			PRedAliqMun  string `xml:"pRedAliqMun"`
			PAliqEfetMun string `xml:"pAliqEfetMun"`
		} `xml:"mun"`
		Fed struct {
			PCBS         string `xml:"pCBS"`
			PRedAliqCBS  string `xml:"pRedAliqCBS"`
			PAliqEfetCBS string `xml:"pAliqEfetCBS"`
		} `xml:"fed"`
	} `xml:"valores"`
	TotCIBS struct {
		VTotNF string `xml:"vTotNF"`
		GIBS   struct {
			VIBSTot   string `xml:"vIBSTot"`
			GIBSUFTot struct {
				VIBSUF string `xml:"vIBSUF"`
			} `xml:"gIBSUFTot"`
			GIBSMunTot struct {
				VIBSMun string `xml:"vIBSMun"`
			} `xml:"gIBSMunTot"`
		} `xml:"gIBS"`
		GCBS struct {
			VCBS string `xml:"vCBS"`
		} `xml:"gCBS"`
	} `xml:"totCIBS"`
}

type pessoa struct {
	CNPJ     string `xml:"CNPJ"`
	CPF      string `xml:"CPF"`
	XNome    string `xml:"xNome"`
	Fone     string `xml:"fone"`
	Email    string `xml:"email"`
	EnderNac endNac `xml:"enderNac"`
	End      struct {
		EndNac endNac `xml:"endNac"`
		XLgr   string `xml:"xLgr"`
		Nro    string `xml:"nro"`
		XBair  string `xml:"xBairro"`
	} `xml:"end"`
}

type endNac struct {
	CMun  string `xml:"cMun"`
	CEP   string `xml:"CEP"`
	UF    string `xml:"UF"`
	XMun  string `xml:"xMun"`
	XLgr  string `xml:"xLgr"`
	Nro   string `xml:"nro"`
	XBair string `xml:"xBairro"`
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

func Gerar(w io.Writer, n nfseNacional, direcao, gerador string) error {
	if strings.TrimSpace(n.InfNFSe.ChaveAcesso) == "" {
		n.InfNFSe.ChaveAcesso = strings.TrimPrefix(n.InfNFSe.ID, "NFS")
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(3, 3, 3)
	pdf.SetAutoPageBreak(false, 3)
	pdf.AddPage()
	r := &renderer{pdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}
	r.header(n)
	r.identificacao(n, direcao)
	r.pessoa("PRESTADOR / FORNECEDOR", prestador(n), n.InfNFSe.DPS.InfDPS.Prest.RegTrib.OpSimpNac)
	r.pessoa("TOMADOR / ADQUIRENTE", n.InfNFSe.DPS.InfDPS.Toma, "")
	r.centerLine("DESTINATÁRIO DA OPERAÇÃO NÃO IDENTIFICADO NA NFS-e")
	r.centerLine("INTERMEDIÁRIO DA OPERAÇÃO NÃO IDENTIFICADO NA NFS-e")
	r.servico(n)
	r.tributacaoMunicipal(n)
	r.tributacaoFederal(n)
	r.tributacaoIBSCBS(n)
	r.valores(n)
	r.complementares(n)
	r.rodape(n)
	return pdf.Output(w)
}

type renderer struct {
	pdf *gofpdf.Fpdf
	tr  func(string) string
}

func (r *renderer) cell(w, h float64, txt, border string, ln int, align string, fill bool) {
	r.pdf.CellFormat(w, h, r.tr(txt), border, ln, align, fill, 0, "")
}

func (r *renderer) text(x, y, w float64, s string, size float64, style string) {
	r.pdf.SetFont("Arial", style, size)
	r.pdf.SetXY(x, y)
	r.cell(w, 3.2, r.fit(valueOrDash(s), w), "", 0, "L", false)
}

func (r *renderer) fit(s string, w float64) string {
	s = strings.Join(strings.Fields(s), " ")
	if r.pdf.GetStringWidth(r.tr(s)) <= w {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."
		if r.pdf.GetStringWidth(r.tr(candidate)) <= w {
			return candidate
		}
	}
	return ""
}

func (r *renderer) paragraph(x, y, w, lineH float64, s string, size float64, maxLines int) {
	r.pdf.SetFont("Arial", "", size)
	lines, clipped := r.wrapLines(s, w, maxLines)
	for i, line := range lines {
		if clipped && i == len(lines)-1 {
			line = r.fit(line+"...", w)
		}
		r.pdf.SetXY(x, y+float64(i)*lineH)
		r.cell(w, lineH, line, "", 0, "L", false)
	}
}

func (r *renderer) wrapLines(s string, w float64, maxLines int) ([]string, bool) {
	words := strings.Fields(s)
	if len(words) == 0 || maxLines <= 0 {
		return nil, false
	}
	var lines []string
	line := ""
	for _, word := range words {
		candidate := strings.TrimSpace(line + " " + word)
		if line == "" || r.pdf.GetStringWidth(r.tr(candidate)) <= w {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
		if len(lines) == maxLines {
			return lines, true
		}
	}
	if line != "" && len(lines) < maxLines {
		lines = append(lines, line)
	}
	return lines, false
}

func (r *renderer) section(title string, y, h float64, cols ...float64) {
	r.pdf.Rect(3, y, 204, h, "D")
	r.pdf.SetFillColor(238, 238, 238)
	r.pdf.Rect(3, y, 51, h, "F")
	r.text(5, y+1, 48, title, 7.4, "B")
	for _, c := range cols {
		r.pdf.Line(3+c, y, 3+c, y+h)
	}
}

func (r *renderer) header(n nfseNacional) {
	r.pdf.Rect(3, 3, 204, 14, "D")
	r.pdf.SetTextColor(31, 139, 86)
	r.pdf.SetFont("Arial", "B", 16)
	r.pdf.SetXY(5, 5)
	r.cell(34, 8, "NFS-e", "", 0, "L", false)
	r.pdf.SetTextColor(42, 93, 172)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetFont("Arial", "", 5.5)
	r.pdf.SetXY(38, 7)
	r.cell(26, 3, "Nota Fiscal de", "", 0, "L", false)
	r.pdf.SetXY(38, 10)
	r.cell(30, 3, "Serviço eletrônica", "", 0, "L", false)
	r.pdf.SetFont("Arial", "B", 10)
	r.pdf.SetXY(70, 5.7)
	r.cell(70, 4, "DANFSe v2.0", "", 1, "C", false)
	r.pdf.SetX(70)
	r.cell(70, 4, "Documento Auxiliar da NFS-e", "", 0, "C", false)
	r.pdf.SetFont("Arial", "", 7)
	r.pdf.SetXY(157, 5)
	r.cell(48, 3, "Município: "+loc(n.InfNFSe.XLocEmi, "São Carlos")+" - SP", "", 1, "L", false)
	r.pdf.SetX(157)
	r.cell(48, 3, "Ambiente Gerador: "+valueOrDash(n.InfNFSe.AmbGer), "", 1, "L", false)
	r.pdf.SetX(157)
	r.cell(48, 3, "Tipo de Ambiente: "+valueOrDash(n.InfNFSe.DPS.InfDPS.TpAmb), "", 1, "L", false)
}

func (r *renderer) identificacao(n nfseNacional, direcao string) {
	y := 17.0
	r.text(5, y+2, 80, "CHAVE DE ACESSO DA NFS-e", 7.3, "B")
	r.text(5, y+6, 130, n.InfNFSe.ChaveAcesso, 7, "")
	if img, err := qrCode(n.InfNFSe.ChaveAcesso); err == nil {
		opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		r.pdf.RegisterImageOptionsReader("qr-nfse", opt, img)
		r.pdf.ImageOptions("qr-nfse", 174, y+2, 17, 17, false, opt, 0, "")
	}
	r.pdf.SetFont("Arial", "", 6)
	r.pdf.SetXY(157, y+17)
	r.pdf.MultiCell(46, 3, r.tr("A autenticidade desta NFS-e pode ser verificada\npela leitura deste código QR ou pela consulta da\nchave de acesso no portal nacional da NFS-e"), "", "L", false)
	y = 29
	r.box(3, y, 51, 12, "NÚMERO DA NFS-e", n.InfNFSe.NNFSe)
	r.box(54, y, 51, 12, "COMPETÊNCIA DA NFS-e", dataBR(n.InfNFSe.DPS.InfDPS.DCompet))
	r.box(105, y, 51, 12, "DATA E HORA DA EMISSÃO DA NFS-e", dataHoraBR(n.InfNFSe.DPS.InfDPS.DhEmi))
	r.box(3, y+12, 51, 12, "NÚMERO DA DPS", n.InfNFSe.DPS.InfDPS.NDPS)
	r.box(54, y+12, 51, 12, "SÉRIE DA DPS", n.InfNFSe.DPS.InfDPS.Serie)
	r.box(105, y+12, 51, 12, "DATA E HORA DA EMISSÃO DA DPS", dataHoraBR(n.InfNFSe.DPS.InfDPS.DhEmi))
	r.box(3, y+24, 51, 10, "EMITENTE DA NFS-e", "Prestador")
	r.box(54, y+24, 51, 10, "SITUAÇÃO DA NFS-e", situacao(n.InfNFSe.CStat))
	r.box(105, y+24, 51, 10, "FINALIDADE", finalidade(n.InfNFSe.DPS.InfDPS.IBSCBS.FinNFSe, direcao))
	r.pdf.SetY(63)
}

func (r *renderer) box(x, y, w, h float64, label, value string) {
	r.pdf.Rect(x, y, w, h, "D")
	r.pdf.SetFillColor(238, 238, 238)
	r.pdf.Rect(x, y, w, 5, "F")
	r.text(x+1, y+1, w-2, label, 6.2, "B")
	r.text(x+1, y+5.5, w-2, value, 7, "")
}

func (r *renderer) pessoa(titulo string, p pessoa, simples string) {
	y := r.pdf.GetY()
	h := 27.0
	if titulo == "PRESTADOR / FORNECEDOR" {
		h = 37
	}
	r.section(titulo, y, h, 51, 102, 153)
	r.text(55, y+1, 45, "CNPJ / CPF / NIF", 6.1, "B")
	r.text(55, y+5, 45, docFormatado(p), 7, "")
	r.text(106, y+1, 45, "Indicador Municipal (Inscrição)", 6.1, "B")
	r.text(106, y+5, 45, "-", 7, "")
	r.text(157, y+1, 45, "Telefone", 6.1, "B")
	r.text(157, y+5, 45, fone(p.Fone), 7, "")
	r.text(5, y+10, 48, "Nome / Nome Empresarial", 6.1, "B")
	r.text(5, y+14, 96, p.XNome, 7, "")
	r.text(106, y+10, 45, "Município / Sigla UF", 6.1, "B")
	r.text(106, y+14, 45, municipioUF(p), 7, "")
	r.text(157, y+10, 45, "Código IBGE / CEP", 6.1, "B")
	r.text(157, y+14, 45, codCEP(p), 7, "")
	r.text(5, y+18, 48, "Endereço", 6.1, "B")
	r.text(5, y+22, 96, endereco(p), 7, "")
	r.text(106, y+18, 45, "E-mail", 6.1, "B")
	r.text(106, y+22, 96, strings.ToLower(valueOrDash(p.Email)), 7, "")
	if titulo == "PRESTADOR / FORNECEDOR" {
		r.text(5, y+28, 50, "Simples Nacional na Data de Competência", 6.1, "B")
		r.text(5, y+32, 50, simplesNacional(simples), 7, "")
		r.text(55, y+28, 75, "Regime de Apuração Tributária pelo SN", 6.1, "B")
		r.text(55, y+32, 145, regimeSN(simples), 7, "")
	}
	r.pdf.SetY(y + h)
}

func (r *renderer) centerLine(s string) {
	y := r.pdf.GetY()
	r.pdf.Line(3, y, 207, y)
	r.pdf.SetFont("Arial", "", 7)
	r.pdf.SetXY(3, y+0.4)
	r.cell(204, 4, s, "", 1, "C", false)
	r.pdf.SetY(y + 4)
}

func (r *renderer) servico(n nfseNacional) {
	y := r.pdf.GetY()
	svc := n.InfNFSe.DPS.InfDPS.Serv.CServ
	r.pdf.Rect(3, y, 204, 32, "D")
	r.pdf.SetFillColor(238, 238, 238)
	r.pdf.Rect(3, y, 51, 10, "F")
	r.pdf.Line(54, y, 54, y+10)
	r.pdf.Line(105, y, 105, y+10)
	r.pdf.Line(156, y, 156, y+10)
	r.pdf.Line(3, y+10, 207, y+10)
	r.text(5, y+1, 48, "SERVIÇO PRESTADO", 7.4, "B")
	r.text(55, y+1, 48, "Código de Tributação Nacional/Municipal", 5.8, "B")
	r.text(55, y+5, 45, formatTrib(svc.CTribNac)+" / "+valueOrDash(svc.CTribMun), 7, "")
	r.text(106, y+1, 45, "Código da NBS", 6.1, "B")
	r.text(106, y+5, 45, formatNBS(svc.CNBS), 7, "")
	r.text(157, y+1, 48, "Local da Prestação / Sigla UF / País", 6.1, "B")
	r.text(157, y+5, 48, loc(n.InfNFSe.XLocPrestacao, n.InfNFSe.XLocEmi)+" / SP / -", 7, "")
	r.paragraph(5, y+12, 198, 3, n.InfNFSe.XTribNac, 6.3, 3)
	r.text(5, y+20, 60, "Descrição do Serviço", 6.1, "B")
	r.paragraph(5, y+24, 198, 2.8, svc.XDescServ, 6.15, 3)
	r.pdf.SetY(y + 32)
}

func (r *renderer) tributacaoMunicipal(n nfseNacional) {
	y := r.pdf.GetY()
	v := n.InfNFSe.Valores
	r.section("TRIBUTAÇÃO MUNICIPAL (ISSQN)", y, 18, 51, 102, 153)
	r.text(55, y+1, 45, "Tipo de Tributação do ISSQN", 6.1, "B")
	r.text(55, y+5, 45, tribISS(n.InfNFSe.DPS.InfDPS.Valores.Trib.TribMun.TribISSQN), 7, "")
	r.text(106, y+1, 95, "Município / UF / País de Incidência do ISSQN", 6.1, "B")
	r.text(106, y+5, 95, loc(n.InfNFSe.XLocIncid, n.InfNFSe.XLocEmi)+" / SP / -", 7, "")
	r.row4(y+9, "BC ISSQN", money(v.VBC), "Alíquota Aplicada", percent(v.PAliq), "Retenção do ISSQN", retISS(n.InfNFSe.DPS.InfDPS.Valores.Trib.TribMun.TpRetISSQN), "ISSQN Apurado", money(v.VISSQN))
	r.pdf.SetY(y + 18)
}

func (r *renderer) tributacaoFederal(n nfseNacional) {
	y := r.pdf.GetY()
	f := n.InfNFSe.DPS.InfDPS.Valores.Trib.TribFed
	r.section("TRIBUTAÇÃO FEDERAL (EXCETO CBS)", y, 18, 51, 102, 153)
	r.rowFrom2(y+1, "IRRF", money(f.VRetIRRF), "Contrib. Previdenciária Retida", money(f.VRetCP), "Contribuições Sociais Retidas", "-")
	r.row3(y+9, "PIS - Débito Apuração Própria", money(f.PISCOFINS.VPis), "COFINS - Débito Apuração Própria", money(f.PISCOFINS.VCofins), "Descrição Contrib. Sociais - Retidas", "0 - PIS/COFINS/CSLL Não Retidos")
	r.pdf.SetY(y + 18)
}

func (r *renderer) tributacaoIBSCBS(n nfseNacional) {
	y := r.pdf.GetY()
	i, d := n.InfNFSe.IBSCBS, n.InfNFSe.DPS.InfDPS.IBSCBS
	r.section("TRIBUTAÇÃO IBS/CBS", y, 33, 51, 102, 153)
	r.rowFrom2(y+1, "CST / cClassTrib", valueOrDash(d.Valores.Trib.GIBSCBS.CST)+" / "+valueOrDash(d.Valores.Trib.GIBSCBS.CClassTrib), "Indicador / IBGE / Município / UF", valueOrDash(d.CIndOp)+" / "+valueOrDash(i.CLocalidadeIncid)+" / "+loc(i.XLocalidadeIncid, "-")+" / SP", "", "")
	r.row4(y+9, "Exclusões/Reduções BC", "R$ 0,00", "BC após Exclusões/Reduções", money(i.Valores.VBC), "Red. Alíq. IBS / CBS", percent(i.Valores.UF.PRedAliqUF)+" / "+percent(i.Valores.Mun.PRedAliqMun)+" / "+percent(i.Valores.Fed.PRedAliqCBS), "Alíq. IBS UF / Mun", percent(i.Valores.UF.PIBSUF)+" / "+percent(i.Valores.Mun.PIBSMun))
	r.row4(y+17, "Alíq. Efetiva Municipal - IBS", percent(i.Valores.Mun.PAliqEfetMun), "Valor Apurado Municipal - IBS", money(i.TotCIBS.GIBS.GIBSMunTot.VIBSMun), "Alíq. Efetiva Estadual - IBS", percent(i.Valores.UF.PAliqEfetUF), "Valor Apurado Estadual - IBS", money(i.TotCIBS.GIBS.GIBSUFTot.VIBSUF))
	r.row4(y+25, "Valor Total Apurado - IBS", money(i.TotCIBS.GIBS.VIBSTot), "Alíquota - CBS", percent(i.Valores.Fed.PCBS), "Alíquota Efetiva - CBS", percent(i.Valores.Fed.PAliqEfetCBS), "Valor Total Apurado - CBS", money(i.TotCIBS.GCBS.VCBS))
	r.pdf.SetY(y + 33)
}

func (r *renderer) valores(n nfseNacional) {
	y := r.pdf.GetY()
	v, serv := n.InfNFSe.Valores, n.InfNFSe.DPS.InfDPS.Valores.VServPrest.VServ
	if strings.TrimSpace(serv) == "" {
		serv = v.VBC
	}
	r.section("VALOR TOTAL DA NFS-e", y, 18, 51, 102, 153)
	r.rowFrom2(y+1, "VALOR DA OPERAÇÃO / SERVIÇO", money(serv), "Desconto Incondicionado", "-", "Desconto Condicionado", "-")
	r.row4(y+9, "Total Retenções", money(v.VTotalRet), "VALOR LÍQUIDO DA NFS-e", money(v.VLiq), "Total do IBS/CBS", money(totalIBSCBS(n)), "Líquido NFS-e + IBS/CBS", money(first(n.InfNFSe.IBSCBS.TotCIBS.VTotNF, v.VLiq)))
	r.pdf.SetY(y + 18)
}

func (r *renderer) complementares(n nfseNacional) {
	y := r.pdf.GetY()
	r.pdf.Rect(3, y, 204, 19, "D")
	r.pdf.SetFillColor(238, 238, 238)
	r.pdf.Rect(3, y, 204, 5, "F")
	r.text(5, y+1, 120, "INFORMAÇÕES COMPLEMENTARES", 7.4, "B")
	r.pdf.SetFont("Arial", "", 6.3)
	r.pdf.SetXY(5, y+7)
	r.pdf.MultiCell(198, 3, r.tr(compl(n)), "", "L", false)
	r.pdf.SetY(y + 19)
}

func (r *renderer) rodape(n nfseNacional) {
	y := 276.0
	r.pdf.Rect(3, y, 204, 8, "D")
	r.pdf.Line(54, y, 54, y+8)
	r.pdf.Line(105, y, 105, y+8)
	r.text(5, y+1, 48, "DATA CIENTIFICAÇÃO:", 6.3, "B")
	r.text(56, y+1, 48, "IDENTIFICAÇÃO E ASSINATURA", 6.3, "B")
	r.text(107, y+1, 48, "N° NFS-e / CHAVE NFS-e", 6.3, "B")
	r.text(107, y+4.5, 96, n.InfNFSe.NNFSe+" / "+n.InfNFSe.ChaveAcesso, 6.3, "")
	r.pdf.Rect(2, 2, 206, 293, "D")
}

func (r *renderer) row4(y float64, a, b, c, d, e, f, g, h string) {
	xs := []float64{5, 56, 107, 158}
	labels := []string{a, c, e, g}
	values := []string{b, d, f, h}
	for i := range xs {
		if labels[i] == "" {
			continue
		}
		r.text(xs[i], y, 47, labels[i], 5.9, "B")
		r.text(xs[i], y+4, 47, values[i], 6.8, "")
	}
}

func (r *renderer) row3(y float64, a, b, c, d, e, f string) {
	xs, ws := []float64{5, 56, 107}, []float64{47, 47, 96}
	labels, values := []string{a, c, e}, []string{b, d, f}
	for i := range xs {
		if labels[i] == "" {
			continue
		}
		r.text(xs[i], y, ws[i], labels[i], 5.9, "B")
		r.text(xs[i], y+4, ws[i], values[i], 6.8, "")
	}
}

func (r *renderer) rowFrom2(y float64, a, b, c, d, e, f string) {
	xs, ws := []float64{56, 107, 158}, []float64{47, 47, 47}
	labels, values := []string{a, c, e}, []string{b, d, f}
	for i := range xs {
		if labels[i] == "" {
			continue
		}
		r.text(xs[i], y, ws[i], labels[i], 5.9, "B")
		r.text(xs[i], y+4, ws[i], values[i], 6.8, "")
	}
}

func prestador(n nfseNacional) pessoa {
	p := n.InfNFSe.Emit
	p.Fone = first(n.InfNFSe.DPS.InfDPS.Prest.Fone, p.Fone)
	p.Email = first(n.InfNFSe.DPS.InfDPS.Prest.Email, p.Email)
	return p
}

func docFormatado(p pessoa) string {
	d := onlyDigits(first(p.CNPJ, p.CPF))
	if len(d) == 14 {
		return d[:2] + "." + d[2:5] + "." + d[5:8] + "/" + d[8:12] + "-" + d[12:]
	}
	if len(d) == 11 {
		return d[:3] + "." + d[3:6] + "." + d[6:9] + "-" + d[9:]
	}
	return valueOrDash(first(p.CNPJ, p.CPF))
}

func endereco(p pessoa) string {
	e := end(p)
	return valueOrDash(strings.Join(nonEmpty(e.XLgr, e.Nro, e.XBair), ", "))
}

func municipioUF(p pessoa) string {
	e := end(p)
	return loc(e.XMun, cidadePorCodigo(e.CMun)) + " / " + loc(e.UF, ufPorCodigo(e.CMun))
}

func codCEP(p pessoa) string {
	e := end(p)
	return formatCodMun(e.CMun) + " / " + cep(e.CEP)
}

func end(p pessoa) endNac {
	e := p.EnderNac
	if e.CMun == "" && e.CEP == "" && e.XLgr == "" {
		e = p.End.EndNac
		e.XLgr = first(e.XLgr, p.End.XLgr)
		e.Nro = first(e.Nro, p.End.Nro)
		e.XBair = first(e.XBair, p.End.XBair)
	}
	return e
}

func dataBR(s string) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return valueOrDash(s)
	}
	return t.Format("02/01/2006")
}

func dataHoraBR(s string) string {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(s))
	if err != nil {
		return valueOrDash(s)
	}
	return t.Format("02/01/2006 15:04:05")
}

func money(s string) string {
	v, ok := decimal(s)
	if !ok {
		return "-"
	}
	return "R$ " + formatBR(v)
}

func percent(s string) string {
	v, ok := decimal(s)
	if !ok {
		return "-"
	}
	return formatBR(v) + " %"
}

func decimal(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", "."), 64)
	return v, err == nil
}

func formatBR(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	intPart, decPart := s[:len(s)-3], s[len(s)-2:]
	var out []byte
	for i, c := range []byte(intPart) {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	return string(out) + "," + decPart
}

func valueOrDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return strings.TrimSpace(s)
}

func first(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func nonEmpty(vals ...string) []string {
	var out []string
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			out = append(out, strings.TrimSpace(v))
		}
	}
	return out
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func fone(s string) string {
	d := onlyDigits(s)
	if len(d) == 11 {
		return "(" + d[:2] + ") " + d[2:7] + "-" + d[7:]
	}
	if len(d) == 10 {
		return "(" + d[:2] + ") " + d[2:6] + "-" + d[6:]
	}
	return valueOrDash(s)
}

func cep(s string) string {
	d := onlyDigits(s)
	if len(d) == 8 {
		return d[:2] + "." + d[2:5] + "-" + d[5:]
	}
	return valueOrDash(s)
}

func formatCodMun(s string) string {
	d := onlyDigits(s)
	if len(d) == 7 {
		return d[:2] + "." + d[2:]
	}
	return valueOrDash(s)
}

func formatTrib(s string) string {
	d := onlyDigits(s)
	if len(d) == 6 {
		return d[:2] + "." + d[2:4] + "." + d[4:]
	}
	return valueOrDash(s)
}

func formatNBS(s string) string {
	d := onlyDigits(s)
	if len(d) == 9 {
		return d[:1] + "." + d[1:5] + "." + d[5:7] + "." + d[7:]
	}
	return valueOrDash(s)
}

func loc(s, fallback string) string {
	if strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	return valueOrDash(fallback)
}

func cidadePorCodigo(c string) string {
	switch onlyDigits(c) {
	case "3548906":
		return "São Carlos"
	case "3510005":
		return "Cândido Mota"
	default:
		return ""
	}
}

func ufPorCodigo(c string) string {
	if strings.HasPrefix(onlyDigits(c), "35") {
		return "SP"
	}
	return ""
}

func situacao(cstat string) string {
	if strings.TrimSpace(cstat) == "100" || strings.TrimSpace(cstat) == "" {
		return "NFS-e Gerada"
	}
	return "cStat " + cstat
}

func finalidade(fin, direcao string) string {
	if strings.TrimSpace(fin) == "0" {
		return "NFS-e regular"
	}
	return "-"
}

func simplesNacional(s string) string {
	if strings.TrimSpace(s) == "1" {
		return "Não optante"
	}
	if strings.TrimSpace(s) == "2" || strings.TrimSpace(s) == "3" {
		return "Optante"
	}
	return "-"
}

func regimeSN(s string) string {
	if strings.TrimSpace(s) == "2" || strings.TrimSpace(s) == "3" {
		return "Regime de apuração dos tributos federais e municipal pelo Simples Nacional"
	}
	return "-"
}

func tribISS(s string) string {
	if strings.TrimSpace(s) == "1" || strings.TrimSpace(s) == "" {
		return "Operação Tributável"
	}
	return s
}

func retISS(s string) string {
	if strings.TrimSpace(s) == "2" {
		return "Retido pelo Tomador"
	}
	if strings.TrimSpace(s) == "1" {
		return "Não Retido"
	}
	return "-"
}

func totalIBSCBS(n nfseNacional) string {
	ibs, _ := decimal(n.InfNFSe.IBSCBS.TotCIBS.GIBS.VIBSTot)
	cbs, _ := decimal(n.InfNFSe.IBSCBS.TotCIBS.GCBS.VCBS)
	return fmt.Sprintf("%.2f", ibs+cbs)
}

func compl(n nfseNacional) string {
	info := n.InfNFSe.DPS.InfDPS.Serv.InfoCompl
	var partes []string
	if obra := strings.TrimSpace(n.InfNFSe.DPS.InfDPS.Serv.Obra.CObra); obra != "" {
		partes = append(partes, "Cod. Obra: "+obra)
	}
	if ped := strings.TrimSpace(info.XPed); ped != "" {
		partes = append(partes, "Núm. Ped.: "+ped)
	}
	if item := strings.TrimSpace(info.GItemPed.XItemPed); item != "" {
		partes = append(partes, "Item Ped.: "+item)
	}
	t := n.InfNFSe.DPS.InfDPS.Valores.Trib.TotTrib.PTotTrib
	partes = append(partes, "Totais aproximados dos Tributos cfe. Lei n° 12.741/2012: Federais: "+percent(t.PTotTribFed)+"; Estaduais: "+percent(t.PTotTribEst)+"; Municipais: "+percent(t.PTotTribMun)+";")
	return strings.Join(partes, " | ")
}

func short(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

func qrCode(chave string) (io.Reader, error) {
	code, err := qr.Encode(chave, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}
	scaled, err := barcode.Scale(code, 180, 180)
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
