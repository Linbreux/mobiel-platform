package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/aler9/goroslib"
	"github.com/aler9/goroslib/pkg/msgs/geometry_msgs"
	"github.com/aler9/goroslib/pkg/msgs/nav_msgs"
	"github.com/aler9/goroslib/pkg/msgs/std_msgs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

//straal wiel = 24cm

var (
	l1    = 765.0
	a1    = 0.0 / 180 * math.Pi
	l2    = 1110.0
	a2    = a1 + 0.0/180*math.Pi
	L     = 1000.0
	scale = 1 / 20.0
	//rechtssom  = true
	vWiel1              = 25.0
	vWiel2              = 50.0
	alpha               = 0.0
	r                   = 0.0
	omega               = 0.0
	vm                  = 0.0 // m/s
	vorigeTijd          time.Time
	deltat              time.Duration
	tempWaarde          float64
	stuursnelheid       = 0.5
	setpointSnelheid    = 0.0
	werkelijkeSnelheid  = 0.0
	setpointStuurhoek   = 0.0
	werkelijkeStuurhoek = 0.0
)

func rustig_oplopen(setpoint float64, werkelijkeSnelheid *float64, speed float64) {
	if setpoint > *werkelijkeSnelheid {
		*werkelijkeSnelheid += speed * deltat.Seconds()
	}
	if setpoint < *werkelijkeSnelheid {
		*werkelijkeSnelheid -= speed * deltat.Seconds()
	}
	fmt.Println("setp: ", setpoint, " werkwaarde: ", *werkelijkeSnelheid)

}

func onMessage(msg *geometry_msgs.Twist) {
	setpointSnelheid = msg.Linear.X
	// 1 angualar.z = 20°
	setpointStuurhoek = msg.Angular.Z
	//a2 = float64(-msg.Angular.Z) * 20 / 180 * math.Pi
}

func run() {
	//venster aanmaken
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Bounds:      pixel.R(0, 0, 1500, 750),
		VSync:       true,
		Undecorated: false,
	})
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)
	matrix := pixel.IM.ScaledXY(pixel.ZV, pixel.V(1, 1)).Moved(pixel.V(750, 100))
	win.SetMatrix(matrix)

	// ros node info voor sub
	n, err := goroslib.NewNode(goroslib.NodeConf{
		Name:          "Go_simulatie",
		MasterAddress: "127.0.0.1:11311",
	})
	if err != nil {
		panic(err)
	}
	defer n.Close()

	sub, err := goroslib.NewSubscriber(goroslib.SubscriberConf{
		Node:     n,
		Topic:    "/cmd_vel",
		Callback: onMessage,
	})
	if err != nil {
		panic(err)
	}
	defer sub.Close()

	pub, err := goroslib.NewPublisher(goroslib.PublisherConf{
		Node:  n,
		Topic: "/turn_angle",
		Msg:   &std_msgs.Float64{},
	})
	if err != nil {
		panic(err)
	}
	defer pub.Close()

	odom, err := goroslib.NewPublisher(goroslib.PublisherConf{
		Node:  n,
		Topic: "/voertuig_odom",
		Msg:   &nav_msgs.Odometry{},
	})
	if err != nil {
		panic(err)
	}
	defer odom.Close()

	var theta float64
	var snelheidsvector pixel.Vec
	rotc := pixel.V(r*scale, pixel.ZV.Y)
	vorigeTijd = time.Now()

	for !win.Closed() {
		win.SetClosed(win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ))

		win.Clear(color.NRGBA{44, 44, 84, 255})
		lengteVoertuig1, lengteVoertuig2 := update()

		imd := imdraw.New(nil)

		// bereken het tijdsverschil / FPS
		tijdNu := time.Now()
		deltat = tijdNu.Sub(vorigeTijd)
		vorigeTijd = tijdNu

		theta -= omega * deltat.Seconds()

		rotc.X = r * scale

		// Wielen voertuig berekeningen
		wielLinksX := (pixel.ZV.X - L/2) * scale
		wielRechtsX := (pixel.ZV.X + L/2) * scale

		WielLinks := pixel.V(wielLinksX, pixel.ZV.Y)
		WielRechts := pixel.V(wielRechtsX, pixel.ZV.Y)

		snelheidsvector.Y -= math.Cos(math.Pi-theta) * vm * 1000 * deltat.Seconds() * scale
		snelheidsvector.X -= math.Sin(math.Pi-theta) * vm * 1000 * deltat.Seconds() * scale

		// position
		setpoint := pixel.V(300, 300)
		vectorverschil := pixel.Vec{}

		vectorverschil.X = -(setpoint.X - snelheidsvector.X)
		vectorverschil.Y = -(setpoint.Y - snelheidsvector.Y)

		imd.SetMatrix(pixel.IM)
		imd.Push(setpoint)
		imd.Circle(7, 0)

		//thetaAlwaysWithin360 := math.Mod(theta, 2*math.Pi)

		imd.SetMatrix(pixel.IM.Rotated(pixel.ZV, theta).Moved(snelheidsvector))

		schuineZijde := l2 / math.Sin(a2)

		r = math.Cos(a1) * schuineZijde

		omega = vm * 1000 / r

		vWiel1 = (omega*L)/2 + float64(vm*1000)
		vWiel2 = -(omega*L)/2 + float64(vm*1000)

		rustig_oplopen(setpointSnelheid, &werkelijkeSnelheid, 0.5)
		rustig_oplopen(-setpointStuurhoek, &werkelijkeStuurhoek, 0.5)

		vm = werkelijkeSnelheid
		a2 = werkelijkeStuurhoek

		header := std_msgs.Header{
			Stamp:   n.TimeNow(),
			FrameId: "odom",
		}
		msg := &std_msgs.Float64{
			Data: -a1 - a2,
		}
		pub.Write(msg)

		msg2 := &nav_msgs.Odometry{
			Header:       header,
			ChildFrameId: "achter_wielbasis",
			Pose: geometry_msgs.PoseWithCovariance{
				Pose: geometry_msgs.Pose{
					Position: geometry_msgs.Point{
						X: snelheidsvector.Y / scale / 1000,
						Y: -snelheidsvector.X / scale / 1000,
					},
					Orientation: geometry_msgs.Quaternion{
						Z: theta,
					},
				},
			},
			Twist: geometry_msgs.TwistWithCovariance{
				Twist: geometry_msgs.Twist{
					Linear: geometry_msgs.Vector3{
						X: vm,
					},
				},
			},
		}

		odom.Write(msg2)

		// Teken body van het voertuig
		imd.Color = color.NRGBA{64, 64, 122, 255}
		imd.Push(pixel.ZV, lengteVoertuig1, lengteVoertuig2)
		imd.Line(5)

		// color = black
		imd.Color = color.NRGBA{255, 255, 255, 255}
		// teken linkerwiel
		imd.Push(WielLinks, pixel.V(WielLinks.X, vWiel1*scale))
		imd.Line(5)

		// teken rechterwiel
		imd.Push(WielRechts, pixel.V(WielRechts.X, vWiel2*scale))
		imd.Line(5)

		// teken knikpunt
		imd.Color = color.NRGBA{51, 217, 178, 255}
		imd.Push(lengteVoertuig1)
		imd.Circle(5, 0)

		// teken loodrechte op aandrijvende wielen
		imd.Push(pixel.ZV, rotc)
		imd.Line(2)

		//teken loodrechte van voorwielen
		imd.Push(lengteVoertuig2, rotc)
		imd.Line(2)

		imd.Color = color.NRGBA{100, 25, 178, 255}
		// teken van grootste wielsnelheid naar rot center
		if vWiel1 > vWiel2 {
			imd.Push(pixel.V(WielLinks.X, vWiel1*scale), rotc)
		} else {
			imd.Push(pixel.V(WielRechts.X, vWiel2*scale), rotc)
		}
		imd.Line(1)

		imd.Push(rotc)
		imd.Circle(7, 0)

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
