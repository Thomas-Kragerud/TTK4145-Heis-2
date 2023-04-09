package gui

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image"
	"os"
)

import _ "image/png"

var (
	win          *pixelgl.Window
	elevatorPic  pixel.Picture
	elevator     pixel.Sprite
	elevatorPos  pixel.Vec
	screenWidth  = 300.0
	screenHeight = 600.0
)

func InitGUI() {
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Elevator Simulation",
		Bounds: pixel.R(0, 0, screenWidth, screenHeight),
		VSync:  true,
	}

	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	elevatorPic, err = loadPicture("/Users/thomas/GolandProjects/TTK4145-Heis-2/gui/resources/Elevator_fig2.png")
	if err != nil {
		panic(err)
	}

	elevator = *pixel.NewSprite(elevatorPic, elevatorPic.Bounds())

	for !win.Closed() {
		win.Clear(colornames.Skyblue)
		elevator.Draw(win, pixel.IM.Moved(elevatorPos))
		win.Update()
	}
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}

func UpdateElevatorPosition(floor int) {
	elevatorPos = pixel.V(screenWidth/2, (screenHeight/4)*float64(floor))
}
