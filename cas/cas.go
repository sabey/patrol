package cas

import (
	"math/rand"
	"time"
)

const (
	cas_init_max_range = 100000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
