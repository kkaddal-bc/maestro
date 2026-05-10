package listtable

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/style"
)

const defaultDescriptionLimit = 60

type Row struct {
	Name        string
	Description string
	Status      string
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

	normalizedRows := r.normalizeRows(rows)
	widths := columnWidths(headers, normalizedRows)
	useStyles := isTerminalWriter(w)
	headerStyle := lipgloss.NewStyle().Bold(true)
	installedStyle := style.Accent

	fmt.Fprintln(w, renderRow(headers, widths, headerStyle, nil, useStyles))
	for _, row := range normalizedRows {
		cells := row.cells()
		styles := row.styles(installedStyle)
		fmt.Fprintln(w, renderRow(cells, widths, lipgloss.Style{}, styles, useStyles))
	}

	return nil
}

func (r Renderer) normalizeRows(rows []Row) []Row {
	normalized := make([]Row, len(rows))
	for i, row := range rows {
		normalized[i] = Row{
			Name:        row.Name,
			Description: r.truncateDescription(row.Description),
			Status:      row.Status,
		}
	}

	return normalized
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
		values := row.cells()
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

func (row Row) cells() []string {
	return []string{row.Name, row.Description, row.Status}
}

func (row Row) styles(installedStyle lipgloss.Style) []lipgloss.Style {
	return []lipgloss.Style{
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		statusStyle(row.Status, installedStyle),
	}
}

func statusStyle(status string, installedStyle lipgloss.Style) lipgloss.Style {
	if status == "installed" {
		return installedStyle
	}
	return lipgloss.NewStyle()
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
