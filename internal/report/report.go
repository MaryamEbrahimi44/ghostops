package report

import (
	"encoding/json"
	"fmt"
	"ghostops/internal/model"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type Reporter interface {
	Render(r model.Report, w io.Writer) error
}

type TerminalReporter struct{}

func NewTerminalReporter() *TerminalReporter {
	return &TerminalReporter{}
}

func (tr *TerminalReporter) Render(r model.Report, w io.Writer) error {
	var b strings.Builder

	b.WriteString("\n================================================================================\n")
	b.WriteString("                          GHOSTOPS ANALYSIS REPORT                              \n")
	b.WriteString("================================================================================\n")
	b.WriteString(fmt.Sprintf("Log File    : %s\n", r.LogFile))
	b.WriteString(fmt.Sprintf("Generated At: %s\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString("--------------------------------------------------------------------------------\n\n")

	totalLines := r.Stats.TotalValid + r.Stats.TotalMalformed
	b.WriteString("OVERVIEW\n")
	b.WriteString(fmt.Sprintf("  Total Processed Lines : %d\n", totalLines))
	b.WriteString(fmt.Sprintf("  Valid Log Entries     : %d\n", r.Stats.TotalValid))
	b.WriteString(fmt.Sprintf("  Malformed Lines       : %d\n\n", r.Stats.TotalMalformed))

	b.WriteString("LOG SEVERITY BREAKDOWN\n")
	levels := []string{model.LevelInfo, model.LevelWarn, model.LevelError, model.LevelFatal}
	for _, lvl := range levels {
		count := r.Stats.ByLevel[lvl]
		pct := 0.0
		if r.Stats.TotalValid > 0 {
			pct = (float64(count) / float64(r.Stats.TotalValid)) * 100
		}
		b.WriteString(fmt.Sprintf("  %-6s : %6d (%5.1f%%)\n", lvl, count, pct))
	}
	b.WriteString("\n")

	b.WriteString("SERVICE BREAKDOWN\n")
	if len(r.Stats.ByService) == 0 {
		b.WriteString("  No service activity recorded.\n")
	} else {
		serviceNames := make([]string, 0, len(r.Stats.ByService))
		for svc := range r.Stats.ByService {
			serviceNames = append(serviceNames, svc)
		}
		sort.Strings(serviceNames)

		for _, svc := range serviceNames {
			count := r.Stats.ByService[svc]
			b.WriteString(fmt.Sprintf("  %-25s : %6d entries\n", svc, count))
		}
	}
	b.WriteString("\n")

	b.WriteString("DETECTED INCIDENTS\n")
	if len(r.Incidents) == 0 {
		b.WriteString("  No incidents detected.\n")
	} else {
		b.WriteString(fmt.Sprintf("  %-30s %-22s %-8s %-6s %s\n", "RULE", "SERVICE", "SEVERITY", "COUNT", "TIME RANGE"))
		b.WriteString("  " + strings.Repeat("-", 100) + "\n")

		for _, inc := range r.Incidents {
			timeRange := fmt.Sprintf("%s -> %s",
				inc.FirstSeen.Format("15:04:05"),
				inc.LastSeen.Format("15:04:05"),
			)
			b.WriteString(fmt.Sprintf("  %-30s %-22s %-8s %-6d %s\n",
				inc.RuleName,
				inc.Service,
				inc.TopLevel,
				inc.Count,
				timeRange,
			))
		}
	}

	b.WriteString("\n================================================================================\n")

	_, err := fmt.Fprint(w, b.String())
	return err
}

type JSONReporter struct{}

func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

func (jr *JSONReporter) Render(r model.Report, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}

func ExportJSON(filePath string, r model.Report) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report output file %q: %w", filePath, err)
	}
	defer file.Close()

	return NewJSONReporter().Render(r, file)
}
