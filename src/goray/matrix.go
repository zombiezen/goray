//
//  goray/matrix.go
//  goray
//
//  Created by Ross Light on 2010-05-27.
//

/* The goray/matrix package gives a type for representing and manipulating a 4x4 transformation matrix. */
package matrix

import (
	"fmt"
	"math"
	"fmath"
	"goray/vector"
)

const dim = 4

/* Matrix holds a 4x4 transformation matrix. */
type Matrix struct {
	m [][]float
}

/* New creates a new matrix, filled with the value given. */
func New(fill float) *Matrix {
	m := new(Matrix)
	data := make([]float, dim*dim)
	m.m = [][]float{data[0:4], data[4:8], data[8:12], data[12:16]}
	m.Init(fill)
	return m
}

/* Identity creates a new identity matrix. */
func Identity() *Matrix {
	i := New(0.0)
	i.m[0][0] = 1.0
	i.m[1][1] = 1.0
	i.m[2][2] = 1.0
	i.m[3][3] = 1.0
	return i
}

/* Init sets all of the values of a matrix to a given value. */
func (m *Matrix) Init(fill float) {
	for i := 0; i < dim*dim; i++ {
		(m.m[0][0 : dim*dim])[i] = fill
	}
}

/* Get retrieves a value inside the matrix. */
func (m Matrix) Get(row, col int) float {
	return m.m[row][col]
}

/* GetAll retrieves a copy of the values inside the matrix. */
func (m Matrix) GetAll() [][]float {
	// We do a copy here to prevent the client from rearranging the internal
	// memory layout.
	a := make([][]float, 4)
	copy(a, m.m)
	return a
}

/* Set changes a single value inside the matrix. */
func (m *Matrix) Set(row, col int, val float) {
	m.m[row][col] = val
}

/* SetAll changes all of the values inside the matrix simultaneously. */
func (m *Matrix) SetAll(data [][]float) {
	for i, row := range m.m {
		copy(row, data[i])
	}
}

func (m Matrix) String() (result string) {
	for i, row := range m.m {
		format := "| %5.2f %5.2f %5.2f %5.2f |\n"
		switch i {
		case 0:
			format = "/ %5.2f %5.2f %5.2f %5.2f \\\n"
		case dim - 1:
			format = "\\ %5.2f %5.2f %5.2f %5.2f /\n"
		}
		result += fmt.Sprintf(format, row[0], row[1], row[2], row[3])
	}
	return
}

/* Duplicate creates a new matrix with identical values. */
func (m Matrix) Duplicate() *Matrix {
	dup := New(0.0)
	copy(dup.m[0:dim*dim], m.m[0:dim*dim])
	return dup
}

/* Inverse finds the inverse of the matrix in-place and returns whether it was successful. */
func (m *Matrix) Inverse() bool {
	iden := New(1.0)

	for i := 0; i < dim; i++ {
		max := 0.0
		ci := 0

		for k := i; k < dim; k++ {
			if fmath.Abs(m.m[k][i]) > max {
				max = fmath.Abs(m.m[k][i])
				ci = k
			}
		}

		if max == 0 {
			// Matrix has no inverse
			return false
		}

		swap := func(mat *Matrix) {
			for j := 0; j < dim; j++ {
				mat.m[i][j] = mat.m[ci][j]
			}
		}
		swap(m)
		swap(iden)

		factor := m.m[i][i]
		div := func(mat *Matrix) {
			for j := 0; j < dim; j++ {
				mat.m[i][j] /= factor
			}
		}
		div(m)
		div(iden)

		for k := 0; k < dim; k++ {
			if k != i {
				factor = m.m[k][i]
				res := func(mat *Matrix) {
					for j := 0; j < dim; j++ {
						mat.m[i][j] -= mat.m[k][j] * factor
					}
				}
				res(m)
				res(iden)
			}
		}
	}

	m.SetAll(iden.m)
	return true
}

/* Transpose performs an in-place matrix transposition. */
func (m *Matrix) Transpose() {
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 4; j++ {
			m.m[i][j], m.m[j][i] = m.m[j][i], m.m[i][j]
		}
	}
}

func (m *Matrix) Translate(x, y, z float) {
	transform := New(1.0)
	transform.m[0][3] = x
	transform.m[1][3] = y
	transform.m[2][3] = z
	m.Mul(transform)
}

func normDeg(degrees float) (angle float) {
	angle = fmath.Mod(degrees, 360.0)
	if angle < 0 {
		angle = 360.0 - angle
	}
	angle *= math.Pi / 180.0
	return
}

func (m *Matrix) RotateX(degrees float) {
	angle := normDeg(degrees)
	t := New(1.0)
	t.m[1][1] = fmath.Cos(angle)
	t.m[1][2] = -fmath.Sin(angle)
	t.m[2][1] = fmath.Sin(angle)
	t.m[2][2] = fmath.Cos(angle)
	m.Mul(t)
}

func (m *Matrix) RotateY(degrees float) {
	angle := normDeg(degrees)
	t := New(1.0)
	t.m[0][0] = fmath.Cos(angle)
	t.m[0][2] = fmath.Sin(angle)
	t.m[2][0] = -fmath.Sin(angle)
	t.m[2][2] = fmath.Cos(angle)
	m.Mul(t)
}

func (m *Matrix) RotateZ(degrees float) {
	angle := normDeg(degrees)
	t := New(1.0)
	t.m[0][0] = fmath.Cos(angle)
	t.m[0][1] = -fmath.Sin(angle)
	t.m[1][0] = fmath.Sin(angle)
	t.m[1][1] = fmath.Cos(angle)
	m.Mul(t)
}

func (m *Matrix) Scale(x, y, z float) {
	m.m[0][0] *= x
	m.m[1][0] *= x
	m.m[2][0] *= x

	m.m[0][1] *= y
	m.m[1][1] *= y
	m.m[2][1] *= y

	m.m[0][2] *= z
	m.m[1][2] *= z
	m.m[2][2] *= z
}

func (m1 *Matrix) Mul(m2 *Matrix) {
	m1.SetAll(Mul(m1, m2).m)
}

func Mul(m1, m2 *Matrix) (result *Matrix) {
	result = New(0.0)

	for i := 0; i < dim; i++ {
		for k := 0; k < dim; k++ {
			for j := 0; j < dim; j++ {
				result.m[i][k] += m1.m[i][j] * m2.m[j][k]
			}
		}
	}
	return
}

/* VecMul transforms a vector by a transformation matrix. */
func VecMul(m *Matrix, v vector.Vector3D) vector.Vector3D {
	return vector.Vector3D{
		m.m[0][0]*v.X + m.m[0][1]*v.Y + m.m[0][2]*v.Z,
		m.m[1][0]*v.X + m.m[1][1]*v.Y + m.m[1][2]*v.Z,
		m.m[2][0]*v.X + m.m[2][1]*v.Y + m.m[2][2]*v.Z,
	}
}
