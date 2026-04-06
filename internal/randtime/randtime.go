package randtime

import (
	"math/rand"
	"time"
)

type Sleeper interface {
	Sleep(d time.Duration)
}

type RealSleeper struct{}

func (RealSleeper) Sleep(d time.Duration) { time.Sleep(d) }

type Rand interface {
	Intn(n int) int
	Float64() float64
}

func NewSeededRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
