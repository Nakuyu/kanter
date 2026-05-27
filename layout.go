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

// SpaceAllocation distributes total pixels/rows across the items so like viewports and other lines
// MinSize reservation and a weight is given to each item (it's not utilised well ik but just did it anyways)
// so this weight is used for surplus space, Every item is allocated so nothing is lost to rounding,
// three ways this should work
// surplus so extra i mean (total > mins) so each gets their required minsize and then the remainder is split propor-
// -tionally by weight is handed to items one by one
// exact is (total == mins required by all items) so yea easy split
// overflow (total < mins) so here items shrink proportionally by minsize ratio so total never exceeds the available
// space, so without this columns would overflow the terminal width,
// Also i m sure theres better ways to do this, prolly even some library, but it works good so it's fine
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

	remaining := total - totalMin
	if remaining >= 0 && weightSum > 0 {
		// Enough space: distribute surplus proportionally by weight
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
	} else if remaining < 0 && totalMin > 0 {
		// Not enough space: scale items proportionally by their MinSize
		var allocated int
		for i, it := range items {
			result[i] = total * it.MinSize / totalMin
			allocated += result[i]
		}
		// Distribute any remainder from rounding errors to reach exactly 'total'
		for i := 0; allocated < total; i++ {
			result[i%len(items)]++
			allocated++
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
	// 2 lines reserved for sticky column header and bottom border,
	// both rendered outside the viewport so the column stays framed
	vpHeight := max(height-2-footerHeight-detailHeight, 3)

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
	case ZoneWide, ZoneNormal:
		return DetailFull, min(bodyLineCount+4, wideDetailMax)
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
