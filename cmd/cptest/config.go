package main

import (
	"time"

	"github.com/kuredoro/cptest"
)

const defaultTL = 6 * time.Second

func getTL(inputs cptest.Inputs) (TL time.Duration) {
	TL = defaultTL

	if inputs.Config.Tl != (cptest.Duration{0 * time.Second}) {
		TL = inputs.Config.Tl.Duration
		return
	}

	return
}
