package gui

import (
	"Project/config"
	"Project/elevio"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	"math"
	"os"
	"sync"
	"time"
)

import _ "image/png"

var (
	win               *pixelgl.Window
	elevatorPic       pixel.Picture
	elevator          pixel.Sprite
	elevatorPos       pixel.Vec
	screenWidth       float64
	screenHeight      float64
	lightOn           bool
	arrow             *imdraw.IMDraw
	arrowMutex        sync.Mutex
	elevatorDirection elevio.MotorDirection
	guiOn             bool
)

func InitGUI(true bool) {
	screenHeight = float64(config.NumFloors) * 150
	screenWidth = 300.0
	if true {
		guiOn = true
		pixelgl.Run(run)
	} else {
		guiOn = false
	}

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
	drawArrow(elevio.MD_Stop) // Init the arrow without direction

	// Drawing elevator shaft
	imd := imdraw.New(nil)
	shaftWidth := 20.0
	shaftHight := screenHeight - 50.0
	shaftPosX := (screenWidth / 2) - (shaftWidth / 2)
	shaftPosY := (screenHeight / 2) - (shaftHight / 2)
	imd.Color = colornames.Lightgrey
	imd.Push(pixel.V(shaftPosX, shaftPosY), pixel.V(shaftPosX+shaftWidth, shaftPosY+shaftHight))
	imd.Rectangle(0)

	for !win.Closed() {
		win.Clear(colornames.Skyblue) // Draw background
		imd.Draw(win)                 // Draw elevator shaft
		drawLight()                   // Draw the light
		arrowMutex.Lock()
		if arrow != nil {
			arrow.Draw(win)
		}
		arrowMutex.Unlock()
		elevator.Draw(win, pixel.IM.Moved(elevatorPos)) // Draw elevator
		win.Update()                                    // Update the window
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

func UpdateElevatorPosition(newFloor int) {
	if guiOn {
		targetPos := pixel.V(screenWidth/2, (screenHeight/float64(config.NumFloors))*float64(newFloor)+70)
		steps := 100
		duration := time.Millisecond * 5

		// Calculate the difference between the target position and the current position
		delta := targetPos.Sub(elevatorPos).Scaled(1 / float64(steps))

		// Gradually move the elevator to the target position
		for i := 0; i < steps; i++ {
			elevatorPos = elevatorPos.Add(delta)
			time.Sleep(duration)
		}
	} else {
		return
	}
}

func SetDoorOpenLight(on bool) {
	lightOn = on
}

func SetArrowDirection(direction elevio.MotorDirection) {
	if guiOn {
		arrowMutex.Lock()
		drawArrow(direction)
		elevatorDirection = direction
		arrowMutex.Unlock()
	} else {
		return
	}
}

func drawLight() {
	imd := imdraw.New(nil)
	lightRadius := 20.0
	lightPosX := screenWidth - lightRadius - 10
	lightPosY := screenHeight - lightRadius - 10
	if lightOn {
		imd.Color = color.RGBA{R: 208, G: 49, B: 45, A: 255} // Red color with full opacity
	} else {
		imd.Color = color.RGBA{R: 140, G: 0, B: 0, A: 255}
	}
	imd.Push(pixel.V(lightPosX, lightPosY))
	//imd.Circle(lightRadius, 0)
	imd.CircleArc(lightRadius, 0, 2*math.Pi, 0)
	imd.Draw(win)
}

func drawArrow(direction elevio.MotorDirection) {
	scalingFactor := 2.0
	arrowSize := 20.0 * scalingFactor
	arrowPosX := screenWidth - arrowSize - 10
	arrowPosY := screenHeight - arrowSize*2 - 30

	arrow = imdraw.New(nil)
	arrow.Color = colornames.Black

	if direction == elevio.MD_Up {
		arrow.Push(pixel.V(arrowPosX, arrowPosY))
		arrow.Push(pixel.V(arrowPosX+arrowSize/2, arrowPosY+arrowSize))
		arrow.Push(pixel.V(arrowPosX+arrowSize, arrowPosY))
	} else if direction == elevio.MD_Down {
		arrow.Push(pixel.V(arrowPosX, arrowPosY+arrowSize))
		arrow.Push(pixel.V(arrowPosX+arrowSize/2, arrowPosY))
		arrow.Push(pixel.V(arrowPosX+arrowSize, arrowPosY+arrowSize))
	} else {
		arrow = nil
		return
	}
	arrow.Polygon(0)
}
