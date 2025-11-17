package browser

import (
	"strings"
)

func renderParentConnector(height, leftSel, rightSel int) string {
	if height <= 0 {
		return ""
	}

	leftSel = clampRow(leftSel, height)
	rightSel = clampRow(rightSel, height)

	lines := make([]string, height)

	if leftSel == rightSel {
		lines[leftSel] = border.Top
		return pStyle.Render(strings.Join(lines, "\n"))
	}

	start := min(leftSel, rightSel)
	end := max(leftSel, rightSel)

	var startGlyph, endGlyph string
	if leftSel < rightSel {
		startGlyph = border.TopRight
		endGlyph = border.BottomLeft
	} else {
		startGlyph = border.TopLeft
		endGlyph = border.BottomRight
	}

	lines[start] = startGlyph
	for row := start + 1; row < end; row++ {
		lines[row] = border.Right
	}
	lines[end] = endGlyph

	return pStyle.Render(strings.Join(lines, "\n"))
}

func renderChildConnector(height, selRow int) string {
	if height <= 0 {
		return ""
	}

	selRow = clampRow(selRow, height)
	lines := make([]string, height)

	if selRow == 0 {
		lines[0] = border.MiddleTop
	} else {
		lines[0] = border.TopLeft
		for row := 1; row < selRow; row++ {
			lines[row] = border.Left
		}
	}

	switch {
	case selRow == height-1:
		lines[selRow] = border.MiddleBottom
	case selRow == 0:
		if height > 1 {
			lines[1] = border.Left
		}
	default:
		lines[selRow] = border.MiddleRight
	}

	for row := selRow + 1; row < height-1; row++ {
		if lines[row] == "" {
			lines[row] = border.Left
		}
	}

	if lines[height-1] == "" {
		lines[height-1] = border.BottomLeft
	}

	return pStyle.Render(strings.Join(lines, "\n"))
}

func clampRow(row, height int) int {
	switch {
	case row < 0:
		return 0
	case row >= height:
		return height - 1
	default:
		return row
	}
}
