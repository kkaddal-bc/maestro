package listtable

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const defaultDescriptionLimit = 60

type Row struct {
	Name        string
	Description string
	Statuses    []string
}

type Renderer struct {
	DescriptionLimit int
}

func NewRenderer() Renderer {
	return Renderer{DescriptionLimit: defaultDescriptionLimit}
}

func (r Renderer) Render(w io.Writer, headers []string, rows []Row) error {
	if len(headers) == 0 {
		return nil
	}

	normalizedRows := make([]Row, 0, len(rows))
	for _, row := range rows {
		normalizedRows = append(normalizedRows, Row{
			Name:        row.Name,
			Description: r.truncateDescription(row.Description),
			Statuses:    append([]string(nil), row.Statuses...),
		})
	}

	widths := columnWidths(headers, normalizedRows)
	useStyles := isTerminalWriter(w)
	headerStyle := lipgloss.NewStyle().Bold(true)
	descriptionStyle := lipgloss.NewStyle().Faint(true)
	installedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	missingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Fprintln(w, renderRow(headers, widths, headerStyle, nil, useStyles))
	for _, row := range normalizedRows {
		cells := make([]string, 0, 2+len(row.Statuses))
		cells = append(cells, row.Name, row.Description)
		cells = append(cells, row.Statuses...)

		styles := make([]lipgloss.Style, 0, len(cells))
		styles = append(styles, lipgloss.NewStyle(), descriptionStyle)
		for _, status := range row.Statuses {
			if status == "installed" {
				styles = append(styles, installedStyle)
				continue
			}
			styles = append(styles, missingStyle)
		}

		fmt.Fprintln(w, renderRow(cells, widths, lipgloss.Style{}, styles, useStyles))
	}

	return nil
}

func (r Renderer) truncateDescription(description string) string {
	limit := r.DescriptionLimit
	if limit <= 0 {
		limit = defaultDescriptionLimit
	}

	runes := []rune(description)
	if len(runes) <= limit {
		return description
	}
	if limit <= 1 {
		return "…"
	}

	return string(runes[:limit-1]) + "…"
}

func columnWidths(headers []string, rows []Row) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = displayWidth(header)
	}

	for _, row := range rows {
		values := append([]string{row.Name, row.Description}, row.Statuses...)
		for i, value := range values {
			if i >= len(widths) {
				break
			}
			if width := displayWidth(value); width > widths[i] {
				widths[i] = width
			}
		}
	}

	return widths
}

func renderRow(values []string, widths []int, rowStyle lipgloss.Style, cellStyles []lipgloss.Style, useStyles bool) string {
	cells := make([]string, len(values))
	for i, value := range values {
		cell := padRight(value, widths[i])
		if useStyles && i < len(cellStyles) {
			cell = cellStyles[i].Render(cell)
		}
		cells[i] = cell
	}

	if useStyles {
		return rowStyle.Render(strings.Join(cells, "  "))
	}
	return strings.Join(cells, "  ")
}

func padRight(value string, width int) string {
	current := displayWidth(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func displayWidth(value string) int {
	return lipgloss.Width(value)
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
