package speculumf01

import (
	"fmt"
	"math"
	"strings"
)

// RenderASCII renders exactly what the device's external state makes
// observable: for every ring (outer to inner, matching the physical
// stacking order a human sees on approach) the full sector sequence
// starting at the marked reading radius, with that radius's letter
// bracketed. No ring identity numbers, hidden alignment, or metadata not
// physically present on the object are added.
func (c Config) RenderASCII(s State) string {
	var b strings.Builder
	m := c.Alphabet.Size()
	fmt.Fprintf(&b, "F01 Speculum -- %d rings, alphabet=%s (%d letters), read mark at sector %d\n", c.NumRings, c.Alphabet.Name, m, c.ReadRadius)
	fmt.Fprintf(&b, "(rendered outer ring first, as seen looking down on the stack)\n\n")
	for ring := c.NumRings - 1; ring >= 0; ring-- {
		off := s.Offsets[ring]
		fmt.Fprintf(&b, "ring %2d: ", ring)
		if off == Missing {
			b.WriteString("[DESTROYED/UNREADABLE]\n")
			continue
		}
		for sector := 0; sector < m; sector++ {
			letter := c.LetterAtRadius(off, sector)
			if sector == c.ReadRadius {
				fmt.Fprintf(&b, "[%c]", letter)
			} else {
				fmt.Fprintf(&b, " %c ", letter)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// RenderSVG draws a printable plan view: one concentric circle per ring,
// letters placed in their physical sectors at the ring's current offset,
// and a single radial line marking the reading radius. This is the
// task76 Block 2 "printable representation" artifact.
func (c Config) RenderSVG(s State) string {
	m := c.Alphabet.Size()
	const ringGap = 26.0
	const centerR = 30.0
	size := centerR + float64(c.NumRings)*ringGap + 40
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" font-family="monospace" font-size="10">`, size*2, size*2)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="white"/>`)
	cx, cy := size, size

	// radius mark line, drawn first so letters sit on top of it
	markAngle := 2 * math.Pi * float64(c.ReadRadius) / float64(m)
	mx := cx + (centerR+float64(c.NumRings)*ringGap+10)*math.Cos(markAngle)
	my := cy + (centerR+float64(c.NumRings)*ringGap+10)*math.Sin(markAngle)
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="red" stroke-width="1.5"/>`, cx, cy, mx, my)

	for ring := 0; ring < c.NumRings; ring++ {
		r := centerR + float64(ring)*ringGap
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="black" stroke-width="0.5"/>`, cx, cy, r)
		off := s.Offsets[ring]
		if off == Missing {
			continue
		}
		for sector := 0; sector < m; sector++ {
			letter := c.LetterAtRadius(off, sector)
			angle := 2 * math.Pi * float64(sector) / float64(m)
			x := cx + r*math.Cos(angle)
			y := cy + r*math.Sin(angle)
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle">%c</text>`, x, y, letter)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}
