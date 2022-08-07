package utils

import (
	"math/rand"
	"time"
)

// GetRandomTID generates a random number in the
// range of UDP ports
func GetRandomTID() int {
	rand.Seed(time.Now().UnixNano())
	// The range of available UDP ports is 4096-65535
	return rand.Intn(65535-4096) + 4096
}
