package speculumf01

// This file implements task76 Block 5's minimal state-corruption scenarios.
// Every function takes an intact State and returns a damaged State; none
// of them consult the message, matching the requirement that corruption
// be applied to S, not to M.

// SinglePositionDamage rewrites one ring's offset to a different,
// deterministically chosen wrong value (models both "damage of one
// position" and "replacement of one symbol" from the source list — the
// two are the same operation on this device, a finding worth stating
// explicitly rather than papering over).
func SinglePositionDamage(s State, ring int, newOffset int) State {
	out := s.Clone()
	out.Offsets[ring] = newOffset
	return out
}

// RandomRingShift adds delta (mod alphabet size) to one ring's offset,
// modeling an imprecise re-touch of a single ring during handling. On this
// device it produces the same observable effect as SinglePositionDamage;
// kept distinct for coverage of the source's scenario list.
func RandomRingShift(c Config, s State, ring, delta int) State {
	out := s.Clone()
	m := c.Alphabet.Size()
	out.Offsets[ring] = ((out.Offsets[ring]+delta)%m + m) % m
	return out
}

// DeleteRing marks one ring as physically unreadable (removed, shattered,
// illegible). Whether this is a local gap or a cascading desync depends
// entirely on Config.RingIdentityMarked, resolved at decode time by
// DecodeWithGap, not here — the corruption itself is identity-agnostic.
func DeleteRing(s State, ring int) State {
	out := s.Clone()
	out.Offsets[ring] = Missing
	return out
}

// SwapTwoRings exchanges two rings' offsets, producing a transposition
// error in the decoded message.
func SwapTwoRings(s State, ringA, ringB int) State {
	out := s.Clone()
	out.Offsets[ringA], out.Offsets[ringB] = out.Offsets[ringB], out.Offsets[ringA]
	return out
}

// LoseOrientationMark models destruction of the physical reading-radius
// mark itself, not any one ring. Because offset and radius enter the
// letter formula only as a sum (see LetterAtRadius), losing the mark by
// delta sectors is representationally identical to rotating every ring by
// delta — a genuinely global, single-cause misalignment rather than n
// independent local errors, and it is implemented that way deliberately.
func LoseOrientationMark(c Config, s State, delta int) State {
	out := s.Clone()
	m := c.Alphabet.Size()
	for i := range out.Offsets {
		if out.Offsets[i] == Missing {
			continue
		}
		out.Offsets[i] = ((out.Offsets[i]+delta)%m + m) % m
	}
	return out
}

// LoseOuterContour destroys the outermost `count` physical rings
// (edge/contour damage), independent of which end the message reading
// order starts from.
func LoseOuterContour(s State, count int) State {
	out := s.Clone()
	n := len(out.Offsets)
	for i := n - count; i < n; i++ {
		if i >= 0 {
			out.Offsets[i] = Missing
		}
	}
	return out
}

// DecodeWithGap reads a possibly-damaged state. physicalCollapse models
// whether missing rings leave the remaining rings' physical positions
// undisturbed (false: a hole exactly where the ring was, purely local) or
// causes the remaining rings to slide together on a shared axle so every
// later position shifts by one per missing ring (true: a synchronization
// cascade). '?' marks an unresolved character.
func (c Config) DecodeWithGap(state State, length int, physicalCollapse bool) string {
	out := make([]rune, length)
	if !physicalCollapse {
		for i := 0; i < length; i++ {
			pos := c.RingPos(i)
			if pos < 0 || pos >= len(state.Offsets) || state.Offsets[pos] == Missing {
				out[i] = '?'
				continue
			}
			out[i] = c.LetterAtMark(state.Offsets[pos])
		}
		return string(out)
	}

	// Compact the surviving rings in physical order, then read the first
	// `length` of them as if no ring had ever been removed.
	order := identityOrder(c.NumRings)
	if c.Order == OuterToInner {
		order = reversedOrder(c.NumRings)
	}
	var surviving []int
	for _, idx := range order {
		if state.Offsets[idx] != Missing {
			surviving = append(surviving, idx)
		}
	}
	for i := 0; i < length; i++ {
		if i >= len(surviving) {
			out[i] = '?'
			continue
		}
		out[i] = c.LetterAtMark(state.Offsets[surviving[i]])
	}
	return string(out)
}
