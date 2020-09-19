//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 19.09.2020
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package main

import (
	"fmt"

	"github.com/lithammer/shortuuid"
)

func main() {
	u := shortuuid.New()
	fmt.Printf("%s\n", u)
}
