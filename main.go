package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/fogleman/gg" // Replace the cairo import with this
)

const (
	width           = 800
	height          = 600
	dotRadius       = 15
	numDots         = 240
	centerLineY     = height / 2
	iterations      = 10000
	avoidanceHeight = 30 // New constant for the height of the area to avoid
)

type Dot struct {
	x, y           float64
	radius         float64
	xScale, yScale float64
	rotation       float64 // New field for rotation
}

func main() {
	rand.Seed(time.Now().UnixNano())

	dots := make([]Dot, numDots)
	for i := range dots {
		dots[i] = Dot{
			x:        rand.Float64() * width,
			y:        rand.Float64() * height,
			radius:   dotRadius * (0.8 + rand.Float64()*0.4), // Updated this line
			xScale:   0.9 + rand.Float64()*0.2,
			yScale:   0.9 + rand.Float64()*0.2,
			rotation: rand.Float64() * 2 * math.Pi, // Random rotation between 0 and 2π
		}
	}

	for i := 0; i < iterations; i++ {
		applyForces(dots)
	}

	file, err := os.Create("output.svg")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writeSVG(file, dots)

	// Add PNG conversion
	writePNG(dots)
}

func applyForces(dots []Dot) {
	idealSpacing := math.Sqrt((width * height) / float64(numDots))

	for i := range dots {
		dx, dy := 0.0, 0.0

		for j := range dots {
			if i != j {
				diffX := dots[i].x - dots[j].x
				diffY := dots[i].y - dots[j].y
				distance := math.Sqrt(diffX*diffX + diffY*diffY)

				if distance < idealSpacing {
					force := (idealSpacing - distance) / idealSpacing
					dx += force * diffX / distance
					dy += force * diffY / distance
				}
			}
		}

		// Move dot
		dots[i].x += dx
		dots[i].y += dy

		// Avoid center area
		if math.Abs(dots[i].y-centerLineY) < dotRadius+avoidanceHeight/2 {
			if dots[i].y < centerLineY {
				dots[i].y = centerLineY - (dotRadius + avoidanceHeight/2)
			} else {
				dots[i].y = centerLineY + (dotRadius + avoidanceHeight/2)
			}
		}

		// Keep within bounds (adjusted for variable radius)
		dots[i].x = math.Max(dots[i].radius, math.Min(width-dots[i].radius, dots[i].x))
		dots[i].y = math.Max(dots[i].radius, math.Min(height-dots[i].radius, dots[i].y))
	}
}

func writeSVG(file *os.File, dots []Dot) {
	file.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	file.WriteString(`<rect width="100%" height="100%" fill="black"/>`)

	// Remove the line drawing code

	for _, dot := range dots {
		file.WriteString(fmt.Sprintf(`<g transform="rotate(%f %f %f)">`,
			dot.rotation*180/math.Pi, dot.x, dot.y))
		file.WriteString(fmt.Sprintf(`<ellipse cx="%f" cy="%f" rx="%f" ry="%f" fill="white"/>`,
			dot.x, dot.y, dot.radius*dot.xScale, dot.radius*dot.yScale))
		file.WriteString(`</g>`)
	}

	file.WriteString(`</svg>`)
}

func writePNG(dots []Dot) {
	dc := gg.NewContext(width, height)

	// Set background to black
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	// Draw white dots
	dc.SetRGB(1, 1, 1)
	for _, dot := range dots {
		dc.Push()
		dc.Translate(dot.x, dot.y)
		dc.Rotate(dot.rotation)
		dc.DrawEllipse(0, 0, dot.radius*dot.xScale, dot.radius*dot.yScale)
		dc.Fill()
		dc.Pop()
	}

	dc.SavePNG("output.png")
}
