package web

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"sync"
)

// The app icon is drawn here instead of shipped as files: the "○" of a
// pending task, as a ring on the background color. Its inner part is
// inside the safe zone of maskable icons, so launchers can crop it.
var (
	iconBackground = color.RGBA{0x1e, 0x1e, 0x2e, 0xff}
	iconRing       = color.RGBA{0xcb, 0xa6, 0xf7, 0xff}

	iconCache sync.Map // size -> []byte
)

func (s *Server) icon(w http.ResponseWriter, r *http.Request) {
	size := 0
	switch r.URL.Path {
	case "/icon-192.png":
		size = 192
	case "/icon-512.png":
		size = 512
	default:
		http.NotFound(w, r)
		return
	}

	data, ok := iconCache.Load(size)
	if !ok {
		data = drawIcon(size)
		iconCache.Store(size, data)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data.([]byte))
}

func drawIcon(size int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	c := float64(size) / 2
	outer, inner := float64(size)*0.30, float64(size)*0.22

	for y := range size {
		for x := range size {
			d := math.Hypot(float64(x)+0.5-c, float64(y)+0.5-c)
			// one pixel of blend at each edge of the ring
			cover := math.Min(math.Max(outer-d+0.5, 0), 1) * math.Min(math.Max(d-inner+0.5, 0), 1)
			img.SetRGBA(x, y, blend(iconBackground, iconRing, cover))
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func blend(a, b color.RGBA, t float64) color.RGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 0xff}
}
