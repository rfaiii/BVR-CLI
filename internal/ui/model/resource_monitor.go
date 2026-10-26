package model

import (
	"fmt"
	"image/color"
	"runtime"
	"runtime/metrics"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/ui/styles"
)

const resourceMonitorInterval = 2 * time.Second

type resourceSnapshotMsg struct {
	time       time.Time
	cpuSeconds float64
	ramPercent float64
}

// resourceMonitorCmd publishes resource snapshots on a fixed interval.
// It uses runtime/metrics exclusively: runtime.ReadMemStats stops the world,
// which is visible as frame hitches when called once per second.
func (m *UI) resourceMonitorCmd() tea.Cmd {
	return tea.Tick(resourceMonitorInterval, func(now time.Time) tea.Msg {
		samples := []metrics.Sample{
			{Name: "/cpu/classes/total:cpu-seconds"},
			{Name: "/memory/classes/heap/objects:bytes"},
			{Name: "/memory/classes/total:bytes"},
		}
		metrics.Read(samples)

		ramPercent := 0.0
		heapAlloc, heapSys := samples[1].Value, samples[2].Value
		if heapAlloc.Kind() == metrics.KindUint64 && heapSys.Kind() == metrics.KindUint64 && heapSys.Uint64() > 0 {
			ramPercent = float64(heapAlloc.Uint64()) / float64(heapSys.Uint64()) * 100
		}

		return resourceSnapshotMsg{
			time:       now,
			cpuSeconds: samples[0].Value.Float64(),
			ramPercent: min(100, max(0, ramPercent)),
		}
	})
}

// resourceCPUPercent returns the CPU usage percent between two snapshots.
func resourceCPUPercent(previous, current resourceSnapshotMsg) float64 {
	elapsed := current.time.Sub(previous.time).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return min(100, max(0, (current.cpuSeconds-previous.cpuSeconds)/elapsed/float64(runtime.NumCPU())*100))
}

// drawResourceMonitor renders the CPU/RAM status line in the help region.
func (m *UI) drawResourceMonitor(scr uv.Screen, area uv.Rectangle) {
	if !m.resourceReady || area.Dx() < 10 || area.Dy() == 0 {
		return
	}

	margin := max(1, area.Dx()/10)
	width := max(0, area.Dx()-margin*2)
	cpuWidth := width * 40 / 100
	ramWidth := width * 40 / 100
	gapWidth := max(0, width-cpuWidth-ramWidth)
	cpu := renderResourceBar("CPU", m.resourceCPU, cpuWidth, m.com.Styles)
	ram := renderResourceBar("RAM", m.resourceRAM, ramWidth, m.com.Styles)
	gap := strings.Repeat(" ", gapWidth)
	line := strings.Repeat(" ", margin) + cpu + gap + ram
	uv.NewStyledString(lipgloss.NewStyle().Width(area.Dx()).Render(line)).Draw(scr, area)
}

// renderResourceBar builds one labeled usage bar with gradient fill.
func renderResourceBar(label string, percent float64, width int, styleSet *styles.Styles) string {
	if width <= 0 {
		return ""
	}
	value := fmt.Sprintf("%3.0f%%", percent)
	labelText := label + " "
	fixedWidth := lipgloss.Width(labelText) + lipgloss.Width(value) + 1
	barWidth := max(1, width-fixedWidth)
	filledWidth := min(barWidth, max(0, int(float64(barWidth)*percent/100)))
	emptyWidth := barWidth - filledWidth

	filled := ""
	if filledWidth > 0 {
		filled = renderGradientBar(filledWidth, percent, styleSet)
	}
	empty := styleSet.Status.ResourceEmpty.Render(strings.Repeat("░", emptyWidth))
	bar := filled + empty
	return styleSet.Status.ResourceLabel.Render(labelText) + bar + " " + styleSet.Status.ResourceValue.Render(value)
}

// gradientBarCache memoizes rendered bars; the same (width, percent, color)
// combination is rebuilt on every draw otherwise, at one lipgloss style per
// cell. Percent is quantized because the label renders whole percents.
var gradientBarCache struct {
	mu   sync.Mutex
	bars map[string]string
}

func cachedGradientBar(width int, percent float64, end color.Color) string {
	e := colorToRGB(end)
	endHex := fmt.Sprintf("#%02x%02x%02x", e.R, e.G, e.B)
	key := fmt.Sprintf("%d|%d|%s", width, int(percent+0.5), endHex)

	gradientBarCache.mu.Lock()
	defer gradientBarCache.mu.Unlock()
	if gradientBarCache.bars == nil {
		gradientBarCache.bars = make(map[string]string)
	}
	if bar, ok := gradientBarCache.bars[key]; ok {
		return bar
	}
	if len(gradientBarCache.bars) > 1024 {
		gradientBarCache.bars = make(map[string]string)
	}
	bar := buildGradientBar(width, float64(int(percent+0.5)), end)
	gradientBarCache.bars[key] = bar
	return bar
}

// renderGradientBar returns a horizontal bar that interpolates between a dark
// base and the theme primary color. The mix is driven by both position and
// usage percent, so light usage stays darker while high usage glows brighter.
func renderGradientBar(width int, percent float64, styleSet *styles.Styles) string {
	if width <= 0 {
		return ""
	}
	end := styleSet.Status.ResourceFilled.GetForeground()
	return cachedGradientBar(width, percent, end)
}

func buildGradientBar(width int, percent float64, end color.Color) string {
	start := colorColor("#2a2a2a")
	startRGB := colorToRGB(start)
	endRGB := colorToRGB(end)
	parts := make([]string, width)
	for i := 0; i < width; i++ {
		t := 0.0
		if width > 1 {
			t = float64(i) / float64(width-1)
		}
		t = t*0.6 + 0.4*percent/100
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		c := lerpColor(startRGB, endRGB, t)
		parts[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B))).Render("█")
	}
	return strings.Join(parts, "")
}

// colorColor converts a hex string into the color type used by the style API.
func colorColor(s string) color.Color {
	return lipgloss.Color(s)
}

// rgb is a simple sRGB triple used for gradient interpolation.
type rgb struct {
	R, G, B uint8
}

// colorToRGB converts a color.Color to an 8-bit sRGB triple.
func colorToRGB(c color.Color) rgb {
	rr, gg, bb, _ := c.RGBA()
	return rgb{
		R: uint8(rr >> 8),
		G: uint8(gg >> 8),
		B: uint8(bb >> 8),
	}
}

// lerpColor blends two sRGB colors by a normalized factor.
func lerpColor(a, b rgb, t float64) rgb {
	return rgb{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
	}
}
