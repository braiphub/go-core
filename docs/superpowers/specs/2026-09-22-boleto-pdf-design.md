# Boleto PDF Generator - Design Specification

**Data:** 2026-09-22
**Status:** Aprovado
**Autor:** Claude + Matheus Carmo

## Visão Geral

Biblioteca Go para geração de boletos bancários em PDF seguindo o padrão FEBRABAN, com suporte a PIX híbrido. Será integrada ao repositório `go-core` como um novo pacote.

## Requisitos

### Funcionais

- Gerar PDF de boleto no padrão FEBRABAN
- Suportar dados completos: beneficiário, pagador, valores, datas, instruções
- Gerar código de barras ITF-25 a partir da linha digitável (44 dígitos)
- Gerar QR Code PIX a partir da string EMV (opcional)
- Validar dados de entrada (CPF/CNPJ, dígitos verificadores, formatos)
- Suportar customização de layout (cores, logos, seções)
- Suportar dois bancos inicialmente: Banco do Brasil (001) e banco white-label via Swap

### Não-Funcionais

- Sem dependências CGO
- Licença MIT (alinhada com go-core)
- Padrões do go-core: interfaces, functional options, context-aware
- Testes unitários + exemplos executáveis

## Arquitetura

### Estrutura do Pacote

```
boleto/
├── go.mod
├── go.sum
├── README.md
├── doc.go                    # Documentação do pacote
│
├── boleto.go                 # Struct Boleto principal
├── generator.go              # Generator (orquestrador)
├── options.go                # Functional options
├── errors.go                 # Erros tipados
│
├── validator/
│   ├── validator.go          # Interface ValidatorI + implementação
│   ├── cpfcnpj.go            # Validação CPF/CNPJ
│   ├── barcode.go            # Validação código de barras
│   ├── digitable_line.go     # Validação linha digitável
│   └── validator_test.go
│
├── barcode/
│   ├── barcode.go            # Interface BarcodeGeneratorI
│   ├── itf.go                # Geração ITF-25
│   ├── qrcode.go             # Geração QR Code PIX
│   └── barcode_test.go
│
├── renderer/
│   ├── renderer.go           # Interface RendererI
│   ├── febraban.go           # Implementação layout FEBRABAN
│   ├── options.go            # Options do renderer
│   ├── components.go         # Componentes reutilizáveis
│   └── renderer_test.go
│
└── example/
    └── main.go               # Exemplo executável
```

### Responsabilidades dos Componentes

| Componente | Responsabilidade |
|------------|------------------|
| `boleto.go` | Struct de dados, sem lógica |
| `generator.go` | Orquestra validator → barcode → renderer |
| `validator/` | Valida dados de entrada |
| `barcode/` | Gera imagens de código de barras e QR Code |
| `renderer/` | Renderiza o PDF final |

## Modelagem de Dados

### Struct Principal

```go
package boleto

import "time"

// Boleto representa os dados completos de um boleto FEBRABAN
type Boleto struct {
    // Identificação
    Barcode        string  // Código de barras (44 dígitos)
    DigitiableLine string  // Linha digitável (47 dígitos)
    OurNumber      string  // Nosso número
    DocumentNumber string  // Número do documento

    // Banco
    BankCode    string // Código do banco (3 dígitos)
    BankName    string // Nome do banco
    Agency      string // Agência
    AgencyDigit string // Dígito da agência
    Account     string // Conta
    AccountDigit string // Dígito da conta

    // Beneficiário (quem recebe)
    Beneficiary Beneficiary

    // Pagador (quem paga)
    Payer Payer

    // Valores e datas
    Value          float64   // Valor do documento
    DueDate        time.Time // Data de vencimento
    DocumentDate   time.Time // Data do documento
    ProcessingDate time.Time // Data de processamento

    // PIX (opcional)
    PIX *PIXInfo

    // Instruções (texto livre)
    Instructions []string

    // Demonstrativo (descrição do que está sendo cobrado)
    Description []string
}

type Beneficiary struct {
    Name       string
    Document   string // CPF ou CNPJ
    Address    string
    City       string
    State      string
    PostalCode string
}

type Payer struct {
    Name       string
    Document   string // CPF ou CNPJ
    Address    string
    City       string
    State      string
    PostalCode string
}

type PIXInfo struct {
    EMV  string // String EMV para gerar QR Code
    TxID string // ID da transação (opcional, para exibição)
}
```

## Interfaces

### ValidatorI

```go
package validator

import "context"

type ValidatorI interface {
    Validate(ctx context.Context, b *boleto.Boleto) error
}

type ValidationError struct {
    Field   string
    Value   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Value)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string { /* ... */ }
```

### GeneratorI (Barcode)

```go
package barcode

import "context"

type GeneratorI interface {
    GenerateITF(ctx context.Context, code string) ([]byte, error)
    GenerateQRCode(ctx context.Context, emv string) ([]byte, error)
}
```

### RendererI

```go
package renderer

import "context"

type RendererI interface {
    Render(ctx context.Context, data *RenderData) ([]byte, error)
}

type RenderData struct {
    Boleto       *boleto.Boleto
    BarcodeImage []byte
    QRCodeImage  []byte
    BankLogo     []byte
}
```

## Generator (Orquestrador)

```go
package boleto

import "context"

type Generator struct {
    validator ValidatorI
    barcode   barcode.GeneratorI
    renderer  renderer.RendererI
    bankLogos map[string][]byte
}

func NewGenerator(opts ...Option) (*Generator, error)

func (g *Generator) Generate(ctx context.Context, b *Boleto) ([]byte, error)

func (g *Generator) GenerateFile(ctx context.Context, b *Boleto, filepath string) error
```

### Fluxo de Geração

1. **Validar** — `validator.Validate(ctx, boleto)`
2. **Gerar código de barras** — `barcode.GenerateITF(ctx, boleto.Barcode)`
3. **Gerar QR Code PIX** — `barcode.GenerateQRCode(ctx, boleto.PIX.EMV)` (se houver)
4. **Buscar logo do banco** — `bankLogos[boleto.BankCode]`
5. **Renderizar PDF** — `renderer.Render(ctx, renderData)`

## Functional Options

### Generator Options

```go
package boleto

type Option func(*Generator)

func WithValidator(v validator.ValidatorI) Option
func WithBarcodeGenerator(b barcode.GeneratorI) Option
func WithRenderer(r renderer.RendererI) Option
func WithBankLogo(bankCode string, logo []byte) Option
```

### Renderer Options

```go
package renderer

type FebOption func(*Febraban)

func WithPrimaryColor(hex string) FebOption
func WithMargin(mm float64) FebOption
func WithoutReceipt() FebOption
func WithoutPIXSection() FebOption
func WithPageSize(width, height float64) FebOption
```

## Renderer FEBRABAN

### Estrutura do Layout

O layout padrão FEBRABAN inclui:

1. **Recibo do Pagador** (parte destacável superior)
2. **Linha de corte**
3. **Header** (logo do banco + código + agência/conta)
4. **Dados do Beneficiário**
5. **Dados do Documento** (datas, valores, nosso número)
6. **Dados do Pagador**
7. **Instruções**
8. **Seção PIX** (QR Code, quando disponível)
9. **Código de Barras**
10. **Linha Digitável**

### Customizações Suportadas

| Customização | Option | Padrão |
|--------------|--------|--------|
| Cor principal | `WithPrimaryColor(hex)` | `#000000` |
| Margem | `WithMargin(mm)` | `10mm` |
| Tamanho página | `WithPageSize(w, h)` | A4 (210x297mm) |
| Recibo do pagador | `WithoutReceipt()` | Exibido |
| Seção PIX | `WithoutPIXSection()` | Exibida quando há PIX |

## Erros Tipados

### Erros Gerais

```go
var (
    ErrNilBoleto        = errors.New("boleto não pode ser nil")
    ErrGenerateFailed   = errors.New("falha ao gerar PDF")
    ErrBankLogoNotFound = errors.New("logo do banco não encontrada")
)
```

### Erros de Validação

```go
var (
    // Campos obrigatórios
    ErrBarcodeRequired       = errors.New("código de barras é obrigatório")
    ErrDigitableLineRequired = errors.New("linha digitável é obrigatória")
    ErrBankCodeRequired      = errors.New("código do banco é obrigatório")
    ErrBeneficiaryRequired   = errors.New("beneficiário é obrigatório")
    ErrPayerRequired         = errors.New("pagador é obrigatório")
    ErrValueRequired         = errors.New("valor é obrigatório")
    ErrDueDateRequired       = errors.New("data de vencimento é obrigatória")

    // Formato inválido
    ErrInvalidBarcode         = errors.New("código de barras inválido (deve ter 44 dígitos)")
    ErrInvalidDigitableLine   = errors.New("linha digitável inválida (deve ter 47 dígitos)")
    ErrInvalidBarcodeChecksum = errors.New("dígito verificador do código de barras inválido")
    ErrInvalidCPF             = errors.New("CPF inválido")
    ErrInvalidCNPJ            = errors.New("CNPJ inválido")
    ErrInvalidBankCode        = errors.New("código do banco inválido (deve ter 3 dígitos)")

    // PIX
    ErrInvalidPIXEMV = errors.New("string EMV do PIX inválida")
)
```

### Erros de Barcode

```go
var (
    ErrInvalidITFCode   = errors.New("código inválido para ITF-25")
    ErrITFOddLength     = errors.New("ITF-25 requer quantidade par de dígitos")
    ErrQRCodeGeneration = errors.New("falha ao gerar QR Code")
    ErrEmptyEMV         = errors.New("string EMV não pode ser vazia")
)
```

## Dependências

```go
// go.mod

module github.com/braiphub/go-core/boleto

go 1.24

require (
    github.com/go-pdf/fpdf v0.9.0       // Geração de PDF
    github.com/boombuler/barcode v1.0.1 // ITF-25 e QR Code
    github.com/pkg/errors v0.9.1        // Error wrapping
)

require (
    github.com/stretchr/testify v1.8.4  // Testes
)
```

### Justificativa

| Dependência | Motivo |
|-------------|--------|
| `go-pdf/fpdf` | Sem CGO, API simples, controle preciso de layout, MIT |
| `boombuler/barcode` | Suporta ITF-25 e QR Code, amplamente usada |
| `pkg/errors` | Já usado no go-core para error wrapping |
| `stretchr/testify` | Já usado no go-core para testes |

## Testes

### Estrutura

```
boleto/
├── validator/
│   ├── validator_test.go
│   ├── cpfcnpj_test.go
│   └── barcode_test.go
│
├── barcode/
│   └── barcode_test.go
│
├── renderer/
│   └── renderer_test.go
│
└── generator_test.go
```

### Cobertura Esperada

| Componente | Testes |
|------------|--------|
| `validator/` | CPF válido/inválido, CNPJ válido/inválido, dígito verificador barcode, campos obrigatórios |
| `barcode/` | Geração ITF-25 (verifica PNG válido), geração QR Code, erros para input inválido |
| `renderer/` | PDF não vazio, magic bytes corretos |
| `generator` | Integração completa, boleto com/sem PIX |

## Exemplo de Uso

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/braiphub/go-core/boleto"
    "github.com/braiphub/go-core/boleto/renderer"
)

func main() {
    // Carregar logos dos bancos
    logoBB, _ := os.ReadFile("assets/logo_bb.png")
    logoMeuBanco, _ := os.ReadFile("assets/logo_meubanco.png")

    // Criar generator com customizações
    gen, err := boleto.NewGenerator(
        boleto.WithBankLogo("001", logoBB),
        boleto.WithBankLogo("999", logoMeuBanco),
        boleto.WithRenderer(renderer.NewFebraban(
            renderer.WithPrimaryColor("#003366"),
        )),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Criar boleto
    b := &boleto.Boleto{
        Barcode:        "00190000090300009802889500012000174720000010000",
        DigitiableLine: "00190.00009 03000.098028 89500.012009 1 74720000010000",
        OurNumber:      "12345678",
        DocumentNumber: "DM-001",

        BankCode: "001",
        BankName: "Banco do Brasil",
        Agency:   "1234",
        Account:  "56789",

        Beneficiary: boleto.Beneficiary{
            Name:       "Minha Empresa LTDA",
            Document:   "12.345.678/0001-90",
            Address:    "Rua das Flores, 100",
            City:       "São Paulo",
            State:      "SP",
            PostalCode: "01234-567",
        },

        Payer: boleto.Payer{
            Name:       "João da Silva",
            Document:   "123.456.789-00",
            Address:    "Av. Brasil, 500",
            City:       "Rio de Janeiro",
            State:      "RJ",
            PostalCode: "20000-000",
        },

        Value:          100.00,
        DueDate:        time.Date(2024, 12, 15, 0, 0, 0, 0, time.Local),
        DocumentDate:   time.Now(),
        ProcessingDate: time.Now(),

        PIX: &boleto.PIXInfo{
            EMV:  "00020126580014br.gov.bcb.pix...",
            TxID: "ABC123",
        },

        Instructions: []string{
            "Não receber após o vencimento",
            "Multa de 2% após o vencimento",
            "Juros de 1% ao mês",
        },

        Description: []string{
            "Referente à mensalidade de Dezembro/2024",
        },
    }

    // Gerar PDF
    ctx := context.Background()

    if err := gen.GenerateFile(ctx, b, "boleto.pdf"); err != nil {
        log.Fatal(err)
    }

    log.Println("Boleto gerado: boleto.pdf")
}
```

## Decisões de Design

| Decisão | Justificativa |
|---------|---------------|
| Componentes modulares | Testabilidade isolada, separação de responsabilidades |
| Struct tipada (não map) | Type safety, autocomplete, documentação implícita |
| Logos via bytes | Flexibilidade, sem acoplamento, atualizações sem rebuild da lib |
| Apenas pt-BR | Boleto é instrumento exclusivamente brasileiro |
| Instruções como texto livre | Máxima flexibilidade, sem restringir formatação |
| gofpdf | Sem CGO, controle preciso, MIT, madura |
| Validação completa | Feedback claro de erros antes da geração |

## Limitações e Escopo Futuro

### Fora do Escopo Atual

- Geração em lote (batch)
- Saída em HTML ou imagem
- Layouts específicos por banco
- Multi-idioma

### Possíveis Evoluções

- Adicionar mais bancos conforme necessidade
- Suportar carnê (múltiplos boletos por PDF)
- Modo "preview" com imagem PNG
