//
//  goray/bound.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package goray

import "fmath"

type Bound struct {
    a, g Vector3D
}

func NewBound(a, g Vector3D) Bound {
    return Bound{a, g}
}

func BoundUnion(b1, b2 Bound) Bound {
    newBound := Bound{}
    newBound.a.X = fmath.Min(b1.a.X, b2.a.X)
    newBound.a.Y = fmath.Min(b1.a.Y, b2.a.Y)
    newBound.a.Z = fmath.Min(b1.a.Z, b2.a.Z)
    newBound.g.X = fmath.Max(b1.g.X, b2.g.X)
    newBound.g.Y = fmath.Max(b1.g.Y, b2.g.Y)
    newBound.g.Z = fmath.Max(b1.g.Z, b2.g.Z)
    return newBound
}

func (b Bound) Get() (a, g Vector3D) { return b.a, b.g }
func (b *Bound) Set(a, g Vector3D) { b.a = a; b.g = g }

func (b Bound) Cross(from, ray Vector3D) bool {
    // TODO
    return true
}

func (b Bound) GetVolume() float {
    return (b.g.Y - b.a.Y) * (b.g.X - b.a.X) * (b.g.Z - b.a.Z)
}

func (b Bound) GetXLength() float { return b.g.X - b.a.X }
func (b Bound) GetYLength() float { return b.g.Y - b.a.Y }
func (b Bound) GetZLength() float { return b.g.Z - b.a.Z }

func (b *Bound) Include(p Vector3D) {
    b.a.X = fmath.Min(b.a.X, p.X)
    b.a.Y = fmath.Min(b.a.Y, p.Y)
    b.a.Z = fmath.Min(b.a.Z, p.Z)
    b.g.X = fmath.Max(b.g.X, p.X)
    b.g.Y = fmath.Max(b.g.Y, p.Y)
    b.g.Z = fmath.Max(b.g.Z, p.Z)
}

func (b Bound) Includes(p Vector3D) bool {
    return (p.X >= b.a.X && p.X <= b.g.X &&
            p.Y >= b.a.Y && p.Y <= b.g.Y &&
            p.Z >= b.a.Z && p.Z <= b.g.Z)
}

func (b Bound) GetCenter() Vector3D {
    return VectorScalarMul(VectorAdd(b.g, b.a), 0.5)
}

func (b Bound) GetCenterX() { return (b.g.X + b.a.X) * 0.5 }
func (b Bound) GetCenterY() { return (b.g.Y + b.a.Y) * 0.5 }
func (b Bound) GetCenterZ() { return (b.g.Z + b.a.Z) * 0.5 }
