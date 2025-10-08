package math

import "math"

// Vector3 represents a 3D vector for astronomical calculations
type Vector3 struct {
    X, Y, Z float64
}

// Add returns the sum of two vectors
func (v Vector3) Add(other Vector3) Vector3 {
    return Vector3{
        X: v.X + other.X,
        Y: v.Y + other.Y,
        Z: v.Z + other.Z,
    }
}

// Sub returns the difference between two vectors
func (v Vector3) Sub(other Vector3) Vector3 {
    return Vector3{
        X: v.X - other.X,
        Y: v.Y - other.Y,
        Z: v.Z - other.Z,
    }
}

// Scale returns the vector scaled by a scalar
func (v Vector3) Scale(s float64) Vector3 {
    return Vector3{
        X: v.X * s,
        Y: v.Y * s,
        Z: v.Z * s,
    }
}

// Dot returns the dot product of two vectors
func (v Vector3) Dot(other Vector3) float64 {
    return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product of two vectors
func (v Vector3) Cross(other Vector3) Vector3 {
    return Vector3{
        X: v.Y*other.Z - v.Z*other.Y,
        Y: v.Z*other.X - v.X*other.Z,
        Z: v.X*other.Y - v.Y*other.X,
    }
}

// Magnitude returns the length of the vector
func (v Vector3) Magnitude() float64 {
    return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normalize returns a unit vector in the same direction
func (v Vector3) Normalize() Vector3 {
    mag := v.Magnitude()
    if mag == 0 {
        return v
    }
    return v.Scale(1.0 / mag)
}

// Distance returns the distance between two vectors
func (v Vector3) Distance(other Vector3) float64 {
    return v.Sub(other).Magnitude()
}

// IsZero checks if the vector is zero
func (v Vector3) IsZero() bool {
    return v.X == 0 && v.Y == 0 && v.Z == 0
}
