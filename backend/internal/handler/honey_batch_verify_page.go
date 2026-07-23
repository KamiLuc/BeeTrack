package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beetrack/backend/internal/model"
)

// amoyExplorerBaseURL is the Polygon Amoy testnet block explorer used to
// link a certification's transaction hash. Thesis scope is Amoy-only (see
// HONEY_BLOCKCHAIN_BACKLOG.md), so this isn't chain-id-parameterized.
const amoyExplorerBaseURL = "https://amoy.polygonscan.com"

// verifyPageStrings holds every translated label the public verification
// page needs, for one language.
type verifyPageStrings struct {
	HTMLLang              string
	Title                 string
	NotFoundMessage       string
	HoneyType             string
	ProcessingMethod      string
	GatheringDate         string
	Amount                string
	BatchID               string
	ProofTitle            string
	ProofIntro            string
	PDFHash               string
	PDFHashExplainer      string
	MetadataHash          string
	MetadataHashExplainer string
	ContractAddress       string
	BlockNumber           string
	TransactionHash       string
	ViewOnExplorer        string
	NotCertified          string
	InProgress            string
	ChainVerified         string
	ChainMismatch         string
	ChainUnavailable      string
	DownloadPDF           string
	ProcessingMethods     map[string]string
	CertStatuses          map[model.CertificationStatus]string
}

var verifyPageEN = verifyPageStrings{
	HTMLLang:              "en",
	Title:                 "Honey Batch Verification",
	NotFoundMessage:       "This verification link is invalid or the batch no longer exists.",
	HoneyType:             "Honey type",
	ProcessingMethod:      "Processing method",
	GatheringDate:         "Gathering date",
	Amount:                "Amount",
	BatchID:               "Batch ID",
	ProofTitle:            "Certification proof",
	ProofIntro:            "The two hashes below are SHA-256 checksums computed from this batch's lab PDF and metadata, then permanently written to the blockchain when it was certified. If the file or the data were changed afterward, the checksum would no longer match - the checks below compare the checksum stored here against the one recorded on-chain, live.",
	PDFHash:               "Lab PDF hash (SHA-256)",
	PDFHashExplainer:      "Checksum of the uploaded lab report.",
	MetadataHash:          "Metadata hash (SHA-256)",
	MetadataHashExplainer: "Checksum of the batch's gathering date, amount, processing method, honey type, and PDF hash together.",
	ContractAddress:       "Smart contract address",
	BlockNumber:           "Block number",
	TransactionHash:       "Transaction hash",
	ViewOnExplorer:        "View on block explorer",
	NotCertified:          "Not certified",
	InProgress:            "Certification in progress",
	ChainVerified:         "Matches the record on the blockchain",
	ChainMismatch:         "Does not match the blockchain record - data may have changed",
	ChainUnavailable:      "Live blockchain check unavailable right now",
	DownloadPDF:           "Download lab PDF",
	ProcessingMethods: map[string]string{
		"raw":         "Raw",
		"filtered":    "Filtered",
		"pasteurized": "Pasteurized",
	},
	CertStatuses: map[model.CertificationStatus]string{
		model.CertificationStatusQueued:              "Queued",
		model.CertificationStatusSubmitting:          "Submitting",
		model.CertificationStatusSubmitted:           "Submitted",
		model.CertificationStatusPendingConfirmation: "Pending confirmation",
		model.CertificationStatusConfirmed:           "Confirmed",
		model.CertificationStatusFailed:              "Failed",
		model.CertificationStatusReverted:            "Reverted",
	},
}

var verifyPagePL = verifyPageStrings{
	HTMLLang:              "pl",
	Title:                 "Weryfikacja partii miodu",
	NotFoundMessage:       "Ten link weryfikacyjny jest nieprawidłowy lub partia już nie istnieje.",
	HoneyType:             "Rodzaj miodu",
	ProcessingMethod:      "Metoda przetwarzania",
	GatheringDate:         "Data pozyskania",
	Amount:                "Ilość",
	BatchID:               "ID partii",
	ProofTitle:            "Dowód certyfikacji",
	ProofIntro:            "Poniższe dwa hashe to sumy kontrolne SHA-256 obliczone z PDF-u badania laboratoryjnego i metadanych tej partii, zapisane trwale w blockchainie podczas certyfikacji. Gdyby plik lub dane zostały później zmienione, suma kontrolna przestałaby się zgadzać - poniższe sprawdzenia porównują sumę kontrolną zapisaną tutaj z tą zapisaną w blockchainie, na żywo.",
	PDFHash:               "Hash PDF z badania (SHA-256)",
	PDFHashExplainer:      "Suma kontrolna przesłanego badania laboratoryjnego.",
	MetadataHash:          "Hash metadanych (SHA-256)",
	MetadataHashExplainer: "Suma kontrolna daty pozyskania, ilości, metody przetwarzania, rodzaju miodu i hasha PDF-u razem.",
	ContractAddress:       "Adres kontraktu",
	BlockNumber:           "Numer bloku",
	TransactionHash:       "Hash transakcji",
	ViewOnExplorer:        "Zobacz w eksploratorze bloków",
	NotCertified:          "Niecertyfikowane",
	InProgress:            "Certyfikacja w toku",
	ChainVerified:         "Zgodne z rekordem w blockchainie",
	ChainMismatch:         "Niezgodne z rekordem w blockchainie - dane mogły zostać zmienione",
	ChainUnavailable:      "Weryfikacja na żywo chwilowo niedostępna",
	DownloadPDF:           "Pobierz PDF z badania",
	ProcessingMethods: map[string]string{
		"raw":         "Surowy",
		"filtered":    "Filtrowany",
		"pasteurized": "Pasteryzowany",
	},
	CertStatuses: map[model.CertificationStatus]string{
		model.CertificationStatusQueued:              "W kolejce",
		model.CertificationStatusSubmitting:          "Wysyłanie",
		model.CertificationStatusSubmitted:           "Wysłano",
		model.CertificationStatusPendingConfirmation: "Oczekuje na potwierdzenie",
		model.CertificationStatusConfirmed:           "Certyfikowano",
		model.CertificationStatusFailed:              "Niepowodzenie",
		model.CertificationStatusReverted:            "Wycofano",
	},
}

// pickVerifyPageLanguage resolves the language for the public verification
// page: an explicit "?lang=pl"/"?lang=en" query param wins, otherwise the
// Accept-Language header is checked for a "pl" preference, defaulting to
// English.
func pickVerifyPageLanguage(r *http.Request) verifyPageStrings {
	switch r.URL.Query().Get("lang") {
	case "pl":
		return verifyPagePL
	case "en":
		return verifyPageEN
	}
	if strings.Contains(strings.ToLower(r.Header.Get("Accept-Language")), "pl") {
		return verifyPagePL
	}
	return verifyPageEN
}

// verifyPageData is the template view-model for the public verification page.
type verifyPageData struct {
	Strings verifyPageStrings
	Found   bool

	HoneyType        string
	ProcessingMethod string
	GatheringDate    string
	AmountKg         string
	BatchID          string

	PDFHash                string
	PDFHashCheckLabel      string // empty if no live check was performed
	PDFHashCheckClass      string
	MetadataHash           string
	MetadataHashCheckLabel string
	MetadataHashCheckClass string

	HasCertification bool
	StatusLabel      string
	StatusClass      string
	ContractAddress  string
	BlockNumber      string
	TransactionHash  string
	ExplorerURL      string
	PDFDownloadURL   string
}

// chainCheckState is the outcome of comparing a hash against the live
// on-chain record: empty means no check was attempted (e.g. not yet
// confirmed, or no chain reader configured).
type chainCheckState string

const (
	chainCheckMatch       chainCheckState = "match"
	chainCheckMismatch    chainCheckState = "mismatch"
	chainCheckUnavailable chainCheckState = "unavailable"
)

// label returns the display label/CSS class for state — "", "" if state is empty.
func (s chainCheckState) label(strs verifyPageStrings) (label, class string) {
	switch s {
	case chainCheckMatch:
		return strs.ChainVerified, "check-ok"
	case chainCheckMismatch:
		return strs.ChainMismatch, "check-mismatch"
	case chainCheckUnavailable:
		return strs.ChainUnavailable, "check-unavailable"
	default:
		return "", ""
	}
}

var verifyPageTemplate = template.Must(template.New("verify").Parse(`<!DOCTYPE html>
<html lang="{{.Strings.HTMLLang}}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Strings.Title}}</title>
<style>
  :root { color-scheme: light; }
  body {
    margin: 0; padding: 24px 16px 48px;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: #fafaf5; color: #1c1b16;
  }
  main { max-width: 560px; margin: 0 auto; }
  h1 { font-size: 2rem; margin: 0 0 8px; }
  .honey-type { font-size: 1.4rem; font-weight: 600; margin: 0 0 16px; color: #555; }
  .card {
    background: #fff; border-radius: 12px; padding: 16px 20px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.08); margin-bottom: 16px;
  }
  .field { margin-bottom: 12px; }
  .field:last-child { margin-bottom: 0; }
  .label {
    font-size: 0.72rem; font-weight: 600; letter-spacing: 0.4px;
    text-transform: uppercase; color: #a36b00; margin-bottom: 2px;
  }
  .explainer { font-size: 0.82rem; color: #777; margin: 0 0 6px; }
  .intro { font-size: 0.88rem; color: #555; line-height: 1.5; margin: 0 0 16px; }
  .value { font-size: 0.95rem; word-break: break-all; }
  .value.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
  .section-title { font-size: 1rem; font-weight: 600; margin: 24px 0 8px; }
  .badge {
    display: inline-block; padding: 3px 10px; border-radius: 20px;
    font-size: 0.78rem; font-weight: 600;
  }
  .badge-confirmed { background: #d4edda; color: #1e7e34; }
  .badge-progress { background: #fff3cd; color: #a36b00; }
  .badge-failed { background: #f8d7da; color: #a71d2a; }
  .badge-none { background: #eee; color: #555; }
  .check {
    display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;
    font-size: 0.78rem; font-weight: 600;
  }
  .check-ok { color: #1e7e34; }
  .check-mismatch { color: #a71d2a; }
  .check-unavailable { color: #999; font-weight: 500; }
  .button-row { display: flex; gap: 12px; margin-top: 12px; flex-wrap: wrap; }
  a.explorer-link {
    display: inline-block; padding: 8px 14px;
    border: 1px solid #a36b00; border-radius: 8px; color: #a36b00;
    text-decoration: none; font-size: 0.9rem; font-weight: 600;
    text-align: center;
  }
  .button-row a.explorer-link { flex: 1; }
  .not-found { text-align: center; padding: 64px 16px; color: #555; }
</style>
</head>
<body>
<main>
{{if not .Found}}
  <div class="not-found">{{.Strings.NotFoundMessage}}</div>
{{else}}
  <h1>{{.Strings.Title}}</h1>
  <p class="honey-type">
    {{.HoneyType}}
    <span class="badge {{.StatusClass}}">{{.StatusLabel}}</span>
  </p>

  <div class="card">
    <div class="field">
      <div class="label">{{.Strings.ProcessingMethod}}</div>
      <div class="value">{{.ProcessingMethod}}</div>
    </div>
    <div class="field">
      <div class="label">{{.Strings.GatheringDate}}</div>
      <div class="value">{{.GatheringDate}}</div>
    </div>
    <div class="field">
      <div class="label">{{.Strings.Amount}}</div>
      <div class="value">{{.AmountKg}} kg</div>
    </div>
    <div class="field">
      <div class="label">{{.Strings.BatchID}}</div>
      <div class="value mono">{{.BatchID}}</div>
    </div>
  </div>

  <div class="section-title">{{.Strings.ProofTitle}}</div>
  <p class="intro">{{.Strings.ProofIntro}}</p>
  <div class="card">
    <div class="field">
      <div class="label">{{.Strings.PDFHash}}</div>
      <div class="explainer">{{.Strings.PDFHashExplainer}}</div>
      <div class="value mono">{{.PDFHash}}</div>
      {{if .PDFHashCheckLabel}}<div class="check {{.PDFHashCheckClass}}">{{.PDFHashCheckLabel}}</div>{{end}}
    </div>
    <div class="field">
      <div class="label">{{.Strings.MetadataHash}}</div>
      <div class="explainer">{{.Strings.MetadataHashExplainer}}</div>
      <div class="value mono">{{.MetadataHash}}</div>
      {{if .MetadataHashCheckLabel}}<div class="check {{.MetadataHashCheckClass}}">{{.MetadataHashCheckLabel}}</div>{{end}}
    </div>
    {{if .HasCertification}}
    <div class="field">
      <div class="label">{{.Strings.ContractAddress}}</div>
      <div class="value mono">{{.ContractAddress}}</div>
    </div>
    <div class="field">
      <div class="label">{{.Strings.BlockNumber}}</div>
      <div class="value">{{.BlockNumber}}</div>
    </div>
    <div class="field">
      <div class="label">{{.Strings.TransactionHash}}</div>
      <div class="value mono">{{.TransactionHash}}</div>
    </div>
    <div class="button-row">
      {{if .PDFDownloadURL}}<a class="explorer-link" href="{{.PDFDownloadURL}}">{{.Strings.DownloadPDF}}</a>{{end}}
      <a class="explorer-link" href="{{.ExplorerURL}}" target="_blank" rel="noopener">{{.Strings.ViewOnExplorer}}</a>
    </div>
    {{end}}
  </div>
{{end}}
</main>
</body>
</html>
`))

// VerifyPage handles GET /verify/{token} — public, renders a self-contained,
// dependency-free HTML page (no SPA, no JS runtime required) so the link
// encoded in a batch's QR code opens directly in any browser, including one
// launched by a phone's regular camera app. Localized via "?lang=pl"/"en" or
// the Accept-Language header, defaulting to English.
func (h *HoneyBatchVerifyHandler) VerifyPage(w http.ResponseWriter, r *http.Request) {
	strs := pickVerifyPageLanguage(r)

	result, err := h.batches.GetBatchWithVerification(r.Context(), r.PathValue("token"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil || result == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = verifyPageTemplate.Execute(w, verifyPageData{Strings: strs, Found: false})
		return
	}

	batch := result.Batch
	cert := result.Certification

	data := verifyPageData{
		Strings:          strs,
		Found:            true,
		HoneyType:        batch.HoneyType,
		ProcessingMethod: processingMethodLabel(strs, batch.ProcessingMethod),
		GatheringDate:    batch.GatheringDate.Format("2006-01-02"),
		AmountKg:         formatKg(batch.AmountGrams),
		BatchID:          batch.VerificationToken,
		PDFHash:          batch.PDFFileHash,
		MetadataHash:     batch.MetadataHash,
	}

	switch {
	case cert == nil:
		data.StatusLabel = strs.NotCertified
		data.StatusClass = "badge-none"
	case !cert.Status.IsTerminal():
		data.StatusLabel = strs.InProgress
		data.StatusClass = "badge-progress"
	case cert.Status == model.CertificationStatusConfirmed:
		data.StatusLabel = strs.CertStatuses[cert.Status]
		data.StatusClass = "badge-confirmed"
	default:
		data.StatusLabel = strs.CertStatuses[cert.Status]
		data.StatusClass = "badge-failed"
	}

	if cert != nil {
		data.HasCertification = cert.Status == model.CertificationStatusConfirmed
		data.ContractAddress = cert.ContractAddress
		if cert.BlockNumber != nil {
			data.BlockNumber = formatInt64(*cert.BlockNumber)
		}
		if cert.TransactionHash != nil {
			data.TransactionHash = *cert.TransactionHash
			data.ExplorerURL = amoyExplorerBaseURL + "/tx/" + *cert.TransactionHash
		}
	}

	if data.HasCertification {
		pdfCheck, metadataCheck := h.checkAgainstChain(r.Context(), batch)
		data.PDFHashCheckLabel, data.PDFHashCheckClass = pdfCheck.label(strs)
		data.MetadataHashCheckLabel, data.MetadataHashCheckClass = metadataCheck.label(strs)
		// Same gate as the backend's own PDF endpoint (GetBatchPDFByToken):
		// only a confirmed batch's lab PDF is publicly downloadable.
		data.PDFDownloadURL = "/api/v1/verify/" + batch.VerificationToken + "/pdf"
	}

	_ = verifyPageTemplate.Execute(w, data)
}

// chainCheckTimeout bounds the live on-chain read in checkAgainstChain so a
// slow or unresponsive testnet RPC endpoint can never hang the page.
const chainCheckTimeout = 5 * time.Second

// checkAgainstChain live-compares batch's stored PDF/metadata hashes against
// the record on the smart contract, returning "" states if no chain reader
// is configured (see cmd/api/main.go) or the read fails/times out — a
// verification page must never hard-fail just because an RPC endpoint is
// briefly unavailable.
func (h *HoneyBatchVerifyHandler) checkAgainstChain(ctx context.Context, batch *model.HoneyBatch) (pdf, metadata chainCheckState) {
	if h.certReader == nil {
		return chainCheckUnavailable, chainCheckUnavailable
	}

	chainCtx, cancel := context.WithTimeout(ctx, chainCheckTimeout)
	defer cancel()

	record, err := h.certReader.GetCertification(chainCtx, batch.ID)
	if err != nil {
		return chainCheckUnavailable, chainCheckUnavailable
	}

	pdfHashHex, metadataHashHex := hexEncodeCertRecord(record)
	pdf = chainCheckMismatch
	if pdfHashHex == batch.PDFFileHash {
		pdf = chainCheckMatch
	}
	metadata = chainCheckMismatch
	if metadataHashHex == batch.MetadataHash {
		metadata = chainCheckMatch
	}
	return pdf, metadata
}

func processingMethodLabel(strs verifyPageStrings, method string) string {
	if label, ok := strs.ProcessingMethods[method]; ok {
		return label
	}
	return method
}

// formatKg formats grams as kilograms with no trailing zeros, e.g. 1500 -> "1.5", 2000 -> "2".
func formatKg(grams int64) string {
	return strconv.FormatFloat(float64(grams)/1000, 'f', -1, 64)
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
