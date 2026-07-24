package handler

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/go-pdf/fpdf"
)

//go:embed assets/fonts/DejaVuSansCondensed.ttf
var reportFontRegular []byte

//go:embed assets/fonts/DejaVuSansCondensed-Bold.ttf
var reportFontBold []byte

const reportFont = "DejaVuSansCondensed"

// renderReportPDF lays out report as a plain, single-column document —
// title block, then one numbered section per hive, with a numbered
// subsection per record category, each entry listed as its own paragraph
// stacked vertically (a PDF page can't scroll, so nothing is laid out
// side-by-side the way the in-app report cards are).
func renderReportPDF(report *service.Report) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(reportFont, "", reportFontRegular)
	pdf.AddUTF8FontFromBytes(reportFont, "B", reportFontBold)
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(reportFont, "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 10, fmt.Sprintf("BeeTrack — Strona %d z {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()

	pdf.SetFont(reportFont, "B", 20)
	pdf.CellFormat(0, 12, "Raport aktywności pasieki", "", 1, "C", false, 0, "")

	pdf.SetFont(reportFont, "B", 14)
	pdf.CellFormat(0, 8, report.Apiary.Name, "", 1, "C", false, 0, "")

	pdf.SetFont(reportFont, "", 11)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(0, 6, fmt.Sprintf("Okres: %s – %s", report.From.Format("02.01.2006"), report.To.Format("02.01.2006")), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Wygenerowano: %s", nowFormatted()), "", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(6)

	for hiveIdx, rh := range report.Hives {
		pdf.AddPage()
		renderReportHive(pdf, hiveIdx+1, rh)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderReportHive(pdf *fpdf.Fpdf, n int, rh service.ReportHive) {
	pdf.SetFont(reportFont, "B", 15)
	pdf.CellFormat(0, 10, fmt.Sprintf("%d. %s", n, rh.Hive.Name), "", 1, "L", false, 0, "")
	x, y := pdf.GetX(), pdf.GetY()
	pdf.SetDrawColor(220, 170, 60)
	pdf.SetLineWidth(0.5)
	pdf.Line(x, y, x+170, y)
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Ln(4)

	grouped := make(map[service.ReportCategory][]service.ReportEntry, len(service.ReportCategoryOrder))
	for _, e := range rh.Entries {
		grouped[e.Category] = append(grouped[e.Category], e)
	}

	if len(rh.Entries) == 0 {
		pdf.SetFont(reportFont, "", 10)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 6, "Brak wpisów w wybranym zakresie.", "", 1, "L", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(4)
		return
	}

	subIdx := 0
	for _, cat := range service.ReportCategoryOrder {
		entries := grouped[cat]
		if len(entries) == 0 {
			continue
		}
		subIdx++
		pdf.SetFont(reportFont, "B", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("%d.%d %s (%d)", n, subIdx, reportCategoryLabel(cat), len(entries)), "", 1, "L", false, 0, "")
		pdf.Ln(1)

		for i, e := range entries {
			renderReportEntry(pdf, i+1, e)
		}
		pdf.Ln(3)
	}
}

func renderReportEntry(pdf *fpdf.Fpdf, n int, e service.ReportEntry) {
	pdf.SetFont(reportFont, "B", 10)
	pdf.SetX(pdf.GetX() + 4)
	pdf.CellFormat(0, 6, fmt.Sprintf("%d) %s", n, e.Date.Format("02.01.2006 15:04")), "", 1, "L", false, 0, "")

	pdf.SetFont(reportFont, "", 10)
	for _, line := range reportEntryLines(e) {
		pdf.SetX(28)
		pdf.MultiCell(162, 5, line, "", "L", false)
	}
	pdf.Ln(2)
}

func reportCategoryLabel(c service.ReportCategory) string {
	switch c {
	case service.ReportCategoryInspections:
		return "Inspekcje"
	case service.ReportCategoryFeedings:
		return "Podkarmianie"
	case service.ReportCategoryTreatments:
		return "Leczenia"
	case service.ReportCategoryHarvests:
		return "Zbiory"
	default:
		return string(c)
	}
}

func reportEntryLines(e service.ReportEntry) []string {
	switch e.Category {
	case service.ReportCategoryInspections:
		return inspectionLines(e.Inspection)
	case service.ReportCategoryTreatments:
		return treatmentLines(e.Treatment)
	case service.ReportCategoryFeedings:
		return feedingLines(e.Feeding)
	case service.ReportCategoryHarvests:
		return harvestLines(e.Harvest)
	default:
		return nil
	}
}

func queenStatusLabel(s string) string {
	if s == "seen" {
		return "Matka widziana"
	}
	if s == "not_seen" {
		return "Matka niewidziana"
	}
	return ""
}

func broodPatternLabel(s string) string {
	switch s {
	case "excellent":
		return "Czerw: Dużo"
	case "good":
		return "Czerw: Średnio"
	case "poor":
		return "Czerw: Mało"
	case "none":
		return "Czerw: Brak"
	default:
		return ""
	}
}

func aggressivenessLabel(s string) string {
	switch s {
	case "calm":
		return "Spokojne"
	case "mild":
		return "Łagodne"
	case "aggressive":
		return "Agresywne"
	case "very_aggressive":
		return "Bardzo agresywne"
	default:
		return ""
	}
}

func nowFormatted() string {
	return time.Now().Format("02.01.2006 15:04")
}

func inspectionLines(i *model.Inspection) []string {
	var lines []string

	obs := []string{}
	if s := queenStatusLabel(i.QueenStatus); s != "" {
		obs = append(obs, s)
	}
	if s := broodPatternLabel(i.BroodPattern); s != "" {
		obs = append(obs, s)
	}
	if s := aggressivenessLabel(i.Aggressiveness); s != "" {
		obs = append(obs, s)
	}
	if len(obs) > 0 {
		lines = append(lines, "Obserwacje: "+strings.Join(obs, " · "))
	}

	frames := []string{}
	if i.FramesBrood != nil {
		frames = append(frames, fmt.Sprintf("Ramki z czerwiem: %d", *i.FramesBrood))
	}
	if i.FramesFeed != nil {
		frames = append(frames, fmt.Sprintf("Ramki z pokarmem: %d", *i.FramesFeed))
	}
	if i.FramesPollen != nil {
		frames = append(frames, fmt.Sprintf("Ramki z pyłkiem: %d", *i.FramesPollen))
	}
	if len(frames) > 0 {
		lines = append(lines, "Ramki: "+strings.Join(frames, " · "))
	}

	if i.QueenCellsCount != nil && *i.QueenCellsCount > 0 {
		lines = append(lines, fmt.Sprintf("Mateczniki: %d", *i.QueenCellsCount))
	}
	if i.QueenAdded {
		lines = append(lines, "Poddano matkę")
	}
	if i.Notes != "" {
		lines = append(lines, "Notatka: "+i.Notes)
	}
	return lines
}

func treatmentLines(t *model.Treatment) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Preparat: %s · %s", t.MedicineName, t.Dose))
	if t.Notes != "" {
		lines = append(lines, "Notatka: "+t.Notes)
	}
	return lines
}

func feedingLines(f *model.Feeding) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Pokarm: %s · %s", f.FeedType, f.Amount))
	if f.Notes != "" {
		lines = append(lines, "Notatka: "+f.Notes)
	}
	return lines
}

func harvestLines(h *model.Harvest) []string {
	var lines []string
	frames := fmt.Sprintf("%d ramek", h.Frames)
	if h.HalfFrames > 0 {
		frames += fmt.Sprintf(" + %d półramek", h.HalfFrames)
	}
	lines = append(lines, "Ramki: "+frames)
	lines = append(lines, fmt.Sprintf("Kilogramy: %.2f kg", h.Kilograms))
	if h.Notes != "" {
		lines = append(lines, "Notatka: "+h.Notes)
	}
	return lines
}
