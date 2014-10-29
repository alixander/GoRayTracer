package main

import (
    "fmt"
    "math"
    "./vector"
    "os"
    "image"
    "image/color"
    "image/png"
    "time"
)

type Color struct {
    R float64
    G float64
    B float64
}


type Ray struct {
    start raytracer.Vector
    direction raytracer.Vector
}

type Sphere struct {
    center raytracer.Vector
    radius float64
}

//globals
var (
    width int = 800
    height int = 800
    viewport = image.Rect(0, 0, width, height)
    viewportColors = image.NewRGBA(viewport)
    eye = raytracer.Vector{X:400, Y:400, Z:700}
    sphereA = Sphere{center: raytracer.Vector{X: 400, Y:400, Z:-200}, radius: 320}
    //sphereA = Sphere{center: raytracer.Vector{X: 400, Y:400, Z:-200}, radius: 120}
    //sphereB = Sphere{center: raytracer.Vector{X: 650, Y:400, Z:-200}, radius: 120}
    //sphereC = Sphere{center: raytracer.Vector{X: 150, Y:400, Z:-200}, radius: 120}
    //sphereD = Sphere{center: raytracer.Vector{X: 400, Y:150, Z:-200}, radius: 120}
    //sphereE = Sphere{center: raytracer.Vector{X: 400, Y:650, Z:-200}, radius: 120}
    tempColor = raytracer.Vector{X: 0.6, Y: 0.6, Z: 0.6}
    lightA = (raytracer.Vector{X: 300, Y: -400, Z: 350}).Normalize()
    lightB = (raytracer.Vector{X: -300, Y: 400, Z: 350}).Normalize()
    //allSpheres = []Sphere {sphereA, sphereB, sphereC, sphereD, sphereE}
    allSpheres = []Sphere {sphereA}
    pointLights = []raytracer.Vector {lightA}
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

func getPixelsRoutine(pixelChannel chan raytracer.Vector, doneChannel chan bool) {
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            doneChannel <- false
            pixelChannel <- raytracer.Vector{X:float64(x), Y:float64(y), Z:0}
        }
    }
    doneChannel <- true
}

func floatToRGB(color float64) uint8 {
    return uint8(math.Floor(color*255))
}


func getRayIntersection(t float64, ray Ray) raytracer.Vector {
    return eye.VectorAdd(ray.direction.VectorScale(t))
}

//p(t) = e + t(s-e)
func computeRay(pixel raytracer.Vector) Ray {
    return Ray{start: eye, direction: pixel.VectorSub(eye)}
}

func emptyVector() raytracer.Vector {
    return raytracer.Vector{X:0, Y:0, Z:0}
}

func calculateDiffuseColor(normal raytracer.Vector) raytracer.Vector {
    tempDiffuse := raytracer.Vector{X: 1, Y: 1, Z: 0}
    diffuseColor := emptyVector()

    var theta float64
    var color raytracer.Vector

    for _, light := range pointLights {
        theta = math.Max(0, normal.DotProduct(light))
        color = tempDiffuse.VectorScale(theta).VectorMult(tempColor)
        diffuseColor = diffuseColor.VectorAdd(color)
    }
    return diffuseColor
}

func calculateAmbientColor() raytracer.Vector {
    tempAmbient := raytracer.Vector{X: 0.1, Y: 0.1, Z: 0}
    return tempColor.VectorMult(tempAmbient) 
}


//func changeZDirection(light raytracer.Vector) raytracer.Vector {
//    return raytracer.Vector{
//        X: light.X,
//        Y: light.Y,
//        Z: -light.Z,
//    }
//}

// R = I - 2N(I . N)
func getReflectanceLight(light raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    lightDotNormal := math.Max(0.0, light.DotProduct(normal))
    return normal.VectorScale(2.0*lightDotNormal).VectorSub(light)
}

func calculateSpecularColor(intersection raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    tempSpecular := raytracer.Vector{X: 0.8, Y: 0.8, Z: 0.8}
    tempShininess := 16.0 
    specularColor := emptyVector()

    var reflectanceLight raytracer.Vector
    var incomingLight raytracer.Vector
    var color raytracer.Vector
    var directionToViewer raytracer.Vector
    var specularTerm float64

    for _, light := range pointLights {
        incomingLight = light
        reflectanceLight = getReflectanceLight(incomingLight, normal).Normalize()
        directionToViewer = eye.VectorSub(intersection)
        specularTerm = math.Max(0, reflectanceLight.DotProduct(directionToViewer.Normalize()))
        color = tempSpecular.VectorMult(tempColor.VectorScale(math.Pow(specularTerm, tempShininess)))
        specularColor = specularColor.VectorAdd(color)
    }
    return specularColor
}

// Formula from http://www.csee.umbc.edu/~olano/435f02/ray-sphere.html
func (sphere Sphere) hit(ray Ray) (float64, raytracer.Vector) {
    a := ray.direction.DotProduct(ray.direction) 
    b := 2.0 * ray.direction.DotProduct(eye.VectorSub(sphere.center)) 
    c := eye.VectorSub(sphere.center).DotProduct(eye.VectorSub(sphere.center)) - math.Pow(sphere.radius, 2)
    discriminant := math.Pow(b, 2) - 4.0*a*c

    if discriminant < 0 {
        return -1, raytracer.Vector{}
    }

    tNeg := (-b - math.Sqrt(discriminant))/(2*a)
    tPos := (-b + math.Sqrt(discriminant))/(2*a)
    var t float64
    t = math.Min(tNeg, tPos)

    intersection := getRayIntersection(t, ray)
    surfaceNormal := intersection.VectorSub(sphere.center).VectorDiv(sphere.radius)

    // Maybe an un-normalized tempLight. Does it make a difference?
    // AKA, is (vector1 - vector2) == (normalized(vector1) - normalized(vector2))?
    // incomingLight := vectorSub(intersection, tempLight)

    ambientColor := calculateAmbientColor()
    diffuseColor := calculateDiffuseColor(surfaceNormal)
    specularColor := calculateSpecularColor(intersection, surfaceNormal)

    //finalColor := specularColor
    //finalColor := vectorAdd(specularColor, diffuseColor)
    finalColor := ambientColor.VectorAdd(diffuseColor.VectorAdd(specularColor))

    return t, finalColor
}

func renderScene() {
    doneChannel := make(chan bool)
    pixelChannel := make(chan raytracer.Vector)
    go getPixelsRoutine(pixelChannel, doneChannel)

    for done := <- doneChannel; done == false; done = <- doneChannel{
        pixel := <- pixelChannel
        drawPixel(viewportColors, pixel.X, pixel.Y, 0, 0, 0) //This is probably extra work
        ray := computeRay(pixel)
        var color raytracer.Vector
        isHit := false
        minT := math.MaxFloat64
        for _, singleSphere := range allSpheres {
            rayHit, rayColor := singleSphere.hit(ray)
            if (rayHit != -1 && rayHit < minT) {
                color = rayColor
                isHit = true
                minT = rayHit
                //refactor this into clipping method later
                if color.X > 1.0 {
                    color.X = 1.0
                }
                if color.Y > 1.0 {
                    color.Y = 1.0
                }
                if color.Z > 1.0 {
                    color.Z = 1.0
                }
                if color.X < 0 {
                    color.X = 0
                }
                if color.Y < 0 {
                    color.Y = 0
                }
                if color.Z < 0 {
                    color.Z = 0
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
