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
    direction Point
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
    eye = Point{X:400, Y:400, Z:-550}
    sphereA = Sphere{center: Point{X: 400, Y:400, Z:100}, radius: 100}
    sphereB = Sphere{center: Point{X: 600, Y:400, Z:100}, radius: 100}
    sphereC = Sphere{center: Point{X: 200, Y:400, Z:100}, radius: 100}
    sphereD = Sphere{center: Point{X: 400, Y:200, Z:100}, radius: 100}
    sphereE = Sphere{center: Point{X: 400, Y:600, Z:100}, radius: 100}
    tempColor = pointNormalize(Point{X: 50, Y: 150, Z: 200})
    tempLight = pointNormalize(Point{X: 0, Y: -150, Z: -100})
    allSpheres = []Sphere {sphereA, sphereB, sphereC, sphereD, sphereE}
)


func drawPixel(canvas *image.RGBA, x float64, y float64, r float64, g float64, b float64) {
    canvas.SetRGBA(int(x), int(y), color.RGBA {
        R: floatToRGB(r),
        G: floatToRGB(g),
        B: floatToRGB(b),
        A: 255,
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

func floatToRGB(color float64) uint8 {
    return uint8(math.Floor(color));
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
        X: float64(a.X)/magnitude,
        Y: float64(a.Y)/magnitude,
        Z: float64(a.Z)/magnitude,
    }
}

func getRayPoint(t float64, ray Ray) Point {
    return pointAdd(eye, pointScale(ray.direction, t))
}

//p(t) = e + t(s-e)
func computeRay(pixel Point) Ray {
    return Ray{start: eye, direction: pointSub(pixel, eye)}
}

func calculateDiffuseColor(normal Point) Point {
    tempDiffuse := pointNormalize(Point{X: 100, Y: 200, Z: 60})
    theta := math.Max(0, getDotProduct(normal, tempLight))
    return pointMult(scale(tempDiffuse, theta), tempColor)
}

func calculateAmbientColor() Point {
    tempAmbient := pointNormalize(Point{X: 0, Y: 200, Z: 70})
    return pointMult(tempColor, tempAmbient) 
}

func getReflectanceLight(light Point, normal Point) Point {
    lightDotNormal := math.Max(0.0, getDotProduct(light, normal))
    return pointSub(light, pointScale(normal, 2.0*lightDotNormal))
}

func calculateSpecularColor(normal Point) Point {
    tempSpecular := pointNormalize(Point{X: 70, Y: 80, Z: 80})
    tempShininess := 16.0 
    reflectanceLight := getReflectanceLight(tempLight, normal)
    specularTerm := math.Max(0, getDotProduct(reflectanceLight, eye))
    return pointMult(tempSpecular, pointScale(tempColor, math.Pow(specularTerm, tempShininess)))
}

// Formula from http://www.csee.umbc.edu/~olano/435f02/ray-sphere.html
func (sphere Sphere) hit(ray Ray) (float64, Point) {
    a := getDotProduct(ray.direction, ray.direction) 
    b := 2.0 * getDotProduct(ray.direction, pointSub(eye, sphere.center)) 
    c := getDotProduct(pointSub(eye, sphere.center), pointSub(eye, sphere.center)) - math.Pow(sphere.radius, 2)
    discriminant := math.Pow(b, 2) - 4.0*a*c

    if discriminant < 0 {
        return -1, Point{}
    }

    tNeg := (-b - math.Sqrt(discriminant))/(2*a)
    tPos := (-b + math.Sqrt(discriminant))/(2*a)
    var t float64
    t = math.Min(tNeg, tPos)

    rayPoint := getRayPoint(t, ray)
    surfaceNormal := pointDiv(pointSub(rayPoint, sphere.center), sphere.radius)

    diffuseColor := calculateDiffuseColor(surfaceNormal)
    ambientColor := calculateAmbientColor()
    specularColor := calculateSpecularColor(surfaceNormal)

    finalColor := pointAdd(pointAdd(ambientColor, diffuseColor), specularColor)

    return t, finalColor
}

func renderScene() {
    doneChannel := make(chan bool)
    pixelChannel := make(chan Point)
    go getPixelsRoutine(pixelChannel, doneChannel)

    for done := <- doneChannel; done == false; done = <- doneChannel{
        pixel := <- pixelChannel
        drawPixel(viewportColors, pixel.X, pixel.Y, 0, 100, 180) //This is probably extra work
        ray := computeRay(pixel)
        var color Point;
        isHit := false;
        minT := math.MaxFloat64;
        for _, singleSphere := range allSpheres {
            rayHit, rayColor := singleSphere.hit(ray)
            if (rayHit != -1 && rayHit < minT) {
                isHit = true;
                minT = rayHit;
                color = scale(rayColor, 255)
                if color.X > 255{
                    color.X = 255
                }
                if color.Y > 255{
                    color.Y = 255
                }
                if color.Z > 255{
                    color.Z = 255
                }
            }
        }
        if (isHit) {
            drawPixel(viewportColors, pixel.X, pixel.Y, color.X, color.Y, color.Z)
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
