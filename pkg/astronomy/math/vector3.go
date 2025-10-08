package math

import "math"

type Vector3 struct {
    X, Y, Z float64
}

func (v Vector3) Add(other Vector3) Vector3 {
    return Vector3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vector3) Sub(other Vector3) Vector3 {
    return Vector3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vector3) Scale(s float64) Vector3 {
    return Vector3{v.X * s, v.Y * s, v.Z * s}
}

func (v Vector3) Magnitude() float64 {
    return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector3) Normalize() Vector3 {
    mag := v.Magnitude()
    if mag == 0 {
        return v
    }
    return v.Scale(1.0 / mag)
}

func (v Vector3) Dot(other Vector3) float64 {
    return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vector3) Cross(other Vector3) Vector3 {
    return Vector3{
        X: v.Y*other.Z - v.Z*other.Y,
        Y: v.Z*other.X - v.X*other.Z,
        Z: v.X*other.Y - v.Y*other.X,
    }
}
