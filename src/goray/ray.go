//
//  goray/ray.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package goray

type Ray struct {
    From, Dir Vector3D
    TMin, TMax, Time float
}
