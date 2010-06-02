//
//  goray/partition.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

package partition

import (
	"./goray/bound"
	"./goray/color"
	"./goray/primitive"
	"./goray/render"
	"./goray/ray"
)

type Partitioner interface {
	Intersect(r ray.Ray, dist float) (hit bool, prim primitive.Primitive, z float)
	IntersectS(r ray.Ray, dist float) (hit bool, prim primitive.Primitive)
	IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (hit bool, prim primitive.Primitive)
	GetBound() *bound.Bound
}
