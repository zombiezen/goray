//
//  goray/version.go
//  goray
//
//  Created by Ross Light on 2010-06-04.
//

package version

import "fmt"

const Major = 0
const Minor = 1
const Patch = 0

func GetString() string { return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch) }
