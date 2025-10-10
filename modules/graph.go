package modules

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var Graph = tview.NewImage().
	SetDithering(tview.DitheringNone).
	SetColors(tview.TrueColor)

const GraphSize = 16 // graph proportionality in the flex that it is inside of, all other components scale off of it
var lastImageWidth = 800
var lastImageHeight = 600
var queRedraw = false

/*
func graphMouseHandling(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {

}
*/
func GraphTraversial() *tview.Flex {
	/*
		buttons needed:
			- Home button
			- zoom out
			- zoom in
			- up, right, left, down
	*/

	return tview.NewFlex().AddItem(tview.NewBox().SetBorder(true), 0, 1, false)
}

// Updates BEFORE frame is drawn, returns true if drawing should not occur
func GraphUpdate() bool {
	// redraws graph if the dimensions of the image have changed
	_, _, newWidth, newHeight := Graph.GetRect()
	if newWidth != lastImageWidth || newHeight != lastImageHeight {
		queRedraw = true // delays redraw to when the window is no longer being resized
	} else if queRedraw && newWidth == lastImageWidth && newHeight == lastImageHeight {
		lastImageWidth = newWidth
		lastImageHeight = newHeight
		RedrawGraph()
		queRedraw = false
	}
	return false
}

func RedrawGraph() {
	Graph.SetImage(createGraph())
	InfoPrint("Graph Redrawn")
}

const (
	xMin        = -10.0 // these should be taken from information.go
	xMax        = 10.0  // ^
	yMin        = -5.0  // ^
	yMax        = 5.0   // ^
	lineWidth   = 20    // Width of the function line
	axisWidth   = 20    // Width of the axes
	tickLength  = 50    // Length of tick marks
	tickWidth   = 20    // Width of tick marks
	tickSpacing = 50    // Space between ticks in graph units
)

// map a value from one range to another
func mapRange(value, inMin, inMax, outMin, outMax float64) float64 {
	return (value-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

// generate an image of all function expressions on a cartesian coordinate plane
func createGraph() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, lastImageWidth, lastImageHeight))

	// Fill background with white
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw axes
	xAxis := int(mapRange(0, yMin, yMax, float64(lastImageHeight-1), 0))
	yAxis := int(mapRange(0, xMin, xMax, 0, float64(lastImageWidth-1)))
	for x := 0; x < lastImageWidth; x++ {
		for dy := -axisWidth / 2; dy <= axisWidth/2; dy++ {
			for dx := -axisWidth / 2; dx <= axisWidth/2; dx++ {
				if dx*dx+dy*dy <= (axisWidth/2)*(axisWidth/2) {
					y := xAxis + dy
					if y >= 0 && y < lastImageHeight && x+dx >= 0 && x+dx < lastImageWidth {
						img.Set(x+dx, y, color.Black)
					}
				}
			}
		}
	}
	for y := 0; y < lastImageHeight; y++ {
		for dx := -axisWidth / 2; dx <= axisWidth/2; dx++ {
			for dy := -axisWidth / 2; dy <= axisWidth/2; dy++ {
				if dx*dx+dy*dy <= (axisWidth/2)*(axisWidth/2) {
					x := yAxis + dx
					if x >= 0 && x < lastImageWidth && y+dy >= 0 && y+dy < lastImageHeight {
						img.Set(x, y+dy, color.Black)
					}
				}
			}
		}
	}

	///// Not visable when image is compressed
	// // Draw tick marks
	// for x := math.Ceil(xMin/tickSpacing) * tickSpacing; x <= xMax; x += tickSpacing {
	// 	px := int(mapRange(x, xMin, xMax, 0, float64(lastImageWidth-1)))
	// 	if px >= 0 && px < lastImageWidth {
	// 		for dy := -tickLength / 2; dy <= tickLength/2; dy++ {
	// 			for dx := -tickWidth / 2; dx <= tickWidth/2; dx++ {
	// 				if dx*dx+dy*dy <= (tickWidth/2)*(tickWidth/2) {
	// 					y := xAxis + dy
	// 					if y >= 0 && y < lastImageHeight && px+dx >= 0 && px+dx < lastImageWidth {
	// 						img.Set(px+dx, y, color.Black)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// for y := math.Ceil(yMin/tickSpacing) * tickSpacing; y <= yMax; y += tickSpacing {
	// 	py := int(mapRange(y, yMin, yMax, float64(lastImageHeight-1), 0))
	// 	if py >= 0 && py < lastImageHeight {
	// 		for dx := -tickLength / 2; dx <= tickLength/2; dx++ {
	// 			for dy := -tickWidth / 2; dy <= tickWidth/2; dy++ {
	// 				if dx*dx+dy*dy <= (tickWidth/2)*(tickWidth/2) {
	// 					x := yAxis + dx
	// 					if x >= 0 && x < lastImageWidth && py+dy >= 0 && py+dy < lastImageHeight {
	// 						img.Set(x, py+dy, color.Black)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// Plot all function expressions

	for _, expression := range Expressions {
		if expression.err != nil || !expression.enabledCheckbox.IsChecked() {
			continue
		}
		resolution := lastImageWidth * 2
		prevY := -1
		prevX := -1

		for i := 0; i < resolution; i++ {
			// Map x coordinate
			xVal := mapRange(float64(i), 0, float64(resolution-1), xMin, xMax)
			yVal, err := expression.function(xVal)
			if err != nil {
				continue
			}

			// Map to image coordinates
			if yVal >= yMin && yVal <= yMax {
				px := int(mapRange(float64(i), 0, float64(resolution-1), 0, float64(lastImageWidth-1)))
				py := int(mapRange(yVal, yMin, yMax, float64(lastImageHeight-1), 0))

				if py >= 0 && py < lastImageHeight && px >= 0 && px < lastImageWidth {
					// Draw point with circular thickness
					for dx := -lineWidth / 2; dx <= lineWidth/2; dx++ {
						for dy := -lineWidth / 2; dy <= lineWidth/2; dy++ {
							if dx*dx+dy*dy <= (lineWidth/2)*(lineWidth/2) {
								if px+dx >= 0 && px+dx < lastImageWidth && py+dy >= 0 && py+dy < lastImageHeight {
									img.Set(px+dx, py+dy, convertColorType(expression.color))
								}
							}
						}
					}

					// Fill gaps between points
					if prevY != -1 && prevX != -1 {
						dx := px - prevX
						dy := py - prevY
						steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))
						if steps > 0 {
							for step := 0; step <= steps; step++ {
								x := prevX + dx*step/steps
								y := prevY + dy*step/steps
								// Draw circular points along the line
								for ddx := -lineWidth / 2; ddx <= lineWidth/2; ddx++ {
									for ddy := -lineWidth / 2; ddy <= lineWidth/2; ddy++ {
										if ddx*ddx+ddy*ddy <= (lineWidth/2)*(lineWidth/2) {
											if x+ddx >= 0 && x+ddx < lastImageWidth && y+ddy >= 0 && y+ddy < lastImageHeight {
												img.Set(x+ddx, y+ddy, convertColorType(expression.color))
											}
										}
									}
								}
							}
						}
					}
					prevX = px
					prevY = py
				}
			}
		}
	}

	return img
}

// Converts a tcell.Color to color.Color
func convertColorType(c tcell.Color) color.Color {
	r, g, b := c.RGB()
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}
