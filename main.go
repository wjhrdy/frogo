package main

import (
	"fmt"
	"image/color"
	"image/png"
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

	// Create SVG output
	file, err := os.Create("output.svg")
	if err != nil {
		fmt.Println("Error creating SVG file:", err)
		return
	}
	defer file.Close()

	writeSVG(file, dots)

	// Create PNG output
	writePNG(dots)

	// Generate the stippled SVG image
	writeStippledSVG()
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

// Helper struct to represent 2D points
type vec2 struct {
	X, Y float64
}

// Poisson Disk Sampling function
func poissonDiskSampling(width, height int, minDist float64, k int) []vec2 {
	cellSize := minDist / math.Sqrt2

	gridWidth := int(math.Ceil(float64(width) / cellSize))
	gridHeight := int(math.Ceil(float64(height) / cellSize))

	grid := make([][]*vec2, gridWidth)
	for i := range grid {
		grid[i] = make([]*vec2, gridHeight)
	}

	var processList []vec2
	var points []vec2

	// Random initial point
	p := vec2{rand.Float64() * float64(width), rand.Float64() * float64(height)}
	processList = append(processList, p)
	points = append(points, p)

	gx := int(p.X / cellSize)
	gy := int(p.Y / cellSize)
	grid[gx][gy] = &p

	for len(processList) > 0 {
		idx := rand.Intn(len(processList))
		p := processList[idx]
		found := false

		for i := 0; i < k; i++ {
			angle := rand.Float64() * 2 * math.Pi
			radius := minDist + rand.Float64()*minDist
			newX := p.X + radius*math.Cos(angle)
			newY := p.Y + radius*math.Sin(angle)

			if newX >= 0 && newX < float64(width) && newY >= 0 && newY < float64(height) {
				// Calculate distance from center line
				distFromCenter := math.Abs(newY - float64(height)/2)
				densityFactor := distFromCenter / (float64(height) / 1.5)

				// Adjust minDist based on density factor
				// Decrease the minimum distance near the center
				adjustedMinDist := minDist * (0.7 + densityFactor)

				gx = int(newX / cellSize)
				gy = int(newY / cellSize)

				tooClose := false
				for i := -2; i <= 2; i++ {
					for j := -2; j <= 2; j++ {
						nx := gx + i
						ny := gy + j
						if nx >= 0 && nx < gridWidth && ny >= 0 && ny < gridHeight {
							neighbor := grid[nx][ny]
							if neighbor != nil {
								dx := neighbor.X - newX
								dy := neighbor.Y - newY
								if dx*dx+dy*dy < adjustedMinDist*adjustedMinDist {
									tooClose = true
									break
								}
							}
						}
					}
					if tooClose {
						break
					}
				}
				if !tooClose {
					newPoint := vec2{newX, newY}
					processList = append(processList, newPoint)
					points = append(points, newPoint)
					grid[gx][gy] = &newPoint
					found = true
				}
			}
		}

		if !found {
			processList = append(processList[:idx], processList[idx+1:]...)
		}
	}

	return points
}

func writeStippledSVG() {
	// Open the source grayscale image
	imgFile, err := os.Open("output.png") // Replace with your image file
	if err != nil {
		fmt.Println("Error opening image:", err)
		return
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create an SVG file
	svgFile, err := os.Create("stippled_output.svg")
	if err != nil {
		fmt.Println("Error creating SVG file:", err)
		return
	}
	defer svgFile.Close()

	// Write SVG header
	svgFile.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svgFile.WriteString(`<rect width="100%" height="100%" fill="white"/>`)

	// Parameters for Poisson Disk Sampling
	minDist := 8.0 // Further decreased from 7.0
	k := 30

	// Generate points using Poisson Disk Sampling
	points := poissonDiskSampling(width, height, minDist, k)

	// Constant dot size
	dotRadius := minDist / 4.5 // Decreased from 4

	// For each point, determine whether to draw a dot based on image brightness
	for _, p := range points {
		x, y := int(p.X), int(p.Y)
		if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
			continue
		}
		grayColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
		brightness := float64(grayColor.Y) / 255

		// Adjust the threshold for drawing dots
		distFromCenter := math.Abs(float64(y)-float64(height)/2) / (float64(height) / 2)
		densityFactor := 1.3 - 0.6*distFromCenter // Higher density in the center

		if rand.Float64() > brightness*densityFactor {
			// Generate random rotation and scaling factors
			rotation := rand.Float64() * 2 * math.Pi
			xScale := 0.9 + rand.Float64()*0.2
			yScale := 0.9 + rand.Float64()*0.2

			// Write an ellipse element with rotation for each dot
			svgFile.WriteString(fmt.Sprintf(`<g transform="rotate(%f %f %f)">`, rotation*180/math.Pi, p.X, p.Y))
			svgFile.WriteString(fmt.Sprintf(`<ellipse cx="%f" cy="%f" rx="%f" ry="%f" fill="black"/>`,
				p.X, p.Y, dotRadius*xScale, dotRadius*yScale))
			svgFile.WriteString(`</g>`)
		}
	}

	// Close SVG tag
	svgFile.WriteString(`</svg>`)
}
