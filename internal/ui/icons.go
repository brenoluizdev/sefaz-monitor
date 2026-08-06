package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/lxn/walk"

	"sefazmonitor/internal/monitor"
)

// statusRGB devolve a cor (R, G, B) associada a cada nível de status, usada
// tanto no ícone da bandeja quanto na tabela de UFs.
func statusRGB(s monitor.Status) (byte, byte, byte) {
	switch s {
	case monitor.OK:
		return 0x2E, 0xB1, 0x4C // verde
	case monitor.Degraded:
		return 0xE8, 0xA8, 0x1B // amarelo
	case monitor.Down:
		return 0xD6, 0x3B, 0x3B // vermelho
	default:
		return 0x9A, 0x9A, 0x9A // cinza (desconhecido)
	}
}

func statusImageColor(s monitor.Status) color.RGBA {
	r, g, b := statusRGB(s)
	return color.RGBA{r, g, b, 0xFF}
}

func statusWalkColor(s monitor.Status) walk.Color {
	r, g, b := statusRGB(s)
	return walk.RGB(r, g, b)
}

// dotImage desenha um círculo preenchido de tamanho size x size na cor c,
// usado como base tanto do ícone estático quanto dos ícones de status da
// bandeja.
func dotImage(size int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 1

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if math.Hypot(dx, dy) <= r {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

// statusIcon gera (em memória, sem depender de arquivos .ico externos) o
// ícone de bandeja correspondente a um status.
func statusIcon(s monitor.Status) (*walk.Icon, error) {
	return walk.NewIconFromImageForDPI(dotImage(32, statusImageColor(s)), 96)
}
