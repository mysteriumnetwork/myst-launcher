package main

import (
	"fmt"
	"github.com/mysteriumnetwork/myst-launcher/myst"
)

func main() {
	m, err := myst.NewManagerWithDefaults()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(m)
}
