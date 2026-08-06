// Programa auxiliar (não faz parte do app): gera resources/app.ico, o ícone
// estático do executável e do instalador. Usa o formato ICO "PNG embutido"
// (suportado desde o Windows Vista), o que evita depender de um encoder BMP
// manual — o pacote image/png da stdlib já basta.
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
)

// satelliteDish desenha um ícone simples: um círculo azul (fundo) com um
// "sinal" em arcos concêntricos claros, remetendo a monitoramento/rádio.
func satelliteDish(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 1

	bg := color.RGBA{0x1E, 0x3A, 0x5F, 0xFF}
	fg := color.RGBA{0xF2, 0xC9, 0x2E, 0xFF}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			d := math.Hypot(dx, dy)
			if d > r {
				continue
			}
			img.Set(x, y, bg)
		}
	}

	// três arcos concêntricos no canto inferior esquerdo, como um ícone de
	// sinal/wifi, para remeter a "status de serviço".
	originX, originY := cx-r*0.35, cy+r*0.35
	for _, ring := range []float64{0.35, 0.55, 0.75} {
		ringR := r * ring
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx, dy := float64(x)+0.5-originX, float64(y)+0.5-originY
				d := math.Hypot(dx, dy)
				// só desenha o quadrante superior-direito do arco
				if dx < 0 || dy > 0 {
					continue
				}
				if d > ringR-1.2 && d < ringR+1.2 {
					cdx, cdy := float64(x)+0.5-cx, float64(y)+0.5-cy
					if math.Hypot(cdx, cdy) <= r {
						img.Set(x, y, fg)
					}
				}
			}
		}
	}
	// ponto central do "sinal"
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-originX, float64(y)+0.5-originY
			if math.Hypot(dx, dy) <= float64(size)*0.045 {
				img.Set(x, y, fg)
			}
		}
	}

	return img
}

type icoDirEntry struct {
	Width, Height      uint8
	ColorCount, Reserved uint8
	Planes, BitCount    uint16
	BytesInRes          uint32
	ImageOffset         uint32
}

func writeICO(path string, sizes []int) error {
	type entry struct {
		size int
		png  []byte
	}
	entries := make([]entry, 0, len(sizes))
	for _, s := range sizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, satelliteDish(s)); err != nil {
			return err
		}
		entries = append(entries, entry{size: s, png: buf.Bytes()})
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// ICONDIR
	binary.Write(f, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(f, binary.LittleEndian, uint16(1)) // type = icon
	binary.Write(f, binary.LittleEndian, uint16(len(entries)))

	headerSize := 6 + 16*len(entries)
	offset := uint32(headerSize)
	for _, e := range entries {
		wh := uint8(e.size)
		if e.size >= 256 {
			wh = 0
		}
		binary.Write(f, binary.LittleEndian, wh)                 // width
		binary.Write(f, binary.LittleEndian, wh)                 // height
		binary.Write(f, binary.LittleEndian, uint8(0))            // color count
		binary.Write(f, binary.LittleEndian, uint8(0))            // reserved
		binary.Write(f, binary.LittleEndian, uint16(1))           // planes
		binary.Write(f, binary.LittleEndian, uint16(32))          // bit count
		binary.Write(f, binary.LittleEndian, uint32(len(e.png)))  // bytes in res
		binary.Write(f, binary.LittleEndian, offset)              // image offset
		offset += uint32(len(e.png))
	}
	for _, e := range entries {
		f.Write(e.png)
	}
	return nil
}

func main() {
	if err := os.MkdirAll("resources", 0o755); err != nil {
		log.Fatal(err)
	}
	if err := writeICO("resources/app.ico", []int{16, 32, 48, 256}); err != nil {
		log.Fatal(err)
	}
	log.Println("resources/app.ico gerado")
}
