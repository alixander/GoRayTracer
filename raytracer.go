package main

import (
    "log"
    "bufio"
    "fmt"
    "math"
    "strings"
    "strconv"
    "./vector"
    "os"
    "image"
    "image/color"
    "image/png"
    "time"
)

type Material struct {
    ambient raytracer.Vector
    diffuse raytracer.Vector
    specular raytracer.Vector
    shininess float64
    reflective raytracer.Vector
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
    LL = emptyVector()
    LR = emptyVector()
    UL = emptyVector()
    UR = emptyVector()
    width int = 0
    height int = 0
    viewport = image.Rect(0, 0, width, height)
    viewportColors = image.NewRGBA(viewport)

    eye = emptyVector()

    //sphereA = Sphere{center: raytracer.Vector{X: 0, Y:0, Z:-200}, radius: 320}
    //sphereA = Sphere{center: raytracer.Vector{X: 0, Y:0, Z:-400}, radius: 120}
    //sphereB = Sphere{center: raytracer.Vector{X: 250, Y:0, Z:-300}, radius: 120}
    //sphereC = Sphere{center: raytracer.Vector{X: -250, Y:0, Z:-700}, radius: 120}
    //sphereD = Sphere{center: raytracer.Vector{X: 0, Y:-250, Z:-200}, radius: 120}
    //sphereE = Sphere{center: raytracer.Vector{X: 0, Y:250, Z:-200}, radius: 120}

    pointLights = map[raytracer.Vector]raytracer.Vector{}
    directionalLights = map[raytracer.Vector]raytracer.Vector{}
    spheres = map[Sphere]Material{}
    transformations = [][]int{}
    ambientLight = emptyVector()

    //tempColor = raytracer.Vector{X: 0.6, Y: 0.6, Z: 0.6}
    //lightA = (raytracer.Vector{X: 300, Y: -400, Z: 350}).Normalize()
    //lightB = (raytracer.Vector{X: -100, Y: 400, Z: 100}).Normalize()


    //allSpheres = []Sphere {sphereA, sphereB, sphereC, sphereD, sphereE}
    //allSpheres = []Sphere {sphereA}
    //pointLights = []raytracer.Vector {lightA, lightB}
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

//P = u (vLL+ (1-v)UL)+(1-u)(vLR+ (1-v)UR)
func getP(u float64, v float64) raytracer.Vector {
    a := LL.VectorScale(v).VectorAdd(UL.VectorScale(1-v)).VectorScale(u)
    b := LR.VectorScale(v).VectorAdd(UR.VectorScale(1-v)).VectorScale(1-u)
    return a.VectorAdd(b)
}

func getPixelsRoutine(pixelChannel chan raytracer.Vector, doneChannel chan bool) {
    pixels := 1000.0
    //horizontalDistance := UL.DistanceTo(UR)
    //verticalDistance := UL.DistanceTo(LL)
    //horizontalStep := horizontalDistance/pixels
    //verticalStep := verticalDistance/pixels
    for u := 0.5; u < pixels; u ++ {
        for v := 0.5; v < pixels; v ++ {
            doneChannel <- false
            p := getP(u/pixels, v/pixels)
            pixelChannel <- p
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
func computeRay(start raytracer.Vector, pixel raytracer.Vector) Ray {
    return Ray{start: start, direction: pixel.VectorSub(start)}
}

func emptyVector() raytracer.Vector {
    return raytracer.Vector{X:0, Y:0, Z:0}
}

func calculateDiffuseColor(diffuse raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    diffuseColor := emptyVector()

    var theta float64
    var color raytracer.Vector

    for light, lightColor := range directionalLights {
        theta = math.Max(0, normal.DotProduct(light.Normalize()))
        color = diffuse.VectorScale(theta).VectorMult(lightColor)
        diffuseColor = diffuseColor.VectorAdd(color)
    }
    for light, lightColor := range pointLights {
        theta = math.Max(0, normal.DotProduct(light.Normalize()))
        color = diffuse.VectorScale(theta).VectorMult(lightColor)
        diffuseColor = diffuseColor.VectorAdd(color)
    }
    return diffuseColor
}

func calculateAmbientColor(ambient raytracer.Vector) raytracer.Vector {
    ambientColor := emptyVector()
    for _, lightColor := range directionalLights {
        ambientColor = ambientColor.VectorAdd(lightColor.VectorMult(ambient))
    }
    for _, lightColor := range pointLights {
        ambientColor = ambientColor.VectorAdd(lightColor.VectorMult(ambient))
    }
    ambientColor = ambientColor.VectorAdd(ambientLight)
    return ambientColor
}

// R = 2N(I . N) - I
func getreflectedLight(light raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    lightDotNormal := math.Max(0.0, light.DotProduct(normal))
    return normal.VectorScale(2.0*lightDotNormal).VectorSub(light)
}

func calculateSpecularColor(specular raytracer.Vector, shininess float64, intersection raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    specularColor := emptyVector()

    var reflectedLight raytracer.Vector
    var incomingLight raytracer.Vector
    var color raytracer.Vector
    var directionToViewer raytracer.Vector
    var specularTerm float64

    for light, lightColor := range directionalLights {
        incomingLight = light.VectorScale(-1)
        reflectedLight = getreflectedLight(incomingLight, normal).Normalize()
        directionToViewer = eye.VectorSub(intersection)
        specularTerm = math.Max(0, reflectedLight.DotProduct(directionToViewer.Normalize()))
        color = specular.VectorMult(lightColor.VectorScale(math.Pow(specularTerm, shininess)))
        specularColor = specularColor.VectorAdd(color)
    }
    for light, lightColor := range pointLights {
        incomingLight = light
        reflectedLight = getreflectedLight(incomingLight, normal).Normalize()
        directionToViewer = eye.VectorSub(intersection)
        specularTerm = math.Max(0, reflectedLight.DotProduct(directionToViewer.Normalize()))
        color = specular.VectorMult(lightColor.VectorScale(math.Pow(specularTerm, shininess)))
        specularColor = specularColor.VectorAdd(color)
    }
    return specularColor
}

func errorMargin(pointA raytracer.Vector, pointB raytracer.Vector) raytracer.Vector {
    //stepSize := pointA.DistanceTo(pointB)/40
    //fmt.Println(stepSize)
    stepSize := 50.0
    return raytracer.Vector{X:stepSize, Y:stepSize, Z:stepSize}
}

func (sphere Sphere) calculateColor(material Material, intersection raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    //ambientColor := calculateAmbientColor(material.ambient.VectorAdd(ambientLight))
    ambientColor := calculateAmbientColor(material.ambient)
    diffuseColor := calculateDiffuseColor(material.diffuse, normal)
    specularColor := calculateSpecularColor(material.specular, material.shininess, intersection, normal)

    shadedColor := ambientColor.VectorAdd(diffuseColor.VectorAdd(specularColor))

    isShadow := false
    for light, _ := range directionalLights {
        shadowRay := computeRay(intersection.VectorAdd(errorMargin(intersection, light)), light.VectorScale(-1))
        for otherSphere, _ := range spheres {
            if (otherSphere != sphere) {
                hitValue, _ := otherSphere.hit(shadowRay, true)
                if hitValue != -1 {
                    isShadow = true
                }
            }
        }
    }
    for light, _ := range pointLights {
        // Start a little bit after actual pixel to avoid errors with hitting itself
        shadowRay := computeRay(intersection.VectorAdd(errorMargin(intersection, light)), light)
        for otherSphere, _ := range spheres {
            if (otherSphere != sphere) {
                hitValue, _ := otherSphere.hit(shadowRay, true)
                if hitValue != -1 {
                    isShadow = true
                }
            }
        }
    }
    if isShadow {
        return ambientColor
    }

    return shadedColor
}

// Formula from http://www.csee.umbc.edu/~olano/435f02/ray-sphere.html
func (sphere Sphere) hit(ray Ray, isShadowRay bool) (float64, raytracer.Vector) {
    a := ray.direction.DotProduct(ray.direction) 
    b := 2.0 * ray.direction.DotProduct(ray.start.VectorSub(sphere.center)) 
    c := ray.start.VectorSub(sphere.center).DotProduct(ray.start.VectorSub(sphere.center)) - math.Pow(sphere.radius, 2)
    discriminant := math.Pow(b, 2) - 4.0*a*c

    if discriminant < 0 {
        return -1, emptyVector()
    }

    if isShadowRay {
        return 1, emptyVector()
    }

    tNeg := (-b - math.Sqrt(discriminant))/(2*a)
    tPos := (-b + math.Sqrt(discriminant))/(2*a)
    var t float64
    t = math.Min(tNeg, tPos)

    intersection := getRayIntersection(t, ray)
    surfaceNormal := intersection.VectorSub(sphere.center).VectorDiv(sphere.radius)

    color := sphere.calculateColor(spheres[sphere], intersection, surfaceNormal)

    return t, color
}

func clip(color *raytracer.Vector) {
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

func renderScene() {
    doneChannel := make(chan bool)
    pixelChannel := make(chan raytracer.Vector)
    go getPixelsRoutine(pixelChannel, doneChannel)

    for done := <- doneChannel; done == false; done = <- doneChannel{
        pixel := <- pixelChannel
        drawPixel(viewportColors, pixel.X+float64(width/2), -1*pixel.Y+float64(height/2), 0, 0, 0)
        ray := computeRay(eye, pixel)
        var color raytracer.Vector
        isHit := false
        minT := math.MaxFloat64
        for sphere, _ := range spheres {
            rayHit, rayColor := sphere.hit(ray, false)
            if (rayHit != -1 && rayHit < minT) {
                color = rayColor
                isHit = true
                minT = rayHit
                clip(&color)
            }
        }
        if (isHit) {
            drawPixel(viewportColors, pixel.X+float64(width/2), -1*pixel.Y+float64(height/2), color.X, color.Y, color.Z)
        }
    }
}

func updateIndices(currentIndex int, nextIndex int, line string) (int, int) {
    digits := "-0123456789"
    nextChar := strings.IndexAny(line[nextIndex:], digits)
    currentIndex = nextIndex + nextChar
    nextSpace := strings.Index(line[currentIndex:], " ")
    if nextSpace == -1 {
        nextIndex = len(line)
    } else {
        nextIndex = currentIndex + nextSpace
    }
    return currentIndex, nextIndex
}

func interpretScene(lines []string) {
    SCALE_FACTOR := 10.0
    var currentMaterial Material
    var currentIndex int
    var nextIndex int
    for _, line := range lines {
        currentIndex = 4
        nextIndex = currentIndex + strings.Index(line[currentIndex:], " ")
        // Comment lines in scene files
        if strings.Contains(line, "#") {
            continue
        }
        if strings.Contains(line, "cam") {
            camX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            camY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            camZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            eye.X, eye.Y, eye.Z = camX, camY, camZ
            eye = eye.VectorScale(SCALE_FACTOR)

            LLX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LLY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LLZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LL.X, LL.Y, LL.Z = LLX, LLY, LLZ

            LRX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LRY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LRZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            LR.X, LR.Y, LR.Z = LRX, LRY, LRZ

            ULX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ULY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ULZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            UL.X, UL.Y, UL.Z = ULX, ULY, ULZ

            URX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            URY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            URZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            UR.X, UR.Y, UR.Z = URX, URY, URZ

            LL = LL.VectorScale(SCALE_FACTOR)
            LR = LR.VectorScale(SCALE_FACTOR)
            UL = UL.VectorScale(SCALE_FACTOR)
            UR = UR.VectorScale(SCALE_FACTOR)

            width = int(math.Abs(LL.X - LR.X))
            height = int(math.Abs(LL.Y - UL.Y))
            //width = 1000
            //height = 1000
            viewport = image.Rect(0, 0, width, height)
            viewportColors = image.NewRGBA(viewport)
        } else if strings.Contains(line, "lta") {
            ambientR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ambientG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ambientB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            ambientLight.X, ambientLight.Y, ambientLight.Z = ambientR, ambientG, ambientB
        } else if strings.Contains(line, "ltp") {
            pointLight := emptyVector()
            lightX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            pointLight.X, pointLight.Y, pointLight.Z = lightX, lightY, lightZ
            lightColor := emptyVector()
            lightR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            lightColor.X, lightColor.Y, lightColor.Z = lightR, lightG, lightB
            pointLights[pointLight.VectorScale(SCALE_FACTOR)] = lightColor
        } else if strings.Contains(line, "ltd") {
            directionalLight := emptyVector()
            lightX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            directionalLight.X, directionalLight.Y, directionalLight.Z = lightX, lightY, lightZ
            lightColor := emptyVector()
            lightR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            lightB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            lightColor.X, lightColor.Y, lightColor.Z = lightR, lightG, lightB
            directionalLights[directionalLight.VectorScale(SCALE_FACTOR)] = lightColor 
        } else if strings.Contains(line, "mat") {
            ambient := emptyVector()
            ambientR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ambientG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ambientB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ambient.X, ambient.Y, ambient.Z = ambientR, ambientG, ambientB
            currentMaterial.ambient = ambient

            diffuse := emptyVector()
            diffuseR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            diffuseG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            diffuseB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            diffuse.X, diffuse.Y, diffuse.Z = diffuseR, diffuseG, diffuseB
            currentMaterial.diffuse = diffuse

            specular := emptyVector()
            specularR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            specularG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            specularB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            specular.X, specular.Y, specular.Z = specularR, specularG, specularB
            currentMaterial.specular = specular

            shininess, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentMaterial.shininess = shininess

            reflective := emptyVector()
            reflectiveR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            reflectiveG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            reflectiveB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            reflective.X, reflective.Y, reflective.Z = reflectiveR, reflectiveG, reflectiveB
            currentMaterial.reflective = reflective
        } else if strings.Contains(line, "sph") {
            centerX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            centerY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            centerZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            radius, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)

            sphere := Sphere{center: raytracer.Vector{X:centerX, Y:centerY, Z:centerZ}.VectorScale(SCALE_FACTOR), radius: radius*SCALE_FACTOR}
            spheres[sphere] = currentMaterial
        }
    }
}

func parseScene(filename string) {
    lines := []string{}
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
    interpretScene(lines)
}

func main() {
    fmt.Println("\n------------Starting--------------\n")
    startTime := time.Now()
    parseScene(os.Args[1])
    renderScene()
    saveScene(viewportColors)
    fmt.Println("Program finished running in", time.Since(startTime))
}
