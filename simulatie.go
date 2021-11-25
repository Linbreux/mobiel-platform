package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

var (
	l1        = 765.0
	a1        = 0.0 / 180 * math.Pi
	l2        = 1110.0
	a2        = a1 + 45.0/180*math.Pi
	L         = 1000.0
	scale     = 1 / 8.0
	rechtssom = true
	vWiel1    = 25.0
	vWiel2    = 50.0
	alpha     = 0.0
	r         = 0.0
	omega     = 0.00
	vm        = 50.0
)

func run() {
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Bounds:      pixel.R(0, 0, 1500, 750),
		VSync:       true,
		Undecorated: false,
	})
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)
	matrix := pixel.IM.ScaledXY(pixel.ZV, pixel.V(1, 1)).Moved(pixel.V(750, 50))
	win.SetMatrix(matrix)
	rotc := pixel.Vec{}

	for !win.Closed() {
		win.SetClosed(win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ))
		win.Clear(color.NRGBA{44, 44, 84, 255})
		lengteVoertuig1, lengteVoertuig2 := update()

		// Laat deel 2 van het voertuig draaien, om de beurt links en rechts.
		// 0.785 = 60°

		if rechtssom {
			omega += 0.0001
			if a2 >= 0.785 {
				rechtssom = false
			}
		} else {
			omega -= 0.0001
			if a2 <= -0.785 {
				rechtssom = true
			}
		}

		fmt.Println("INPUT V: ", vm)
		fmt.Println("INPUT W: ", omega)

		imd := imdraw.New(nil)

		// Wielen voertuig berekeningen
		wielLinksX := (pixel.ZV.X - L/2) * scale
		wielRechtsX := (pixel.ZV.X + L/2) * scale

		WielLinks := pixel.V(wielLinksX, pixel.ZV.Y)
		WielRechts := pixel.V(wielRechtsX, pixel.ZV.Y)

		// snelheid wielen berekenen uit omega en lineaire snelheid

		vWiel1 = (omega*L)/2 + float64(vm)
		vWiel2 = -(omega*L)/2 + float64(vm)

		//bereken rotatiecentrum achteras wielen
		r = L / 2 * (vWiel2 + vWiel1) / (vWiel1 - vWiel2)

		// stuurhoek bepalen

		if math.Atan(l1/r) > 0 {
			alpha = math.Atan(l1/r) + math.Asin(l2/math.Sqrt(l1*l1+r*r))
		} else {
			alpha = math.Atan(l1/r) - math.Asin(l2/math.Sqrt(l1*l1+r*r))
		}
		a2 = alpha

		// Teken body van het voertuig
		imd.Color = color.NRGBA{64, 64, 122, 255}
		imd.Push(pixel.ZV, lengteVoertuig1, lengteVoertuig2)
		imd.Line(5)

		// color = black
		imd.Color = color.NRGBA{255, 255, 255, 255}
		// teken linkerwiel
		imd.Push(WielLinks, pixel.V(WielLinks.X, vWiel1))
		imd.Line(5)

		// teken rechterwiel
		imd.Push(WielRechts, pixel.V(WielRechts.X, vWiel2))
		imd.Line(5)

		// teken knikpunt
		imd.Color = color.NRGBA{51, 217, 178, 255}
		imd.Push(lengteVoertuig1)
		imd.Circle(10, 0)

		rotc = pixel.V(r*scale, pixel.ZV.Y)
		// teken loodrechte op aandrijvende wielen
		imd.Push(pixel.ZV, rotc)
		imd.Line(2)

		//teken loodrechte van voorwielen
		imd.Push(lengteVoertuig2, rotc)
		imd.Line(2)

		imd.Color = color.NRGBA{100, 25, 178, 255}
		// teken van grootste wielsnelheid naar rot center
		if vWiel1 > vWiel2 {
			imd.Push(pixel.V(WielLinks.X, vWiel1), rotc)
		} else {
			imd.Push(pixel.V(WielRechts.X, vWiel2), rotc)
		}
		imd.Line(1)

		imd.Push(rotc)
		imd.Circle(7, 0)

		fmt.Println("OUTPUT alpha: ", a2)
		fmt.Println("OUTPUT WIELV1: ", vWiel1)
		fmt.Println("OUTPUT WIELV2: ", vWiel2)
		fmt.Println("------------------------")
		imd.Draw(win)
		win.Update()
	}
}

// Return L1 and l2 as vectors
func update() (pixel.Vec, pixel.Vec) {

	x1 := math.Sin(a1) * l1 * scale
	y1 := math.Cos(a1) * l1 * scale
	vec1 := pixel.V(x1, y1)

	x2 := x1 + math.Sin(a2)*l2*scale
	y2 := y1 + math.Cos(a2)*l2*scale
	vec2 := pixel.V(x2, y2)
	return vec1, vec2
}

func main() {
	switch len(os.Args) {
	case 2:
		vm, _ = strconv.ParseFloat(os.Args[1], 64)
	default:
		break
	}
	pixelgl.Run(run)
}
