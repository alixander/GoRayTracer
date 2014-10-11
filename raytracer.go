package main

import (
    "fmt"
    //"./vector"
    "math"
    "os"
    "image"
    "image/color"
    "image/png"
    "time"
)

type Point struct {
    X float64
    Y float64
    Z float64
}

type Color struct {
    R float64
    G float64
    B float64
}

func (a Point) equals(b Point) bool {
    if a.X == b.X && a.Y == b.Y && a.Z == b.Z {
        return true
    }
    return false
}

type Ray struct {
    start Point
    end Point
}

type Sphere struct {
    center Point
    radius float64
}

//globals
var (
    width int = 800
    height int = 800
    viewport = image.Rect(0, 0, width, height)
    viewportColors = image.NewRGBA(viewport)
    eye = Point{X:600, Y:400, Z:-50}
    sphere = Sphere{Point{X: 400, Y:400, Z:100}, 150}
    tempColor = pointNormalize(Point{X: 50, Y: 150, Z: 200})
    tempLight = pointNormalize(Point{X: -200, Y: -150, Z: 0})
)


func drawPixel(canvas *image.RGBA, x float64, y float64, r uint8, g uint8, b uint8) {
    canvas.SetRGBA(int(x), int(y), color.RGBA {
        R: r,
        G: g,
        B: b,
        A: 0xff,
    })
}

func saveScene(canvas *image.RGBA) {
    outputImage, _ := os.Create("output.png")
    defer outputImage.Close()
    png.Encode(outputImage, canvas)
}

func getPixelsRoutine(pixelChannel chan Point, doneChannel chan bool) {
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            doneChannel <- false
            pixelChannel <- Point{X:float64(x), Y:float64(y), Z:0}
        }
    }
    doneChannel <- true
}

func pointScale(a Point, b float64) Point {
    return Point{
        X: a.X * b,
        Y: a.Y * b,
        Z: a.Z * b,
    }
}

func pointSub(a Point, b Point) Point {
    return Point{
        X: a.X - b.X,
        Y: a.Y - b.Y,
        Z: a.Z - b.Z,
    }
}

func pointAdd(a Point, b Point) Point {
    return Point{
        X: a.X + b.X,
        Y: a.Y + b.Y,
        Z: a.Z + b.Z,
    }
}

func pointMult(a Point, b Point) Point {
    return Point{
        X: a.X * b.X,
        Y: a.Y * b.Y,
        Z: a.Z * b.Z,
    }
}

func pointDiv(a Point, b float64) Point {
    return Point{
        X: a.X/b,
        Y: a.Y/b,
        Z: a.Z/b,
    }
}

func scale(a Point, b float64) Point {
    return Point{
        X: a.X * b,
        Y: a.Y * b,
        Z: a.Z * b,
    }
}

func getDotProduct(a Point, b Point) float64 {
   return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func pointNormalize(a Point) Point {
    magnitude := math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z))
    return Point{
        X: float64(float64(a.X)/magnitude),
        Y: float64(float64(a.Y)/magnitude),
        Z: float64(float64(a.Z)/magnitude),
    }
}

func computeRay(pixel Point) func(t float64) Point {
    return func(t float64) Point {
        return pointAdd(eye, pointScale(pointSub(pixel, eye), t))
    }
}

func withinRadius(sphere Sphere, x float64, y float64, z float64) bool {
    distToCenter := math.Pow(float64(sphere.center.X - x), 2) + 
                    math.Pow(float64(sphere.center.Y - y), 2) +
                    math.Pow(float64(sphere.center.Z - z), 2) 

    if float64(distToCenter - math.Pow(float64(sphere.radius), 2)) <= 0 {
        return true
    }
    return false
}

func calculateDiffuseColor(normal Point) Point {
    tempDiffuse := pointNormalize(Point{X: 100, Y: 200, Z: 60})
    dotProduct := getDotProduct(normal, tempLight)
    if dotProduct < 0 {
        dotProduct = 0
    }
    //fmt.Println(scale(tempDiffuse, dotProduct))
    return pointMult(scale(tempDiffuse, dotProduct), tempColor)
}

func calculateAmbientColor() Point {
    tempAmbient := pointNormalize(Point{X: 200, Y: 100, Z: 70})
    return pointMult(tempColor, tempAmbient) 
}

func hitObject(sphere Sphere, rayPoint Point) (bool, Point) {
    if withinRadius(sphere, rayPoint.X, rayPoint.Y, rayPoint.Z) {
        //Normal on a sphere (p-c)/R?
        surfaceNormal := pointDiv(pointSub(rayPoint, sphere.center), sphere.radius)
        diffuseColor := calculateDiffuseColor(surfaceNormal)
        ambientColor := calculateAmbientColor()
        finalColor := pointAdd(diffuseColor, ambientColor)
        return true, finalColor
    }
    return false, Point{}
}

func shootRay(ray func(t float64) Point) (float64, Point) {
    for t := 1; t < 100; t++ {
        isHit, color := hitObject(sphere, ray(float64(t)))
        if isHit {
            return float64(t), color
        }
    }
    return -1, Point{}
}

func renderScene() {
    doneChannel := make(chan bool)
    pixelChannel := make(chan Point)
    go getPixelsRoutine(pixelChannel, doneChannel)

    for done := <- doneChannel; done == false; done = <- doneChannel{
        pixel := <- pixelChannel
        ray := computeRay(pixel)
        rayHit, color := shootRay(ray)
        color = scale(color, 255)
        if rayHit != -1 {
            if color.X > 255{
                color.X = 255
            }
            if color.Y > 255{
                color.Y = 255
            }
            if color.Z > 255{
                color.Z = 255
            }
            drawPixel(viewportColors, pixel.X, pixel.Y, uint8(color.X), uint8(color.Y), uint8(color.Z))
        } else {
            drawPixel(viewportColors, pixel.X, pixel.Y, 0, 100, 180)
        }
    }
}

func main() {
    fmt.Println("\n------------Starting--------------\n")
    startTime := time.Now()

    renderScene()
    saveScene(viewportColors)

    fmt.Println("Program finished running in", time.Since(startTime))
}
