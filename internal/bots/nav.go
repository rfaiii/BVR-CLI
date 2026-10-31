package bots

// Direction is a picker move.
type Direction int

const (
	DirLeft Direction = iota
	DirRight
	DirUp
	DirDown
)

// Columns reports how many sprites fit on a row. The boolean is list
// mode: the terminal is too narrow for even one sprite card.
func Columns(width, count int) (cols int, list bool) {
	if count < 1 {
		return 1, false
	}
	if width < SpriteW {
		return 1, true
	}
	fit := (width + CardGap) / (SpriteW + CardGap)
	if fit < 1 {
		return 1, true
	}
	if fit > count {
		fit = count
	}
	return fit, false
}

// Move wraps selection through the grid. Left and right walk the roster.
// Up and down walk the column, clamping a short last row.
func Move(selected, count, cols int, dir Direction) int {
	if count <= 0 {
		return 0
	}
	if selected < 0 || selected >= count {
		selected = 0
	}
	if cols < 1 {
		cols = 1
	}
	switch dir {
	case DirLeft:
		return (selected - 1 + count) % count
	case DirRight:
		return (selected + 1) % count
	case DirUp, DirDown:
		col := selected % cols
		row := selected / cols
		rows := (count + cols - 1) / cols
		if dir == DirUp {
			row--
			if row < 0 {
				row = rows - 1
			}
		} else {
			row++
			if row >= rows {
				row = 0
			}
		}
		idx := row*cols + col
		if idx >= count {
			idx = count - 1
		}
		return idx
	default:
		return selected
	}
}
