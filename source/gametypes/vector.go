package gametypes

import "math"

type Vector2Int struct {
	X int
	Y int
}

// Add returns a new vector that is the sum of two vectors
func (v *Vector2Int) Add(other *Vector2Int) *Vector2Int {
	return &Vector2Int{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub returns a new vector that is the difference of two vectors
func (v *Vector2Int) Sub(other *Vector2Int) *Vector2Int {
	return &Vector2Int{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Multiply returns a new vector with coordinates multiplied by the given scalar
func (v *Vector2Int) Multiply(scalar int) *Vector2Int {
	return &Vector2Int{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

// Length returns the length of the vector (as a float64 for accuracy)
func (v *Vector2Int) Length() float64 {
	return math.Sqrt(float64(v.X*v.X + v.Y*v.Y))
}

// ManhattanDistance returns the Manhattan distance between two vectors
func (v *Vector2Int) ManhattanDistance(other *Vector2Int) int {
	return abs(v.X-other.X) + abs(v.Y-other.Y)
}

// Equal returns true if two vectors are equal
func (v *Vector2Int) Equal(other *Vector2Int) bool {
	return v.X == other.X && v.Y == other.Y
}

// Zero returns true if the vector is (0,0)
func (v *Vector2Int) Zero() bool {
	return v.X == 0 && v.Y == 0
}

// Clone returns a new vector with the same coordinates
func (v *Vector2Int) Clone() *Vector2Int {
	return &Vector2Int{
		X: v.X,
		Y: v.Y,
	}
}

// helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
