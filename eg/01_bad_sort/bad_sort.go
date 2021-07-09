package main

import (
	"fmt"
	"sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    a := make([]string, n)
    for i := range a {
        fmt.Scan(&a[i])
    }

    sort.Strings(a)

    for _, v := range a {
        fmt.Printf("%v ", v)
    }
    fmt.Println()
}
