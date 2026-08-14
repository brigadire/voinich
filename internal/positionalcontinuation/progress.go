package positionalcontinuation

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// progressReporter is the same stage/elapsed/ETA status bar used by
// structural-projection-analyze and every later confirmatory stage.
type progressReporter struct {
	w             io.Writer
	interactive   bool
	started, last time.Time
	stage, stages int
	clock         func() time.Time
}

func newProgress(w io.Writer) *progressReporter {
	p := &progressReporter{w: w, stages: 15, clock: time.Now}
	if f, ok := w.(*os.File); ok {
		if info, e := f.Stat(); e == nil {
			p.interactive = info.Mode()&os.ModeCharDevice != 0
		}
	}
	return p
}

func (p *progressReporter) begin(stage int, label string) {
	if p == nil || p.w == nil {
		return
	}
	p.stage, p.started, p.last = stage, p.clock(), time.Time{}
	p.line(fmt.Sprintf("[%d/%d] %s", stage, p.stages, label), true)
}

func (p *progressReporter) update(done, total int, label string) {
	if p == nil || p.w == nil || total <= 0 {
		return
	}
	now := p.clock()
	final := done >= total
	if !final && !p.last.IsZero() && now.Sub(p.last) < time.Second {
		return
	}
	p.last = now
	elapsed := now.Sub(p.started)
	eta := time.Duration(0)
	if done > 0 && done < total {
		eta = time.Duration(float64(elapsed) * float64(total-done) / float64(done))
	}
	width := 20
	filled := done * width / total
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(".", width-filled)
	msg := fmt.Sprintf("[%d/%d] %s [%s] %d/%d (%d%%) | elapsed %s", p.stage, p.stages, label, bar, done, total, done*100/total, shortDuration(elapsed))
	if done > 0 && done < total {
		msg += " | ETA " + shortDuration(eta)
	}
	p.line(msg, final)
}

func (p *progressReporter) line(message string, final bool) {
	if p.interactive {
		end := "\r"
		if final {
			end = "\n"
		}
		fmt.Fprint(p.w, "\r\x1b[2K", message, end)
	} else {
		fmt.Fprintln(p.w, message)
	}
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	h := int(d / time.Hour)
	m := int(d/time.Minute) % 60
	s := int(d/time.Second) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
