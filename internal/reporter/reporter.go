package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/AshitomW/Drifty/internal/models"
	"gopkg.in/yaml.v3"
)

type Format string 

const (
	FormatJson Format = "json"
	FormatYaml Format = "yaml"
	FormatTable Format = "table"
	FormatText Format = "text"
)


type Reporter struct {
	format Format
	writer io.Writer 
}


func New(format Format, writer io.Writer) *Reporter{
	return &Reporter{
		format: format,
		writer: writer,
	}
}



func (r *Reporter) Generate(report *models.DriftReport) error {

	switch r.format {
		case FormatJson:
			return r.generateJSON(report)
		case FormatYaml:
			return r.generateYAML(report)
		case FormatTable:
			return r.generateTable(report)
		case FormatText:
			return r.generateText(report)
		default:
			return r.generateJSON(report);
	}
}


func (r *Reporter) generateJSON(report *models.DriftReport) error{
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent(""," ")
	return encoder.Encode(report)
}


func (r *Reporter) generateYAML(report *models.DriftReport) error {
	encoder := yaml.NewEncoder(r.writer)
	encoder.SetIndent((2))
	return encoder.Encode(report)
}



func (r *Reporter) generateTable(report *models.DriftReport) error {

	w := tabwriter.NewWriter(r.writer,0,0,2,' ',0)


	// Headers 

	fmt.Fprintf(w,"\n%s\n",strings.Repeat("=",80))
	fmt.Fprintf(w,"Environment Drift Report\n")
	fmt.Fprintf(w,"%s\n\n",strings.Repeat("=",80))


	// Showing Summary


	fmt.Fprintf(w,"Source:\t%s\n",report.SourceEnv)
	fmt.Fprintf(w,"Target:\t%s\n",report.TargetEnv)
	fmt.Fprintf(w,"Generated:\t%s\n",report.Timestamp.Format("2006-01-02 15:04:05 UTC"))
  fmt.Fprintf(w, "Has Drift\t%v\n", report.HasDrift)
	fmt.Fprintf(w,"\n")


	// Statistics

	fmt.Fprintf(w, "SUMMARY\n")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 40))
	fmt.Fprintf(w, "Total Drifts:\t%d\n", report.Summary.TotalDrifts)
	fmt.Fprintf(w, "Critical:\t%d\n", report.Summary.CriticalCount)
	fmt.Fprintf(w, "Warning:\t%d\n", report.Summary.WarningCount)
	fmt.Fprintf(w, "Info:\t%d\n", report.Summary.InfoCount)
	fmt.Fprintf(w, "\n")


		// By Category
		fmt.Fprintf(w, "By Category:\n")
		for cat, count := range report.Summary.ByCategory {
			fmt.Fprintf(w, "  %s:\t%d\n", cat, count)
		}
		fmt.Fprintf(w, "\n")

		// Drifts table: build an ASCII table with separators and truncation
		// Flush tabwriter output first so ordering is preserved
		if err := w.Flush(); err != nil {
			return err
		}

		// column widths
		const (
			wSeverity = 10
			wCategory = 12
			wType = 10
			wName = 30
			wSource = 30
			wTarget = 30
			wMessage = 40
		)

		padTrunc := func(s string, width int) string {
			if len(s) > width {
				if width > 3 {
					return s[:width-3] + "..."
				}
				return s[:width]
			}
			return s + strings.Repeat(" ", width-len(s))
		}

		makeSep := func() string {
			parts := []string{
				strings.Repeat("-", wSeverity),
				strings.Repeat("-", wCategory),
				strings.Repeat("-", wType),
				strings.Repeat("-", wName),
				strings.Repeat("-", wSource),
				strings.Repeat("-", wTarget),
				strings.Repeat("-", wMessage),
			}
			return "+" + strings.Join(parts, "+") + "+\n"
		}

		var sb strings.Builder
		sb.WriteString("DRIFTS\n")
		sb.WriteString(makeSep())
		// header
		sb.WriteString("|" + padTrunc("SEVERITY", wSeverity) + "|" + padTrunc("CATEGORY", wCategory) + "|" + padTrunc("TYPE", wType) + "|" + padTrunc("NAME", wName) + "|" + padTrunc("SOURCE", wSource) + "|" + padTrunc("TARGET", wTarget) + "|" + padTrunc("MESSAGE", wMessage) + "|\n")
		sb.WriteString(makeSep())

		for _, drift := range report.Drifts {
			sev := tableSeverityLabel(drift.Severity)
			name := padTrunc(drift.Name, wName)
			srcStr := ""
			tgtStr := ""
			if drift.SourceVal != nil {
				srcStr = padTrunc(formatValue(drift.SourceVal), wSource)
			} else {
				srcStr = padTrunc("", wSource)
			}
			if drift.TargetVal != nil {
				tgtStr = padTrunc(formatValue(drift.TargetVal), wTarget)
			} else {
				tgtStr = padTrunc("", wTarget)
			}
			msg := padTrunc(drift.Message, wMessage)

			sb.WriteString("|" + padTrunc(sev, wSeverity) + "|" + padTrunc(drift.Category, wCategory) + "|" + padTrunc(drift.Type, wType) + "|" + name + "|" + srcStr + "|" + tgtStr + "|" + msg + "|\n")
		}
		sb.WriteString(makeSep())

		if _, err := r.writer.Write([]byte(sb.String())); err != nil {
			return err
		}





	fmt.Fprintf(w, "\n")
	return w.Flush()




}




func (r *Reporter) generateText(report *models.DriftReport) error {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n")
	sb.WriteString("ENVIRONMENT DRIFT REPORT\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n\n")

	sb.WriteString(fmt.Sprintf("Report ID:  %s\n", report.ID))
	sb.WriteString(fmt.Sprintf("Generated:  %s\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("Source:     %s (%s)\n", report.SourceEnv, report.SourceSnapshot))
	sb.WriteString(fmt.Sprintf("Target:     %s (%s)\n", report.TargetEnv, report.TargetSnapshot))
	sb.WriteString("\n")

	// Summary Box
	sb.WriteString("┌" + strings.Repeat("─", 38) + "┐\n")
	sb.WriteString(fmt.Sprintf("│ %-36s │\n", "SUMMARY"))
	sb.WriteString("├" + strings.Repeat("─", 38) + "┤\n")
	sb.WriteString(fmt.Sprintf("│ Total Drifts: %-21d │\n", report.Summary.TotalDrifts))
	sb.WriteString(fmt.Sprintf("│ Critical:     %-21d │\n", report.Summary.CriticalCount))
	sb.WriteString(fmt.Sprintf("│ Warning:      %-21d │\n", report.Summary.WarningCount))
	sb.WriteString(fmt.Sprintf("│ Info:         %-21d │\n", report.Summary.InfoCount))
	sb.WriteString("└" + strings.Repeat("─", 38) + "┘\n\n")

	// Group drifts by category
	categories := map[string][]models.DriftItem{
		"file":    {},
		"envvar":  {},
		"package": {},
		"service": {},
	}

	for _, drift := range report.Drifts {
		categories[drift.Category] = append(categories[drift.Category], drift)
	}

	// Print each category
	categoryNames := map[string]string{
		"file":    "FILES",
		"envvar":  "ENVIRONMENT VARIABLES",
		"package": "PACKAGES",
		"service": "SERVICES",
	}

	for cat, name := range categoryNames {
		drifts := categories[cat]
		if len(drifts) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n%s (%d drifts)\n", name, len(drifts)))
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		for _, drift := range drifts {
			icon := getTypeIcon(drift.Type)
			severityIcon := getSeverityIcon(drift.Severity)
			sb.WriteString(fmt.Sprintf("%s %s [%s] %s\n", icon, severityIcon, drift.Type, drift.Name))
			sb.WriteString(fmt.Sprintf("    %s\n", drift.Message))
			if drift.SourceVal != nil && drift.TargetVal != nil {
				sb.WriteString(fmt.Sprintf("    Source: %v\n", formatValue(drift.SourceVal)))
				sb.WriteString(fmt.Sprintf("    Target: %v\n", formatValue(drift.TargetVal)))
			}
			sb.WriteString("\n")
		}
	}

	_, err := r.writer.Write([]byte(sb.String()))
	return err
}




func colorSeverity(severity string) string {
	switch severity {
	case "critical":
		return "🔴 CRITICAL"
	case "warning":
		return "🟡 WARNING"
	case "info":
		return "🔵 INFO"
	default:
		return severity
	}
}

// tableSeverityLabel returns a plain, fixed-width-friendly label for table output
func tableSeverityLabel(severity string) string {
	switch severity {
	case "critical":
		return "CRITICAL"
	case "warning":
		return "WARNING"
	case "info":
		return "INFO"
	default:
		return strings.ToUpper(severity)
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "⚪"
	}
}

func getTypeIcon(t string) string {
	switch t {
	case "added":
		return "➕"
	case "removed":
		return "➖"
	case "modified":
		return "✏️"
	default:
		return "•"
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}


func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if len(val) > 60 {
			return val[:57] + "..."
		}
		return val
	case models.FileInfo:
		return fmt.Sprintf("hash=%s, mode=%s", val.Hash[:8], val.Mode)
	case models.ServiceInfo:
		return fmt.Sprintf("status=%s, enabled=%v", val.Status, val.Enabled)
	default:
		return fmt.Sprintf("%v", v)
	}
}
