package main

import (

	"strconv"
	"fmt"
	"time"

	"github.com/kuredoro/cptest"
)

const defaultTL = 6 * time.Second

func GetTL(inputs cptest.Inputs) (TL time.Duration) {
    TL = defaultTL

    if str, exists := inputs.Config["tl"]; exists {
        sec, err := strconv.ParseFloat(str, 64)
        if err != nil {
            fmt.Printf("warning: time limit %q is of incorrect format: %v\n", str, err)
            fmt.Printf("using default time limit: %v\n", defaultTL)
            return
        }

        TL = time.Duration(sec * float64(time.Second))
        return
    }

    fmt.Printf("using default time limit: %v\n", defaultTL)
    return
}

