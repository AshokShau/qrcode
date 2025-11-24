package qrcode

import (
	"image"
	"image/color"
	"image/png"
	"io"
)

// WritePNG writes the QR code to the given writer as a PNG.
// scale is the number of pixels per module.
func (qr *QRCode) WritePNG(w io.Writer, scale int) error {
	if scale < 1 {
		scale = 1
	}

	border := 4 // Quiet zone
	dim := (qr.Size + 2*border) * scale

	img := image.NewPaletted(image.Rect(0, 0, dim, dim), color.Palette{
		color.White,
		color.Black,
	})

	// Default to white
	for i := 0; i < len(img.Pix); i++ {
		img.Pix[i] = 0 // Index 0 is White
	}

	for r := 0; r < qr.Size; r++ {
		for c := 0; c < qr.Size; c++ {
			if qr.Modules[r][c] {
				// Draw scaled module
				startX := (c + border) * scale
				startY := (r + border) * scale

				for y := 0; y < scale; y++ {
					for x := 0; x < scale; x++ {
						img.SetColorIndex(startX+x, startY+y, 1) // Index 1 is Black
					}
				}
			}
		}
	}

	return png.Encode(w, img)
}
