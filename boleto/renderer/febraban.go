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
	LogoMaxHeight = 15.0
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

// Render generates the boleto PDF.
func (f *Febraban) Render(ctx context.Context, data *RenderData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(f.margin, f.margin, f.margin)
	pdf.AddPage()

	// Create UTF-8 to cp1252 translator for Portuguese text
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	y := f.margin

	// Receipt section (top)
	if f.showReceipt {
		y = f.renderReceipt(pdf, data, y, tr)
		y = f.renderCutLine(pdf, y)
	}

	// Main boleto section
	y = f.renderHeader(pdf, data, y, tr)
	y = f.renderBeneficiarySection(pdf, data, y, tr)
	y = f.renderAmountSection(pdf, data, y, tr)
	y = f.renderPayerSection(pdf, data, y, tr)
	y = f.renderInstructions(pdf, data, y, tr)

	// PIX section
	if f.showPIXSection && data.QRCodeImage != nil {
		y = f.renderPIXSection(pdf, data, y, tr)
	}

	// Barcode and digitable line
	y = f.renderBarcode(pdf, data, y)
	f.renderDigitableLine(pdf, data, y, tr)

	// Generate PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (f *Febraban) renderReceipt(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, tr("RECIBO DO PAGADOR"))
	y += 7

	pdf.SetFont("Arial", "", 8)
	b := data.Boleto

	// Beneficiary info
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, tr("Beneficiário: "+b.Beneficiary.Name+" - "+FormatDocument(b.Beneficiary.Document)))
	y += 5

	// Payer info
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, tr("Pagador: "+b.Payer.Name+" - "+FormatDocument(b.Payer.Document)))
	y += 5

	// Value and due date
	pdf.SetXY(f.margin, y)
	pdf.Cell(90, 4, tr("Vencimento: "+FormatDate(b.DueDate)))
	pdf.Cell(0, 4, tr("Valor: "+FormatCurrency(b.Value)))
	y += 5

	// Nosso número
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, tr("Nosso Número: "+b.OurNumber))
	y += 10

	return y
}

func (f *Febraban) renderCutLine(pdf *fpdf.Fpdf, y float64) float64 {
	pdf.SetDrawColor(128, 128, 128)
	pdf.SetDashPattern([]float64{2, 2}, 0)
	pdf.Line(f.margin, y, f.pageWidth-f.margin, y)
	pdf.SetDashPattern([]float64{}, 0)
	y += 5
	return y
}

func (f *Febraban) renderHeader(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	b := data.Boleto

	// Bank logo
	if data.BankLogo != nil {
		f.addImageFromBytes(pdf, data.BankLogo, f.margin, y, LogoMaxWidth, LogoMaxHeight)
	}

	// Bank code and name
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(f.margin+LogoMaxWidth+5, y)
	pdf.Cell(20, 8, tr(b.BankCode))

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, tr(b.BankName))

	y += 12

	// Separator line
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.5)
	pdf.Line(f.margin, y, f.pageWidth-f.margin, y)
	y += 3

	return y
}

func (f *Febraban) renderBeneficiarySection(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	b := data.Boleto
	pdf.SetFont("Arial", "", 7)
	colWidth := (f.pageWidth - 2*f.margin) / 3

	// Row 1: Beneficiário | Agência/Código | Nosso Número
	f.renderField(pdf, f.margin, y, colWidth*2, "Beneficiário", b.Beneficiary.Name, tr)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Agência/Código Beneficiário",
		b.Agency+"-"+b.AgencyDigit+" / "+b.Account+"-"+b.AccountDigit, tr)
	y += 10

	// Row 2: Document | Nosso Número | Vencimento
	f.renderField(pdf, f.margin, y, colWidth, "CPF/CNPJ", FormatDocument(b.Beneficiary.Document), tr)
	f.renderField(pdf, f.margin+colWidth, y, colWidth, "Nosso Número", b.OurNumber, tr)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Vencimento", FormatDate(b.DueDate), tr)
	y += 10

	return y
}

func (f *Febraban) renderAmountSection(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	b := data.Boleto
	colWidth := (f.pageWidth - 2*f.margin) / 4

	// Dates and amounts
	f.renderField(pdf, f.margin, y, colWidth, "Data Documento", FormatDate(b.DocumentDate), tr)
	f.renderField(pdf, f.margin+colWidth, y, colWidth, "Nº Documento", b.DocumentNumber, tr)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Espécie", "R$", tr)
	f.renderField(pdf, f.margin+colWidth*3, y, colWidth, "Valor Documento", FormatCurrency(b.Value), tr)
	y += 10

	return y
}

func (f *Febraban) renderPayerSection(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	b := data.Boleto

	f.renderField(pdf, f.margin, y, f.pageWidth-2*f.margin, "Pagador",
		b.Payer.Name+" - "+FormatDocument(b.Payer.Document), tr)
	y += 8

	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(f.margin, y)
	address := b.Payer.Address + " - " + b.Payer.City + "/" + b.Payer.State + " - " + b.Payer.PostalCode
	pdf.Cell(0, 4, tr(address))
	y += 8

	return y
}

func (f *Febraban) renderInstructions(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	b := data.Boleto

	pdf.SetFont("Arial", "B", 7)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, tr("Instruções"))
	y += 5

	pdf.SetFont("Arial", "", 7)
	for _, instruction := range b.Instructions {
		pdf.SetXY(f.margin, y)
		pdf.Cell(0, 4, tr("- "+instruction))
		y += 4
	}
	y += 5

	return y
}

func (f *Febraban) renderPIXSection(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) float64 {
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, tr("Pague com PIX"))
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

func (f *Febraban) renderBarcode(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	if data.BarcodeImage != nil {
		// Center barcode
		x := (f.pageWidth - BarcodeWidth) / 2
		f.addImageFromBytes(pdf, data.BarcodeImage, x, y, BarcodeWidth, BarcodeHeight)
	}
	y += BarcodeHeight + 3
	return y
}

func (f *Febraban) renderDigitableLine(pdf *fpdf.Fpdf, data *RenderData, y float64, tr func(string) string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, tr(FormatDigitableLine(data.Boleto.DigitiableLine)))
}

func (f *Febraban) renderField(pdf *fpdf.Fpdf, x, y, width float64, label, value string, tr func(string) string) {
	pdf.SetFont("Arial", "", 6)
	pdf.SetXY(x, y)
	pdf.Cell(width, 3, tr(label))

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(x, y+3)
	pdf.Cell(width, 5, tr(value))

	// Border
	pdf.Rect(x, y, width, 9, "D")
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
