package main

import (
    "log"
    "bufio"
    "fmt"
    "math"
    "math/rand"
    "reflect"
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

// T for Transform
type TMatrix struct {
    row0 [4]float64
    row1 [4]float64
    row2 [4]float64
    row3 [4]float64
}

type Ray struct {
    start raytracer.Vector
    direction raytracer.Vector
}

type Shape interface {
    hit(Ray, bool, int) (float64, raytracer.Vector)
}

type Triangle struct {
    a raytracer.Vector
    b raytracer.Vector
    c raytracer.Vector
}

type Sphere struct {
    id float64
    center raytracer.Vector
    radius float64
}

//globals
var (
    PIXELS = 1000.0
    IS_SHADOWED = 1.0
    SCALE_FACTOR = 10.0

    EMPTY = emptyMatrix()

    LL = emptyVector()
    LR = emptyVector()
    UL = emptyVector()
    UR = emptyVector()

    width int = 0
    height int = 0
    viewport = image.Rect(0, 0, width, height)
    viewportColors = image.NewRGBA(viewport)

    eye = emptyVector()

    pointLights = map[raytracer.Vector]raytracer.Vector{}
    directionalLights = map[raytracer.Vector]raytracer.Vector{}
    ambientLight = emptyVector()

    spheres = map[Sphere]Material{}
    triangles = map[Triangle]Material{}
    shapes = map[Shape]Material{}
    shapeTransformations = map[Shape]TMatrix{}
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
    //horizontalDistance := UL.DistanceTo(UR)
    //verticalDistance := UL.DistanceTo(LL)
    //horizontalStep := horizontalDistance/PIXELS
    //verticalStep := verticalDistance/PIXELS
    //fmt.Println(horizontalStep)
    index := 0
    for u := 0.5; u < PIXELS; u ++ {
        index += 1
        if index%10 == 0 {
            fmt.Println(u)
        }
        for v := 0.5; v < PIXELS; v++ {
            doneChannel <- false
            p := getP(u/PIXELS, v/PIXELS)
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

func emptyMatrix() TMatrix {
    row0 := [4]float64{0, 0, 0, 0}
    row1 := [4]float64{0, 0, 0, 0}
    row2 := [4]float64{0, 0, 0, 0}
    row3 := [4]float64{0, 0, 0, 0}

    matrix := TMatrix{row0:row0, row1:row1, row2:row2, row3:row3}
    return matrix
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
    ambientColor = ambientColor.VectorAdd(ambientLight.VectorMult(ambient))
    return ambientColor
}

// R = 2N(I . N) - I
func getReflectedLight(light raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    lightDotNormal := math.Max(0.0, light.DotProduct(normal))
    return normal.VectorScale(2.0*lightDotNormal).VectorSub(light)
}

func calculateSpecularColor(specular raytracer.Vector, shininess float64, intersection raytracer.Vector, normal raytracer.Vector, ray Ray, isReflection bool) raytracer.Vector {
    specularColor := emptyVector()

    var reflectedLight raytracer.Vector
    var incomingLight raytracer.Vector
    var color raytracer.Vector
    var directionToViewer raytracer.Vector
    var specularTerm float64

    for light, lightColor := range directionalLights {
        incomingLight = light.VectorScale(-1)
        reflectedLight = getReflectedLight(incomingLight, normal).Normalize()
        directionToViewer = ray.start.VectorSub(intersection)
        specularTerm = math.Max(0, reflectedLight.DotProduct(directionToViewer.Normalize()))
        color = specular.VectorMult(lightColor.VectorScale(math.Pow(specularTerm, shininess)))
        specularColor = specularColor.VectorAdd(color)
    }
    for light, lightColor := range pointLights {
        incomingLight = light
        reflectedLight = getReflectedLight(incomingLight, normal).Normalize()
        directionToViewer = ray.start.VectorSub(intersection)
        if isReflection {
            specularTerm = math.Min(0, reflectedLight.DotProduct(directionToViewer.Normalize()))
        } else {
            specularTerm = math.Max(0, reflectedLight.DotProduct(directionToViewer.Normalize()))
        }
        color = specular.VectorMult(lightColor.VectorScale(math.Pow(specularTerm, shininess)))
        specularColor = specularColor.VectorAdd(color)
    }
    return specularColor
}

func calculateColor(shape Shape, material Material, intersection raytracer.Vector, normal raytracer.Vector, ray Ray, isReflection bool) raytracer.Vector {
    //ambientColor := calculateAmbientColor(material.ambient.VectorAdd(ambientLight))
    ambientColor := calculateAmbientColor(material.ambient)
    diffuseColor := calculateDiffuseColor(material.diffuse, normal)
    specularColor := calculateSpecularColor(material.specular, material.shininess, intersection, normal, ray, isReflection)

    shadedColor := ambientColor.VectorAdd(diffuseColor.VectorAdd(specularColor))
    //if isReflection {
    //    shadedColor = specularColor
    //}

    isShadowed := false
    for light, _ := range directionalLights {
        shadowRay := computeRay(intersection, intersection.VectorAdd(light.VectorScale(-1)))
        for otherShape, _ := range shapes {
            if (!reflect.DeepEqual(otherShape, shape)) {
                hitValue, _ := otherShape.hit(shadowRay, true, 1)
                if hitValue == IS_SHADOWED {
                    return ambientColor
                    isShadowed = true
                    shadedColor = ambientColor
                    break
                }
            }
        }
        // No more need to go on once you find out it's already shadowed
        if isShadowed {
            break
        }
    }
    if !isShadowed {
        for light, _ := range pointLights {
            shadowRay := computeRay(intersection, light)
            for otherShape, _ := range shapes {
                if (!reflect.DeepEqual(otherShape, shape)) {
                    hitValue, _ := otherShape.hit(shadowRay, true, 1)
                    if hitValue == IS_SHADOWED {
                        return ambientColor
                        isShadowed = true
                        shadedColor = ambientColor
                        break
                    }
                }
            }
            if isShadowed {
                break
            }
        }
    }

    return shadedColor
}

func reflectionLight(incoming raytracer.Vector, normal raytracer.Vector) raytracer.Vector {
    d := math.Min(0, incoming.DotProduct(normal))
    return incoming.VectorAdd(normal.VectorScale(2*d))
}

func calculateReflectedColor(shape Shape, incomingRay Ray, intersection raytracer.Vector, normal raytracer.Vector, depth int) raytracer.Vector {
    reflectedColor := emptyVector()
    minT := -math.MaxFloat64
    //incomingLight := intersection.VectorSub(incomingRay.start)
    //fmt.Println(incomingLight)
    //reflectedLight := getReflectedLight(incomingRay.direction.VectorScale(-1), normal)
    reflectedLight := reflectionLight(incomingRay.direction, normal)
    //reflectedLight := reflectionLight(incomingLight, normal).Normalize()
    //outgoingLight := reflectedLight.VectorSub(intersection)
    //reflectedRay := computeRay(intersection, intersection.VectorSub(reflectedLight))
    reflectedRay := computeRay(intersection, reflectedLight)
    //reflectedRay := computeRay(intersection, outgoingLight)
    for otherShape, _ := range shapes {
        if (!reflect.DeepEqual(otherShape, shape)) {
            hitValue, color := otherShape.hit(reflectedRay, false, depth)
            if (hitValue < 0 && hitValue != -1 && hitValue > minT) {
                //fmt.Println(hitValue)
                reflectedColor = color
                minT = hitValue
                //clip(&color)
            }
        }
    }
    return reflectedColor
}

//func traceBack(shape Shape, intersection raytracer.Vector, normal raytracer.Vector, depth int) raytracer.Vector {
//    for light, _ := range pointLights {
//        incomingLight := light
//        reflected := reflectionLight(incomingLight, normal)
//        outgoingLight := reflected.VectorSub(intersection)
//        reflectedRay := computeRay(incomingLight, intersection)
//
//    }
//}

func isInsideTriangle(triangle Triangle, intersection raytracer.Vector, normal raytracer.Vector) bool {
    edge0 := triangle.b.VectorSub(triangle.a)
    c0 := intersection.VectorSub(triangle.a)
    if (normal.DotProduct(edge0.CrossProduct(c0))) < 0 {
        return false
    }
    edge1 := triangle.c.VectorSub(triangle.b)
    c1 := intersection.VectorSub(triangle.b)
    if (normal.DotProduct(edge1.CrossProduct(c1))) < 0 {
        return false
    }

    edge2 := triangle.a.VectorSub(triangle.c)
    c2 := intersection.VectorSub(triangle.c)
    if (normal.DotProduct(edge2.CrossProduct(c2))) < 0 {
        return false
    }
    return true
}

// http://www.scratchapixel.com/lessons/3d-basic-lessons/lesson-9-ray-triangle-intersection/ray-triangle-intersection-geometric-solution/
func (triangle Triangle) hit(ray Ray, isShadowRay bool, reflectionDepth int) (float64, raytracer.Vector) {
    tMatrix := shapeTransformations[triangle]
    usedRay := ray
    if tMatrix != EMPTY {
        //fmt.Println("Transforming triangle by:", tMatrix)
        usedRay.start = applyT(tMatrix, ray.start, true)
        usedRay.direction = applyT(tMatrix, ray.direction, false)
    }
    // n = (V1 - V0) x (V2 - V0)
    surfaceNormal := triangle.b.VectorSub(triangle.a).CrossProduct(triangle.c.VectorSub(triangle.a)).Normalize()
    d := surfaceNormal.DotProduct(triangle.a.Normalize())
    t := -(surfaceNormal.DotProduct(usedRay.start) + d)/(surfaceNormal.DotProduct(usedRay.direction))
    intersection := getRayIntersection(t, usedRay)
    //fmt.Println("Intersection:", intersection)
    originalIntersection := intersection
    if tMatrix != EMPTY {
        intersection = applyT(tMatrix, intersection, true)
    }

    // Still need to handle case of triangle parallel to eye
    if !isInsideTriangle(triangle, originalIntersection, surfaceNormal) {
        return -1, emptyVector()
    }
    if isShadowRay {
        if t > 0 {
            return IS_SHADOWED, emptyVector()
        } else {
            return -1, emptyVector()
        }
    }

    color := calculateColor(triangle, triangles[triangle], originalIntersection, surfaceNormal, usedRay, false)
    if reflectionDepth == 0 {
        color = calculateColor(triangle, triangles[triangle], originalIntersection, surfaceNormal, usedRay, true)
    }
    if reflectionDepth > 0 {
        reflectedColor := calculateReflectedColor(triangle, usedRay, originalIntersection, surfaceNormal, reflectionDepth-1)
        empty := emptyVector()
        if reflectedColor != empty {
            color = color.VectorAdd(reflectedColor.VectorMult(triangles[triangle].reflective))
            //color = color.VectorScale(0.3).VectorAdd(reflectedColor.VectorScale(0.7))
            //color = reflectedColor
        }
    }

    return t, color
}

func transformNormal(matrix TMatrix, normal raytracer.Vector) raytracer.Vector {
    row0 := []float64{matrix.row1[1]*matrix.row2[2] - matrix.row1[2]*matrix.row2[1],
                      matrix.row1[2]*matrix.row2[0] - matrix.row1[0]*matrix.row2[2],
                      matrix.row1[0]*matrix.row2[1] - matrix.row1[1]*matrix.row2[0]}

    row1 := []float64{matrix.row0[2]*matrix.row2[1] - matrix.row0[1]*matrix.row2[2],
                      matrix.row0[0]*matrix.row2[2] - matrix.row0[2]*matrix.row2[0],
                      matrix.row0[1]*matrix.row2[0] - matrix.row0[0]*matrix.row2[1]}

    row2 := []float64{matrix.row0[1]*matrix.row1[2] - matrix.row0[2]*matrix.row1[1],
                      matrix.row0[2]*matrix.row1[0] - matrix.row0[0]*matrix.row1[2],
                      matrix.row0[0]*matrix.row1[1] - matrix.row0[1]*matrix.row1[0]}
    
    product0 := normal.X * row0[0] + normal.Y * row0[1] + normal.Z * row0[2]
    product1 := normal.X * row1[0] + normal.Y * row1[1] + normal.Z * row1[2]
    product2 := normal.X * row2[0] + normal.Y * row2[1] + normal.Z * row2[2]
    return raytracer.Vector{product0, product1, product2}
}

func applyT(matrix TMatrix, v raytracer.Vector, isPoint bool) raytracer.Vector {
    if isPoint {
        product0 := v.X * matrix.row0[0] + v.Y * matrix.row0[1] + v.Z * matrix.row0[2] + 1 * matrix.row0[3]
        product1 := v.X * matrix.row1[0] + v.Y * matrix.row1[1] + v.Z * matrix.row1[2] + 1 * matrix.row1[3]
        product2 := v.X * matrix.row2[0] + v.Y * matrix.row2[1] + v.Z * matrix.row2[2] + 1 * matrix.row2[3]
        //product3 := v.X * matrix.row3[0] + v.Y * matrix.row3[1] + v.Z * matrix.row3[2] + 1 * matrix.row3[3]
        return raytracer.Vector{product0, product1, product2}
    } else {
        product0 := v.X * matrix.row0[0] + v.Y * matrix.row0[1] + v.Z * matrix.row0[2]
        product1 := v.X * matrix.row1[0] + v.Y * matrix.row1[1] + v.Z * matrix.row1[2]
        product2 := v.X * matrix.row2[0] + v.Y * matrix.row2[1] + v.Z * matrix.row2[2]
        //product3 := v.X * matrix.row3[0] + v.Y * matrix.row3[1] + v.Z * matrix.row3[2] + 0 * matrix.row3[3]
        return raytracer.Vector{product0, product1, product2}
    }
}

// Formula from http://www.csee.umbc.edu/~olano/435f02/ray-sphere.html
func (sphere Sphere) hit(ray Ray, isShadowRay bool, reflectionDepth int) (float64, raytracer.Vector) {
    tMatrix := shapeTransformations[sphere]
    usedRay := ray
    if tMatrix != EMPTY {
        usedRay.start = applyT(tMatrix, ray.start, true)
        usedRay.direction = applyT(tMatrix, ray.direction, false)
    }
    a := usedRay.direction.DotProduct(usedRay.direction) 
    b := 2.0 * usedRay.direction.DotProduct(usedRay.start.VectorSub(sphere.center)) 
    c := usedRay.start.VectorSub(sphere.center).DotProduct(usedRay.start.VectorSub(sphere.center)) - math.Pow(sphere.radius, 2)
    discriminant := math.Pow(b, 2) - 4.0*a*c

    if discriminant < 0 {
        return -1, emptyVector()
    }

    tNeg := (-b - math.Sqrt(discriminant))/(2*a)
    tPos := (-b + math.Sqrt(discriminant))/(2*a)
    var t float64
    t = math.Min(tNeg, tPos)
    if isShadowRay {
        if t > 0 {
            return IS_SHADOWED, emptyVector()
        } else {
            return -1, emptyVector()
        }
    }

    intersection := getRayIntersection(t, usedRay)
    originalIntersection := intersection
    if tMatrix != EMPTY {
        intersection = applyT(tMatrix, intersection, true)
    }
    surfaceNormal := intersection.VectorSub(sphere.center).VectorDiv(sphere.radius)

    color := calculateColor(sphere, spheres[sphere], originalIntersection, surfaceNormal, usedRay, false)
    if reflectionDepth == 0 {
        color = calculateColor(sphere, spheres[sphere], originalIntersection, surfaceNormal, usedRay, true)
    }
    if reflectionDepth > 0 {
        reflectedColor := calculateReflectedColor(sphere, usedRay, originalIntersection, surfaceNormal, reflectionDepth-1)
        empty := emptyVector()
        if reflectedColor != empty {
            color = color.VectorAdd(reflectedColor.VectorMult(spheres[sphere].reflective))
            //color = color.VectorScale(0.3).VectorAdd(reflectedColor.VectorScale(0.7))
            //color = reflectedColor
        }
    }

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
        // Refactor these into one for loop with Shape interface
        for shape, _ := range shapes {
            rayHit, rayColor := shape.hit(ray, false, 3)
            if (rayHit != -1 && rayHit < minT) {
                color = rayColor
                isHit = true
                minT = rayHit
            }
        }
        if (isHit) {
            clip(&color)
            drawPixel(viewportColors, pixel.X+float64(width/2), -1*pixel.Y+float64(height/2), color.X, color.Y, color.Z)
        }
    }
}

func rowTimesColumn(row [4]float64, column [4]float64) float64{
    return column[0]*row[0] + column[1]*row[1] + column[2]*row[2] + column[3]*row[3]
}

func matrixMultiply(a TMatrix, b TMatrix) TMatrix {
    column0 := [4]float64{b.row0[0], b.row1[0], b.row2[0], b.row3[0]}
    column1 := [4]float64{b.row0[1], b.row1[1], b.row2[1], b.row3[1]}
    column2 := [4]float64{b.row0[2], b.row1[2], b.row2[2], b.row3[2]}
    column3 := [4]float64{b.row0[3], b.row1[3], b.row2[3], b.row3[3]}

    row0 := [4]float64{rowTimesColumn(a.row0, column0), rowTimesColumn(a.row0, column1), rowTimesColumn(a.row0, column2), rowTimesColumn(a.row0, column3)} 
    row1 := [4]float64{rowTimesColumn(a.row1, column0), rowTimesColumn(a.row1, column1), rowTimesColumn(a.row1, column2), rowTimesColumn(a.row1, column3)} 
    row2 := [4]float64{rowTimesColumn(a.row2, column0), rowTimesColumn(a.row2, column1), rowTimesColumn(a.row2, column2), rowTimesColumn(a.row2, column3)} 
    row3 := [4]float64{rowTimesColumn(a.row3, column0), rowTimesColumn(a.row3, column1), rowTimesColumn(a.row3, column2), rowTimesColumn(a.row3, column3)} 
    
    return TMatrix{row0:row0, row1:row1, row2:row2, row3:row3}
}

func updateIndices(currentIndex int, nextIndex int, line string) (int, int) {
    digits := "-0123456789eE+"
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
    var currentMaterial Material
    var currentTransformation TMatrix
    var currentIndex int
    var nextIndex int
    for _, line := range lines {
        currentIndex = int(math.Min(4, float64(len(line))))
        next := strings.Index(line[currentIndex:], " ")
        if next == -1 {
            nextIndex = len(line)
        } else {
            nextIndex = currentIndex + next
        }
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

            //width = int(math.Abs(LL.X - LR.X))
            //height = int(math.Abs(LL.Y - UL.Y))
            width = int(PIXELS)
            height = int(PIXELS)
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
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            currentMaterial.shininess = shininess

            reflective := emptyVector()
            reflectiveR, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            reflectiveG, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            reflectiveB, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            reflective.X, reflective.Y, reflective.Z = reflectiveR, reflectiveG, reflectiveB
            currentMaterial.reflective = reflective
        } else if strings.Contains(line, "xft") {
            tx, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            ty, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            tz, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            
            row0 := [4]float64{1, 0, 0, -tx*SCALE_FACTOR}
            row1 := [4]float64{0, 1, 0, -ty*SCALE_FACTOR}
            row2 := [4]float64{0, 0, 1, -tz*SCALE_FACTOR}
            row3 := [4]float64{0, 0, 0, 1}

            matrix := TMatrix{row0:row0, row1:row1, row2:row2, row3:row3}
            if currentTransformation != EMPTY {
                // New matrix multiplied on the right side
                fmt.Println("Before:", currentTransformation)
                fmt.Println("Matrix:", matrix)
                currentTransformation = matrixMultiply(currentTransformation, matrix)
                fmt.Println("After:",currentTransformation)
            } else {
                currentTransformation = matrix
            }
        } else if strings.Contains(line, "xfs") {
            //sx, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            //currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            //sy, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            //currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            //sz, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            //
            //row0 := [4]float64{1.0/(sx*SCALE_FACTOR), 0, 0, 0}
            //row1 := [4]float64{0, 1.0/(sy*SCALE_FACTOR), 0, 0}
            //row2 := [4]float64{0, 0, 1.0/(sz*SCALE_FACTOR), 0}
            //row3 := [4]float64{0, 0, 0, 1}

            //matrix := TMatrix{row0:row0, row1:row1, row2:row2, row3:row3}
            //if currentTransformation != emptyMatrix() {
            //    // New matrix multiplied on the right side
            //    currentTransformation = matrixMultiply(currentTransformation, matrix)
            //    //currentTransformation = matrixMultiply(matrix, currentTransformation)
            //} else {
            //    currentTransformation = matrix
            //}
        } else if strings.Contains(line, "xfr") {

        } else if strings.Contains(line, "xfz") {
            currentTransformation = EMPTY
        } else if strings.Contains(line, "obj") {
            parseObj(line[currentIndex:nextIndex], currentTransformation, currentMaterial)
        }else if strings.Contains(line, "sph") {
            centerX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            centerY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            centerZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            radius, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)

            sphere := Sphere{id: rand.Float64(), center: raytracer.Vector{X:centerX, Y:centerY, Z:centerZ}.VectorScale(SCALE_FACTOR), radius: radius*SCALE_FACTOR}
            spheres[sphere] = currentMaterial
            shapes[Shape(sphere)] = currentMaterial
            shapeTransformations[Shape(sphere)] = currentTransformation
        } else if strings.Contains(line, "tri") {
            aX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            aY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            aZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            a := raytracer.Vector{X:aX, Y:aY, Z:aZ}.VectorScale(SCALE_FACTOR)

            bX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            bY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            bZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            b := raytracer.Vector{X:bX, Y:bY, Z:bZ}.VectorScale(SCALE_FACTOR)

            cX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            cY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            cZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            c := raytracer.Vector{X:cX, Y:cY, Z:cZ}.VectorScale(SCALE_FACTOR)

            triangle := Triangle{a:a, b:b, c:c}
            triangles[triangle] = currentMaterial
            shapes[Shape(triangle)] = currentMaterial
            shapeTransformations[Shape(triangle)] = currentTransformation
        }
    }
}

func interpretObj(lines []string, transformation TMatrix, material Material) {
    var vertices []raytracer.Vector = make([]raytracer.Vector, 5000)
    var vertexIndex int = 0
    var currentIndex int
    var nextIndex int
    for _, line := range lines {
        if line == "" {
            continue
        }
        //digits := "-0123456789eE+"
        //nextChar := strings.IndexAny(line[1:], digits)
        currentIndex = int(math.Min(2, float64(len(line))))
        //currentIndex = int(nextChar)
        nextIndex = currentIndex + strings.Index(line[currentIndex:], " ")
        // Comment lines in scene files
        if strings.Contains(line, "#") {
            continue
        } else if strings.Contains(line, "v") {
            vertexX, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            vertexY, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            vertexZ, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            vertex := raytracer.Vector{vertexX, vertexY, vertexZ}.VectorScale(SCALE_FACTOR)
            vertices[vertexIndex] = vertex
            vertexIndex += 1
        } else if strings.Contains(line, "f") {
            index0, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            index1, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            currentIndex, nextIndex = updateIndices(currentIndex, nextIndex, line)
            index2, _ := strconv.ParseFloat(line[currentIndex:nextIndex], 64)
            //fmt.Println(index0, index1, index2)
            triangle := Triangle{vertices[int(index0)-1], vertices[int(index1)-1], vertices[int(index2)-1]} 
            triangles[triangle] = material
            shapes[Shape(triangle)] = material
            shapeTransformations[Shape(triangle)] = transformation
        }
    }
}

func parseObj(filename string, transformation TMatrix, material Material) {
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
    interpretObj(lines, transformation, material)
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
