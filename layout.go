package main

type LayoutZone int

const (
	ZoneWide LayoutZone = iota
	ZoneNormal
	ZoneCompact
	ZoneNarrow
)

type DetailMode int

const (
	DetailFull DetailMode = iota
	DetailCompact
	DetailHidden
)

type SpaceItem struct {
	MinSize int
	Weight  int
}

// SpaceAllocation calculation part, Integer-only no floating point.
func SpaceAllocation(total int, items []SpaceItem) []int {
	if len(items) == 0 {
		return nil
	}

	result := make([]int, len(items))
	var totalMin, weightSum int
	for i, it := range items {
		result[i] = it.MinSize
		totalMin += it.MinSize
		weightSum += it.Weight
	}

	remaining := total - totalMin // dividing the remainder proportionally by Weight. Remainder distributed 1 by 1 in order.
	if remaining > 0 && weightSum > 0 {
		unit := remaining / weightSum
		extra := remaining - unit*weightSum

		for i, it := range items {
			if it.Weight > 0 {
				result[i] += unit * it.Weight
				if extra > 0 {
					result[i]++
					extra--
				}
			}
		}
	} else if remaining < 0 {
		// Not enough space for minimums distribute proportionally by MinSize.
		if totalMin > 0 {
			var allocated int
			for i, it := range items {
				result[i] = total * it.MinSize / totalMin
				allocated += result[i]
			}
			for i := 0; allocated < total; i++ {
				result[i%len(items)]++
				allocated++
			}
		}
	}

	return result
}

type ScreenLayout struct {
	Width  int
	Height int

	Zone       LayoutZone
	DetailMode DetailMode

	ColWidths      [3]int
	ViewportHeight int
	DetailHeight   int
	FooterHeight   int
}

func computeLayout(width, height int, bodyLineCount int, hideDetail bool) ScreenLayout {
	zone := classifyZone(width)
	detailMode, detailHeight := detailInfo(zone, bodyLineCount, hideDetail)

	footerHeight := 4
	vpHeight := max(height-footerHeight-detailHeight, 5)

	colWidths := columnWidths(zone, width)

	return ScreenLayout{
		Width:  width,
		Height: height,

		Zone:       zone,
		DetailMode: detailMode,

		ColWidths:      colWidths,
		ViewportHeight: vpHeight,
		DetailHeight:   detailHeight,
		FooterHeight:   footerHeight,
	}
}

func classifyZone(width int) LayoutZone {
	switch {
	case width >= 130:
		return ZoneWide
	case width >= 85:
		return ZoneNormal
	case width >= 60:
		return ZoneCompact
	default:
		return ZoneNarrow
	}
}

const (
	minColWidth   = 20
	wideDetailMax = 8
)

func detailInfo(zone LayoutZone, bodyLineCount int, hideDetail bool) (DetailMode, int) {
	if hideDetail {
		return DetailHidden, 0
	}

	switch zone {
	case ZoneWide:
		return DetailFull, min(bodyLineCount+4, wideDetailMax)
	case ZoneNormal:
		return DetailCompact, 3
	default:
		return DetailHidden, 0
	}
}

func columnWidths(zone LayoutZone, totalWidth int) [3]int {
	var items []SpaceItem
	var colCount int

	switch zone {
	case ZoneNarrow:
		items = []SpaceItem{
			{MinSize: minColWidth, Weight: 1},
			{MinSize: minColWidth, Weight: 1},
		}
		colCount = 2
	default:
		items = []SpaceItem{
			{MinSize: minColWidth, Weight: 1},
			{MinSize: minColWidth, Weight: 1},
			{MinSize: minColWidth, Weight: 1},
		}
		colCount = 3
	}

	sizes := SpaceAllocation(totalWidth, items)

	var widths [3]int
	for i := 0; i < colCount && i < len(sizes); i++ {
		widths[i] = sizes[i]
	}
	return widths
}
