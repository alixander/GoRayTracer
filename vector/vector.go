package raytracer

import (
    "math"
)

type Vector struct {
    X float64
    Y float64
    Z float64
}

func (a Vector) Equals(b Vector) bool {
    if a.X == b.X && a.Y == b.Y && a.Z == b.Z {
        return true
    }
    return false
}

func (a Vector) VectorScale(b float64) Vector {
    return Vector{
        X: a.X * b,
        Y: a.Y * b,
        Z: a.Z * b,
    }
}

func (a Vector) VectorAdd(b Vector) Vector {
    return Vector{
        X: a.X + b.X,
        Y: a.Y + b.Y,
        Z: a.Z + b.Z,
    }
}

func (a Vector) VectorSub(b Vector) Vector {
    return Vector{
        X: a.X - b.X,
        Y: a.Y - b.Y,
        Z: a.Z - b.Z,
    }
}

func (a Vector) VectorMult(b Vector) Vector {
    return Vector{
        X: a.X * b.X,
        Y: a.Y * b.Y,
        Z: a.Z * b.Z,
    }
}

func (a Vector) VectorDiv(b float64) Vector {
    return Vector{
        X: a.X/b,
        Y: a.Y/b,
        Z: a.Z/b,
    }
}

func (a Vector) VectorIncrement(b float64) Vector {
    return Vector{
        X: a.X+b,
        Y: a.Y+b,
        Z: a.Z+b,
    }
}

func (a Vector) Normalize() Vector {
    magnitude := math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z))
    return Vector{
        X: float64(a.X)/magnitude,
        Y: float64(a.Y)/magnitude,
        Z: float64(a.Z)/magnitude,
    }
}

// a x b = <a2*b3-a3*b2, a3*b1-a1*b3, a1*b2-a2*b1>
func (a Vector) CrossProduct(b Vector) Vector {
    x := a.Y*b.Z - a.Z*b.Y
    y := a.Z*b.X - a.X*b.Z
    z := a.X*b.Y - a.Y*b.X
    return Vector{X:x, Y:y, Z:z}
}

func (a Vector) DotProduct(b Vector) float64 {
    return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vector) DistanceTo(b Vector) float64 {
    return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2) + math.Pow(a.Z-b.Z, 2))
}
