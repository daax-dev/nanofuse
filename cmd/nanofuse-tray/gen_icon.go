//go:build ignore

// gen_icon.go generates the Nanofuse tray icons from code (no external deps).
//
//	go run gen_icon.go
//
// Produces, under assets/:
//   - icon.ico       multi-size Windows icon (blue rounded badge + white hexagon)
//   - icon_mac.png   macOS template glyph (black hexagon on transparent; OS tints it)
//   - icon.png       64px PNG for docs / other platforms
//
// The hexagon evokes a microVM "cell"; the rounded blue badge reads cleanly in
// the Windows notification area at 16px.
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

func main() {
	outDir := "assets"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	sizes := []int{16, 32, 48, 256}
	frames := make([][]byte, 0, len(sizes))
	for _, s := range sizes {
		frames = append(frames, encodePNG(renderBadge(s)))
	}
	write(filepath.Join(outDir, "icon.ico"), encodeICO(sizes, frames))
	write(filepath.Join(outDir, "icon_mac.png"), encodePNG(renderTemplate(44)))
	write(filepath.Join(outDir, "icon.png"), encodePNG(renderBadge(64)))
	log.Println("wrote tray icons to", outDir)
}

func write(path string, data []byte) {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Fatal(err)
	}
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

// renderBadge draws a blue rounded square with a centered white hexagon.
func renderBadge(size int) *image.NRGBA {
	const ss = 4
	w := size * ss
	hi := image.NewNRGBA(image.Rect(0, 0, w, w))
	blue := color.NRGBA{R: 45, G: 108, B: 223, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	cx, cy := float64(w)/2, float64(w)/2
	margin := float64(w) * 0.06
	radius := float64(w) * 0.22
	hexR := float64(w) * 0.30
	verts := hexVerts(cx, cy, hexR)
	for y := 0; y < w; y++ {
		for x := 0; x < w; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			switch {
			case pointInHex(fx, fy, verts):
				hi.SetNRGBA(x, y, white)
			case inRoundedRect(fx, fy, margin, margin, float64(w)-margin, float64(w)-margin, radius):
				hi.SetNRGBA(x, y, blue)
			}
		}
	}
	return downscale(hi, size, ss)
}

// renderTemplate draws a solid black hexagon on transparent for macOS template use.
func renderTemplate(size int) *image.NRGBA {
	const ss = 4
	w := size * ss
	hi := image.NewNRGBA(image.Rect(0, 0, w, w))
	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	cx, cy := float64(w)/2, float64(w)/2
	verts := hexVerts(cx, cy, float64(w)*0.42)
	for y := 0; y < w; y++ {
		for x := 0; x < w; x++ {
			if pointInHex(float64(x)+0.5, float64(y)+0.5, verts) {
				hi.SetNRGBA(x, y, black)
			}
		}
	}
	return downscale(hi, size, ss)
}

// hexVerts returns the 6 vertices of a flat-top regular hexagon.
func hexVerts(cx, cy, r float64) [6][2]float64 {
	var v [6][2]float64
	for k := 0; k < 6; k++ {
		a := float64(k) * math.Pi / 3
		v[k] = [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)}
	}
	return v
}

func pointInHex(px, py float64, v [6][2]float64) bool {
	var pos, neg bool
	for i := 0; i < 6; i++ {
		x1, y1 := v[i][0], v[i][1]
		x2, y2 := v[(i+1)%6][0], v[(i+1)%6][1]
		cross := (x2-x1)*(py-y1) - (y2-y1)*(px-x1)
		if cross > 0 {
			pos = true
		}
		if cross < 0 {
			neg = true
		}
	}
	return !(pos && neg)
}

func inRoundedRect(px, py, x0, y0, x1, y1, r float64) bool {
	if px < x0 || px > x1 || py < y0 || py > y1 {
		return false
	}
	switch {
	case px < x0+r && py < y0+r:
		return hypot(px-(x0+r), py-(y0+r)) <= r
	case px > x1-r && py < y0+r:
		return hypot(px-(x1-r), py-(y0+r)) <= r
	case px < x0+r && py > y1-r:
		return hypot(px-(x0+r), py-(y1-r)) <= r
	case px > x1-r && py > y1-r:
		return hypot(px-(x1-r), py-(y1-r)) <= r
	}
	return true
}

func hypot(a, b float64) float64 { return math.Sqrt(a*a + b*b) }

// downscale box-filters the supersampled image, averaging in premultiplied
// alpha for clean anti-aliased edges.
func downscale(hi *image.NRGBA, size, ss int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	n := float64(ss * ss)
	for oy := 0; oy < size; oy++ {
		for ox := 0; ox < size; ox++ {
			var r, g, b, a float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					c := hi.NRGBAAt(ox*ss+sx, oy*ss+sy)
					af := float64(c.A) / 255
					r += float64(c.R) * af
					g += float64(c.G) * af
					b += float64(c.B) * af
					a += af
				}
			}
			if a <= 0 {
				out.SetNRGBA(ox, oy, color.NRGBA{})
				continue
			}
			out.SetNRGBA(ox, oy, color.NRGBA{
				R: clamp8(r / a),
				G: clamp8(g / a),
				B: clamp8(b / a),
				A: clamp8(a / n * 255),
			})
		}
	}
	return out
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

// encodeICO packs PNG frames into a Windows .ico container.
func encodeICO(sizes []int, frames [][]byte) []byte {
	var buf bytes.Buffer
	le := binary.LittleEndian
	_ = binary.Write(&buf, le, uint16(0)) // reserved
	_ = binary.Write(&buf, le, uint16(1)) // type: icon
	_ = binary.Write(&buf, le, uint16(len(frames)))
	offset := 6 + 16*len(frames)
	for i, s := range sizes {
		dim := byte(s)
		if s >= 256 {
			dim = 0 // 0 means 256 in the ICO format
		}
		buf.WriteByte(dim)                                 // width
		buf.WriteByte(dim)                                 // height
		buf.WriteByte(0)                                   // palette size
		buf.WriteByte(0)                                   // reserved
		_ = binary.Write(&buf, le, uint16(1))              // color planes
		_ = binary.Write(&buf, le, uint16(32))             // bits per pixel
		_ = binary.Write(&buf, le, uint32(len(frames[i]))) // image size
		_ = binary.Write(&buf, le, uint32(offset))         // image offset
		offset += len(frames[i])
	}
	for _, f := range frames {
		buf.Write(f)
	}
	return buf.Bytes()
}
