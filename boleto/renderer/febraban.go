package renderer

import (
	"bytes"
	"context"
	"image"
	_ "image/png"
	"strconv"

	"github.com/go-pdf/fpdf"
)

// FEBRABAN dimension constants (in mm)
const (
	BarcodeWidth  = 103.0
	BarcodeHeight = 13.0
	QRCodeSize    = 25.0
	LogoMaxWidth  = 40.0
	LogoMaxHeight = 10.0

	// Field heights
	fieldHeight      = 8.0
	fieldHeightSmall = 6.0
	labelFontSize    = 5.0
	valueFontSize    = 7.0
	headerFontSize   = 9.0
)

// NewFebraban creates a new Febraban renderer with default settings.
func NewFebraban(opts ...FebOption) *Febraban {
	f := &Febraban{
		primaryColor:   "#000000",
		pageWidth:      210, // A4
		pageHeight:     297, // A4
		margin:         10,
		showReceipt:    true,
		showPIXSection: true,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Render generates the boleto PDF following FEBRABAN standard layout.
func (f *Febraban) Render(ctx context.Context, data *RenderData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(f.margin, f.margin, f.margin)
	pdf.AddPage()

	// Create UTF-8 to cp1252 translator for Portuguese text
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	contentWidth := f.pageWidth - 2*f.margin
	y := f.margin

	// Receipt section (Recibo do Sacado - top)
	if f.showReceipt {
		y = f.renderReciboSacado(pdf, data, y, contentWidth, tr)
		y = f.renderCutLine(pdf, y, contentWidth, tr)
	}

	// Main boleto section (Ficha de Compensação)
	y = f.renderFichaCompensacao(pdf, data, y, contentWidth, tr)

	// Generate PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// renderReciboSacado renders the payer's receipt (top copy)
func (f *Febraban) renderReciboSacado(pdf *fpdf.Fpdf, data *RenderData, y float64, contentWidth float64, tr func(string) string) float64 {
	b := data.Boleto
	startY := y

	// Header with bank logo and digitable line
	y = f.renderBankHeader(pdf, data, y, contentWidth, tr)

	// Define column widths for receipt
	col1 := contentWidth * 0.55 // Left side (beneficiary info)
	col2 := contentWidth * 0.15 // Espécie
	col3 := contentWidth * 0.15 // Quantidade
	col4 := contentWidth * 0.15 // Nosso número

	// Row 1: Cedente | Agência/Código | Espécie | Quantidade | Nosso número
	f.renderFieldBox(pdf, f.margin, y, col1, fieldHeightSmall, "Cedente", b.Beneficiary.Name, tr)
	agenciaCodigo := b.Agency + "-" + b.AgencyDigit + " / " + b.Account + "-" + b.AccountDigit
	f.renderFieldBox(pdf, f.margin+col1, y, col2, fieldHeightSmall, "Agência / Código do Cedente", agenciaCodigo, tr)
	f.renderFieldBox(pdf, f.margin+col1+col2, y, col3, fieldHeightSmall, "Espécie", "R$", tr)
	f.renderFieldBox(pdf, f.margin+col1+col2+col3, y, col4, fieldHeightSmall, "Quantidade", "", tr)
	y += fieldHeightSmall

	// Row 2: Número documento | Contrato | CPF/CNPJ | Vencimento | Valor documento
	col1b := contentWidth * 0.20
	col2b := contentWidth * 0.20
	col3b := contentWidth * 0.20
	col4b := contentWidth * 0.20
	col5b := contentWidth * 0.20

	f.renderFieldBox(pdf, f.margin, y, col1b, fieldHeightSmall, "Número do documento", b.DocumentNumber, tr)
	f.renderFieldBox(pdf, f.margin+col1b, y, col2b, fieldHeightSmall, "Contrato", "", tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b, y, col3b, fieldHeightSmall, "CPF/CEI/CNPJ", FormatDocument(b.Beneficiary.Document), tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b+col3b, y, col4b, fieldHeightSmall, "Vencimento", FormatDate(b.DueDate), tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b+col3b+col4b, y, col5b, fieldHeightSmall, "Valor documento", FormatCurrency(b.Value), tr)
	y += fieldHeightSmall

	// Row 3: Desconto | Outras deduções | Mora/Multa | Outros acréscimos | Valor cobrado
	f.renderFieldBox(pdf, f.margin, y, col1b, fieldHeightSmall, "(-) Desconto / Abatimento", "", tr)
	f.renderFieldBox(pdf, f.margin+col1b, y, col2b, fieldHeightSmall, "(-) Outras deduções", "", tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b, y, col3b, fieldHeightSmall, "(+) Mora / Multa", "", tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b+col3b, y, col4b, fieldHeightSmall, "(+) Outros acréscimos", "", tr)
	f.renderFieldBox(pdf, f.margin+col1b+col2b+col3b+col4b, y, col5b, fieldHeightSmall, "(=) Valor cobrado", "", tr)
	y += fieldHeightSmall

	// Row 4: Sacado
	sacadoText := b.Payer.Name + " - " + FormatDocument(b.Payer.Document)
	if b.Payer.Address != "" {
		sacadoText += " - " + b.Payer.Address
	}
	f.renderFieldBox(pdf, f.margin, y, contentWidth*0.75, fieldHeightSmall, "Sacado", sacadoText, tr)
	f.renderFieldBox(pdf, f.margin+contentWidth*0.75, y, contentWidth*0.25, fieldHeightSmall, "Nosso número", b.OurNumber, tr)
	y += fieldHeightSmall

	// Autenticação mecânica
	pdf.SetFont("Arial", "", 6)
	pdf.SetXY(f.margin+contentWidth*0.7, y)
	pdf.Cell(contentWidth*0.3, 4, tr("Autenticação mecânica"))
	y += 6

	// Draw outer border for entire receipt section
	pdf.SetLineWidth(0.3)
	pdf.SetDrawColor(0, 0, 0)
	headerHeight := 10.0
	pdf.Rect(f.margin, startY+headerHeight, contentWidth, y-startY-headerHeight-6, "D")

	return y
}

// renderFichaCompensacao renders the main bank slip (Ficha de Compensação)
func (f *Febraban) renderFichaCompensacao(pdf *fpdf.Fpdf, data *RenderData, y float64, contentWidth float64, tr func(string) string) float64 {
	b := data.Boleto

	// Header with bank logo and digitable line
	y = f.renderBankHeader(pdf, data, y, contentWidth, tr)

	// Column widths
	leftCol := contentWidth * 0.75   // Left side (main fields)
	rightCol := contentWidth * 0.25  // Right side (values)

	// Row 1: Local de pagamento | Vencimento
	f.renderFieldBox(pdf, f.margin, y, leftCol, fieldHeight, "Local de pagamento", "QUALQUER BANCO ATÉ O VENCIMENTO", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y, rightCol, fieldHeight, "Vencimento", FormatDate(b.DueDate), tr)
	y += fieldHeight

	// Row 2: Cedente | Agência/Código cedente
	cedenteText := b.Beneficiary.Name
	if b.Beneficiary.Document != "" {
		cedenteText += " - " + FormatDocument(b.Beneficiary.Document)
	}
	agenciaCodigo := b.Agency + "-" + b.AgencyDigit + " / " + b.Account + "-" + b.AccountDigit
	f.renderFieldBox(pdf, f.margin, y, leftCol, fieldHeight, "Cedente", cedenteText, tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y, rightCol, fieldHeight, "Agência/Código cedente", agenciaCodigo, tr)
	y += fieldHeight

	// Row 3: Data documento | Nº documento | Espécie doc. | Aceite | Data process. | Nosso número
	col1 := leftCol * 0.18
	col2 := leftCol * 0.18
	col3 := leftCol * 0.14
	col4 := leftCol * 0.10
	col5 := leftCol * 0.22

	f.renderFieldBox(pdf, f.margin, y, col1, fieldHeight, "Data do documento", FormatDate(b.DocumentDate), tr)
	f.renderFieldBox(pdf, f.margin+col1, y, col2, fieldHeight, "Nº documento", b.DocumentNumber, tr)
	f.renderFieldBox(pdf, f.margin+col1+col2, y, col3, fieldHeight, "Espécie doc.", "DM", tr)
	f.renderFieldBox(pdf, f.margin+col1+col2+col3, y, col4, fieldHeight, "Aceite", "N", tr)
	f.renderFieldBox(pdf, f.margin+col1+col2+col3+col4, y, col5, fieldHeight, "Data process.", FormatDate(b.ProcessingDate), tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y, rightCol, fieldHeight, "Nosso número", b.OurNumber, tr)
	y += fieldHeight

	// Row 4: Uso do banco | Carteira | Espécie | Quantidade | x Valor | Valor documento
	col1c := leftCol * 0.15
	col2c := leftCol * 0.12
	col3c := leftCol * 0.12
	col4c := leftCol * 0.15
	col5c := leftCol * 0.46

	f.renderFieldBox(pdf, f.margin, y, col1c, fieldHeight, "Uso do banco", "", tr)
	f.renderFieldBox(pdf, f.margin+col1c, y, col2c, fieldHeight, "Carteira", "", tr)
	f.renderFieldBox(pdf, f.margin+col1c+col2c, y, col3c, fieldHeight, "Espécie", "R$", tr)
	f.renderFieldBox(pdf, f.margin+col1c+col2c+col3c, y, col4c, fieldHeight, "Quantidade", "", tr)
	f.renderFieldBox(pdf, f.margin+col1c+col2c+col3c+col4c, y, col5c, fieldHeight, "x Valor", "", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y, rightCol, fieldHeight, "(=) Valor documento", FormatCurrency(b.Value), tr)
	y += fieldHeight

	// Instructions section with deductions on the right
	instructionsHeight := 40.0

	// Instructions box (left)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.Rect(f.margin, y, leftCol, instructionsHeight, "D")

	pdf.SetFont("Arial", "", labelFontSize)
	pdf.SetXY(f.margin+1, y+1)
	pdf.Cell(leftCol-2, 3, tr("Instruções (Texto de responsabilidade do cedente)"))

	pdf.SetFont("Arial", "", valueFontSize)
	instructionY := y + 5
	for i, instruction := range b.Instructions {
		if i >= 5 {
			break // Limit to 5 instructions
		}
		pdf.SetXY(f.margin+2, instructionY)
		pdf.Cell(leftCol-4, 4, tr(instruction))
		instructionY += 4
	}

	// Deductions boxes (right side)
	deductionHeight := instructionsHeight / 5

	f.renderFieldBoxRight(pdf, f.margin+leftCol, y, rightCol, deductionHeight, "(-) Desconto / Abatimento", "", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y+deductionHeight, rightCol, deductionHeight, "(-) Outras deduções", "", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y+deductionHeight*2, rightCol, deductionHeight, "(+) Mora / Multa", "", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y+deductionHeight*3, rightCol, deductionHeight, "(+) Outros acréscimos", "", tr)
	f.renderFieldBoxRight(pdf, f.margin+leftCol, y+deductionHeight*4, rightCol, deductionHeight, "(=) Valor cobrado", "", tr)
	y += instructionsHeight

	// Sacado (Payer)
	sacadoText := b.Payer.Name + " - " + FormatDocument(b.Payer.Document)
	if b.Payer.Address != "" {
		sacadoText += "\n" + b.Payer.Address + " - " + b.Payer.City + "/" + b.Payer.State + " - " + b.Payer.PostalCode
	}

	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.Rect(f.margin, y, contentWidth, fieldHeight*2, "D")

	pdf.SetFont("Arial", "", labelFontSize)
	pdf.SetXY(f.margin+1, y+1)
	pdf.Cell(contentWidth-2, 3, tr("Sacado"))

	pdf.SetFont("Arial", "B", valueFontSize)
	pdf.SetXY(f.margin+2, y+4)
	pdf.Cell(contentWidth-4, 4, tr(b.Payer.Name+" - "+FormatDocument(b.Payer.Document)))

	pdf.SetFont("Arial", "", valueFontSize)
	if b.Payer.Address != "" {
		pdf.SetXY(f.margin+2, y+8)
		address := b.Payer.Address + " - " + b.Payer.City + "/" + b.Payer.State + " - " + b.Payer.PostalCode
		pdf.Cell(contentWidth-4, 4, tr(address))
	}
	y += fieldHeight * 2

	// Sacador/Avalista
	pdf.SetDrawColor(0, 0, 0)
	pdf.Rect(f.margin, y, contentWidth*0.75, fieldHeight, "D")
	pdf.SetFont("Arial", "", labelFontSize)
	pdf.SetXY(f.margin+1, y+1)
	pdf.Cell(contentWidth*0.75-2, 3, tr("Sacador/Avalista"))

	// Cód. baixa box
	pdf.Rect(f.margin+contentWidth*0.75, y, contentWidth*0.25, fieldHeight, "D")
	pdf.SetXY(f.margin+contentWidth*0.75+1, y+1)
	pdf.Cell(contentWidth*0.25-2, 3, tr("Cód. baixa"))
	y += fieldHeight

	// Autenticação mecânica - Ficha de Compensação
	pdf.SetFont("Arial", "", 6)
	pdf.SetXY(f.margin+contentWidth*0.5, y)
	pdf.Cell(contentWidth*0.5, 4, tr("Autenticação mecânica - Ficha de Compensação"))
	y += 6

	// PIX section (if available)
	if f.showPIXSection && data.QRCodeImage != nil {
		y = f.renderPIXSection(pdf, data, y, contentWidth, tr)
	}

	// Barcode
	y += 2
	if data.BarcodeImage != nil {
		x := f.margin
		f.addImageFromBytes(pdf, data.BarcodeImage, x, y, BarcodeWidth, BarcodeHeight)
	}
	y += BarcodeHeight + 3

	// Cut line at the end
	y = f.renderCutLine(pdf, y, contentWidth, tr)

	return y
}

// renderBankHeader renders the bank header with logo, code and digitable line
func (f *Febraban) renderBankHeader(pdf *fpdf.Fpdf, data *RenderData, y float64, contentWidth float64, tr func(string) string) float64 {
	b := data.Boleto
	headerHeight := 10.0

	// Bank logo section with yellow background
	logoWidth := 45.0
	pdf.SetFillColor(255, 204, 0) // Yellow/gold color like Banco do Brasil
	pdf.Rect(f.margin, y, logoWidth, headerHeight, "F")

	// Add bank logo if available
	if data.BankLogo != nil {
		f.addImageFromBytes(pdf, data.BankLogo, f.margin+2, y+1, logoWidth-4, headerHeight-2)
	} else {
		// Fallback: draw bank name
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(f.margin+2, y+3)
		pdf.Cell(logoWidth-4, 5, tr(b.BankName))
	}

	// Bank code with vertical separators
	codeWidth := 15.0
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.5)

	// Vertical lines around bank code
	pdf.Line(f.margin+logoWidth, y, f.margin+logoWidth, y+headerHeight)
	pdf.Line(f.margin+logoWidth+codeWidth, y, f.margin+logoWidth+codeWidth, y+headerHeight)

	// Bank code
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	bankCodeFormatted := b.BankCode
	if len(bankCodeFormatted) == 3 {
		// Add check digit (simplified - just use first digit as check)
		bankCodeFormatted = bankCodeFormatted + "-9"
	}
	pdf.SetXY(f.margin+logoWidth+1, y+2)
	pdf.Cell(codeWidth-2, 6, tr(bankCodeFormatted))

	// Digitable line
	digitableLine := FormatDigitableLine(b.DigitiableLine)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetXY(f.margin+logoWidth+codeWidth+2, y+2)
	pdf.Cell(contentWidth-logoWidth-codeWidth-4, 6, tr(digitableLine))

	// Bottom border of header
	pdf.SetLineWidth(0.5)
	pdf.Line(f.margin, y+headerHeight, f.margin+contentWidth, y+headerHeight)

	// Reset text color
	pdf.SetTextColor(0, 0, 0)

	return y + headerHeight
}

// renderCutLine renders the dashed cut line with text
func (f *Febraban) renderCutLine(pdf *fpdf.Fpdf, y float64, contentWidth float64, tr func(string) string) float64 {
	pdf.SetFont("Arial", "", 6)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(f.margin, y)
	pdf.Cell(contentWidth, 4, tr("Corte na linha pontilhada"))
	y += 4

	pdf.SetDrawColor(128, 128, 128)
	pdf.SetDashPattern([]float64{2, 1}, 0)
	pdf.SetLineWidth(0.2)
	pdf.Line(f.margin, y, f.margin+contentWidth, y)
	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetTextColor(0, 0, 0)
	y += 5

	return y
}

// renderPIXSection renders the PIX QR code section
func (f *Febraban) renderPIXSection(pdf *fpdf.Fpdf, data *RenderData, y float64, contentWidth float64, tr func(string) string) float64 {
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, tr("Pague também com PIX"))
	y += 6

	if data.QRCodeImage != nil {
		f.addImageFromBytes(pdf, data.QRCodeImage, f.margin, y, QRCodeSize, QRCodeSize)
	}

	if data.Boleto.PIX != nil && data.Boleto.PIX.TxID != "" {
		pdf.SetFont("Arial", "", 7)
		pdf.SetXY(f.margin+QRCodeSize+5, y+10)
		pdf.Cell(0, 4, tr("TxID: "+data.Boleto.PIX.TxID))
	}

	y += QRCodeSize + 5
	return y
}

// renderFieldBox renders a bordered field with label and value
func (f *Febraban) renderFieldBox(pdf *fpdf.Fpdf, x, y, width, height float64, label, value string, tr func(string) string) {
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.Rect(x, y, width, height, "D")

	// Label (small, at top)
	pdf.SetFont("Arial", "", labelFontSize)
	pdf.SetXY(x+1, y+0.5)
	pdf.Cell(width-2, 3, tr(label))

	// Value (bold, below label)
	pdf.SetFont("Arial", "B", valueFontSize)
	pdf.SetXY(x+1, y+3.5)
	pdf.Cell(width-2, height-4, tr(value))
}

// renderFieldBoxRight renders a field box with right-aligned value
func (f *Febraban) renderFieldBoxRight(pdf *fpdf.Fpdf, x, y, width, height float64, label, value string, tr func(string) string) {
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.Rect(x, y, width, height, "D")

	// Label (small, at top)
	pdf.SetFont("Arial", "", labelFontSize)
	pdf.SetXY(x+1, y+0.5)
	pdf.Cell(width-2, 3, tr(label))

	// Value (bold, right-aligned)
	pdf.SetFont("Arial", "B", valueFontSize)
	pdf.SetXY(x+1, y+3.5)
	pdf.CellFormat(width-2, height-4, tr(value), "", 0, "R", false, 0, "")
}

func (f *Febraban) addImageFromBytes(pdf *fpdf.Fpdf, imgBytes []byte, x, y, maxW, maxH float64) {
	// Decode image to get dimensions
	reader := bytes.NewReader(imgBytes)
	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return
	}

	// Calculate scaled dimensions maintaining aspect ratio
	w := float64(img.Width)
	h := float64(img.Height)
	ratio := w / h

	finalW := maxW
	finalH := maxW / ratio

	if finalH > maxH {
		finalH = maxH
		finalW = maxH * ratio
	}

	// Register and place image
	name := strconv.FormatInt(int64(x*1000+y*100), 36)
	pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgBytes))
	pdf.ImageOptions(name, x, y, finalW, finalH, false, fpdf.ImageOptions{}, 0, "")
}
