//
//  goray/ray.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package ray

import "./goray/vector"

type Ray struct {
    From, Dir vector.Vector3D
    TMin, TMax, Time float
}
