package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

func squareImage(img image.Image, maxSize int) image.Image {
	bounds := largestCenterSquare(img.Bounds())
	if bounds.Dx() <= maxSize {
		// Just use subimage if we are already below max size.
		if img, ok := img.(subimager); ok {
			return img.SubImage(bounds)
		}
		dest := image.NewNRGBA(bounds.Sub(bounds.Min))
		draw.Draw(dest, dest.Bounds(), img, bounds.Min, draw.Src)
		return dest
	}
	// Resize it to a smaller image.
	destRect := image.Rect(0, 0, maxSize, maxSize)
	dest := image.NewNRGBA(destRect)
	draw.CatmullRom.Scale(dest, destRect, img, bounds, draw.Src, nil)
	return dest
}

func largestCenterSquare(rect image.Rectangle) image.Rectangle {
	w := rect.Dx()
	h := rect.Dy()
	switch {
	case w == h:
		return rect
	case w > h:
		marginX := (w - h) / 2
		minX := rect.Min.X + marginX
		return image.Rectangle{
			Min: image.Pt(minX, rect.Min.Y),
			Max: image.Pt(minX+h, rect.Max.Y),
		}
	default: // w < h.
		marginY := (h - w) / 2
		minY := rect.Min.Y + marginY
		return image.Rectangle{
			Min: image.Pt(rect.Min.X, minY),
			Max: image.Pt(rect.Max.X, minY+w),
		}
	}
}

type subimager interface {
	SubImage(r image.Rectangle) image.Image
}
