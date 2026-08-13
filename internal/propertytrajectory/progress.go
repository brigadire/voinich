package propertytrajectory

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type progressReporter struct {
	w             io.Writer
	interactive   bool
	started, last time.Time
	stage, stages int
}

func newProgress(w io.Writer) *progressReporter {
	p := &progressReporter{w: w, stages: 7}
	if f, ok := w.(*os.File); ok {
		if i, e := f.Stat(); e == nil {
			p.interactive = i.Mode()&os.ModeCharDevice != 0
		}
	}
	return p
}
func (p *progressReporter) begin(stage int, label string) {
	if p == nil || p.w == nil {
		return
	}
	p.stage = stage
	p.started = time.Now()
	p.last = time.Time{}
	p.line(fmt.Sprintf("[%d/%d] %s", stage, p.stages, label), true)
}
func (p *progressReporter) update(done, total int, label string) {
	if p == nil || p.w == nil || total <= 0 {
		return
	}
	now := time.Now()
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
	filled := done * 20 / total
	if filled > 20 {
		filled = 20
	}
	msg := fmt.Sprintf("[%d/%d] %s [%s%s] %d/%d (%d%%) | elapsed %s", p.stage, p.stages, label, strings.Repeat("=", filled), strings.Repeat(".", 20-filled), done, total, done*100/total, shortDuration(elapsed))
	if done > 0 && done < total {
		msg += " | ETA " + shortDuration(eta)
	}
	p.line(msg, final)
}
func (p *progressReporter) line(s string, final bool) {
	if p.interactive {
		end := "\r"
		if final {
			end = "\n"
		}
		fmt.Fprint(p.w, "\r\x1b[2K", s, end)
	} else {
		fmt.Fprintln(p.w, s)
	}
}
func shortDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < 0 {
		d = 0
	}
	h := int(d / time.Hour)
	m := int(d/time.Minute) % 60
	s := int(d/time.Second) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
